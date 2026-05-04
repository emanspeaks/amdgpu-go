//go:build linux

package amdgpu

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func sysfsRead(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func sysfsUint64(path string) (uint64, bool) {
	s := sysfsRead(path)
	if s == "" {
		return 0, false
	}
	base := 10
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		s = s[2:]
		base = 16
	}
	v, err := strconv.ParseUint(s, base, 64)
	return v, err == nil
}

func sysfsFloat(path string) (float64, bool) {
	s := sysfsRead(path)
	if s == "" {
		return 0, false
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func renderDevPath(card int) string {
	return fmt.Sprintf("/sys/class/drm/renderD%d/device", 128+card)
}

// PCIDevForCard returns the PCI device address string for a GPU render node
// (e.g. "0000:c2:00.0"). Returns empty string on failure.
func PCIDevForCard(card int) string {
	dev, err := filepath.EvalSymlinks(fmt.Sprintf("/sys/class/drm/renderD%d/device", 128+card))
	if err != nil {
		return ""
	}
	return filepath.Base(dev)
}

// ReadGPUName returns the product name from sysfs, falling back to a label.
func ReadGPUName(card int) string {
	p := fmt.Sprintf("/sys/class/drm/renderD%d/device/product_name", 128+card)
	if data, err := os.ReadFile(p); err == nil {
		if name := strings.TrimSpace(string(data)); name != "" {
			return name
		}
	}
	return fmt.Sprintf("AMD GPU %d", card)
}

func hwmonDirForCard(card int) string {
	cardDev := fmt.Sprintf("/sys/class/drm/renderD%d/device", 128+card)
	realCard, err := filepath.EvalSymlinks(cardDev)
	if err != nil {
		return ""
	}
	dirs, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
	for _, dir := range dirs {
		realHwmon, err := filepath.EvalSymlinks(filepath.Join(dir, "device"))
		if err != nil {
			continue
		}
		if realCard == realHwmon {
			return dir
		}
	}
	return ""
}

func findHwmonSensor(dir, kind, label string) (float64, bool) {
	inputs, _ := filepath.Glob(filepath.Join(dir, kind+"*_input"))
	for _, input := range inputs {
		prefix := strings.TrimSuffix(filepath.Base(input), "_input")
		lbl := sysfsRead(filepath.Join(dir, prefix+"_label"))
		if strings.EqualFold(lbl, label) {
			raw := sysfsRead(input)
			if val, err := strconv.ParseFloat(raw, 64); err == nil {
				return val, true
			}
		}
	}
	return 0, false
}

// ReadCPUTctl returns the k10temp "Tctl" temperature in °C and whether it was found.
func ReadCPUTctl() (float64, bool) {
	dirs, _ := filepath.Glob("/sys/class/hwmon/hwmon*")
	for _, dir := range dirs {
		if sysfsRead(filepath.Join(dir, "name")) != "k10temp" {
			continue
		}
		if v, ok := findHwmonSensor(dir, "temp", "Tctl"); ok {
			return v / 1000, true // milli-°C → °C
		}
	}
	return 0, false
}

// readCurrentClockMHz parses a pp_dpm_* sysfs file and returns the active
// clock frequency in MHz (the line marked with " *"). Returns 0 if not found.
func readCurrentClockMHz(path string) float64 {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(line, "*") {
			continue
		}
		for _, field := range strings.Fields(line) {
			lower := strings.ToLower(field)
			if strings.HasSuffix(lower, "mhz") {
				if val, err := strconv.ParseFloat(strings.TrimSuffix(lower, "mhz"), 64); err == nil {
					return val
				}
			}
		}
	}
	return 0
}

func readVRAMType(card int) string {
	return sysfsRead(filepath.Join(renderDevPath(card), "mem_info_vram_type"))
}

func readVRAMVendor(card int) string {
	v := sysfsRead(filepath.Join(renderDevPath(card), "mem_info_vram_vendor"))
	if v == "" || v == "0" {
		return ""
	}
	return v
}

func readPCIDeviceID(card int) (devID uint16, revID uint8) {
	base := renderDevPath(card)
	if v, ok := sysfsUint64(filepath.Join(base, "device")); ok {
		devID = uint16(v)
	}
	if v, ok := sysfsUint64(filepath.Join(base, "revision")); ok {
		revID = uint8(v)
	}
	return
}

func readGPUActivityPct(card int) (gpuPct, memPct, vcnPct float64, hasGPU, hasMem, hasVCN bool) {
	base := renderDevPath(card)
	gpuPct, hasGPU = sysfsFloat(filepath.Join(base, "gpu_busy_percent"))
	memPct, hasMem = sysfsFloat(filepath.Join(base, "mem_busy_percent"))
	vcnPct, hasVCN = sysfsFloat(filepath.Join(base, "vcn_busy_percent"))
	return
}

// ReadCPUCoreFreq returns per-thread CPU frequency data from sysfs cpufreq.
func ReadCPUCoreFreq() []CPUCoreFreq {
	cpuDirs, _ := filepath.Glob("/sys/devices/system/cpu/cpu[0-9]*/cpufreq")
	var result []CPUCoreFreq
	for _, dir := range cpuDirs {
		cpuDir := filepath.Dir(dir)
		threadID, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(cpuDir), "cpu"))
		if err != nil {
			continue
		}
		curKHz, ok1 := sysfsUint64(filepath.Join(dir, "scaling_cur_freq"))
		maxKHz, ok2 := sysfsUint64(filepath.Join(dir, "scaling_max_freq"))
		if !ok1 || !ok2 {
			continue
		}
		minKHz, _ := sysfsUint64(filepath.Join(dir, "scaling_min_freq"))
		coreID, _ := sysfsUint64(filepath.Join(cpuDir, "topology", "core_id"))
		result = append(result, CPUCoreFreq{
			ThreadID: threadID,
			CoreID:   int(coreID),
			CurFreq:  int(curKHz / 1000),
			MinFreq:  int(minKHz / 1000),
			MaxFreq:  int(maxKHz / 1000),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ThreadID < result[j].ThreadID })
	return result
}

// readSensors builds the Sensors map for a GPU card from hwmon and sysfs.
// Returns the map and the raw FCLK value from pp_dpm_fclk (0 if not available
// or no active DPM state marker was found).
func readSensors(card int, cpuTctl float64, hasCPUTctl bool) (map[string]interface{}, float64) {
	s := make(map[string]interface{})
	var fclkMHz float64

	if dir := hwmonDirForCard(card); dir != "" {
		if v, ok := findHwmonSensor(dir, "temp", "edge"); ok {
			s["Edge Temperature"] = SensorValue{"°C", v / 1000}
		}
		if v, ok := findHwmonSensor(dir, "power", "PPT"); ok {
			s["Average Power"] = SensorValue{"W", v / 1_000_000}
		}
		if v, ok := findHwmonSensor(dir, "in", "vddgfx"); ok {
			s["VDDGFX"] = SensorValue{"mV", v}
		}
		if v, ok := findHwmonSensor(dir, "in", "vddnb"); ok {
			s["VDDNB"] = SensorValue{"mV", v}
		}
	}
	if hasCPUTctl {
		s["CPU Tctl"] = SensorValue{"°C", cpuTctl}
	}

	base := renderDevPath(card)
	if clk := readCurrentClockMHz(filepath.Join(base, "pp_dpm_sclk")); clk > 0 {
		s["GFX_SCLK"] = SensorValue{"MHz", clk}
	} else if dir := hwmonDirForCard(card); dir != "" {
		if v, ok := findHwmonSensor(dir, "freq", "sclk"); ok && v > 0 {
			s["GFX_SCLK"] = SensorValue{"MHz", v / 1e6}
		}
	}
	if clk := readCurrentClockMHz(filepath.Join(base, "pp_dpm_mclk")); clk > 0 {
		s["GFX_MCLK"] = SensorValue{"MHz", clk}
	} else if dir := hwmonDirForCard(card); dir != "" {
		if v, ok := findHwmonSensor(dir, "freq", "mclk"); ok && v > 0 {
			s["GFX_MCLK"] = SensorValue{"MHz", v / 1e6}
		}
	}
	fclkMHz = readCurrentClockMHz(filepath.Join(base, "pp_dpm_fclk"))
	if fclkMHz > 0 {
		s["FCLK"] = SensorValue{"MHz", fclkMHz}
	}

	if freqs := ReadCPUCoreFreq(); len(freqs) > 0 {
		s["CPU Core freq"] = freqs
	}

	return s, fclkMHz
}
