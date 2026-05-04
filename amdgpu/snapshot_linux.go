//go:build linux && cgo

package amdgpu

import (
	"fmt"
	"time"
)

// FdinfoState tracks per-device fdinfo snapshots needed for delta computation.
type FdinfoState struct {
	Prev      map[int]fdinfoProcRaw
	PrevTime  time.Time
	HWEngines map[string]bool // engines known to exist on hardware; seeded at init
}

// PollState holds per-device state that persists across poll cycles.
type PollState struct {
	Gen         Generation
	GRBM2Bits   []GRBMBit
	PCIDev      string
	Fdinfo      FdinfoState
	DevInfo     *DeviceInfo // cached; static per device
	DRMVer      string
	LastFCLKMHz float64 // last stable FCLK from pp_dpm_fclk

	// CPU Tctl is cached per-state to avoid hammering hwmon every sample.
	cpuTctl    float64
	hasCPUTctl bool
	cpuTctlAt  time.Time
}

// InitPollState initializes a PollState for the given device and card index.
// It reads static device info once so subsequent snapshots don't need to.
func InitPollState(dev *Device, card int) *PollState {
	state := &PollState{
		PCIDev: PCIDevForCard(card),
		Fdinfo: FdinfoState{Prev: make(map[int]fdinfoProcRaw)},
	}
	if info, err := dev.DeviceInfo(); err == nil {
		state.DevInfo = info
		state.Gen = DetectGeneration(info.Family)
		state.GRBM2Bits = GRBM2BitsForGen(state.Gen)
	} else {
		state.Gen = GenUnknown
		state.GRBM2Bits = GRBM2BitsForGen(GenUnknown)
	}
	state.Fdinfo.HWEngines = HWEnginesForGen(state.Gen)
	if dv, err := dev.DRMVersion(); err == nil {
		state.DRMVer = dv.Version
	}
	return state
}

// SampleGRBMMulti samples GRBM and GRBM2 registers across all devices over the
// given duration with the given number of sample points. Sampling interleaves
// reads across all devices within each step so results are temporally aligned.
// Returns one GRBMSample per device, or an error if any GRBM read fails.
func SampleGRBMMulti(devices []*Device, states []*PollState, duration time.Duration, samples int) ([]GRBMSample, error) {
	result := make([]GRBMSample, len(devices))
	for i := range result {
		result[i].Samples = samples
	}
	if samples <= 0 || duration <= 0 {
		return result, nil
	}
	sleepPer := duration / time.Duration(samples)
	for samp := 0; samp < samples; samp++ {
		for i, dev := range devices {
			g, err := dev.ReadGRBM()
			if err != nil {
				return nil, fmt.Errorf("GRBM read device %d: %w", i, err)
			}
			for b := uint(0); b < 32; b++ {
				if g&(1<<b) != 0 {
					result[i].GRBMCounts[b]++
				}
			}
			g2, _ := dev.ReadGRBM2() // GRBM2 failure is non-fatal
			for b := uint(0); b < 32; b++ {
				if g2&(1<<b) != 0 {
					result[i].GRBM2Counts[b]++
				}
			}
		}
		time.Sleep(sleepPer)
	}
	return result, nil
}

// ReadDeviceSnapshot reads a complete snapshot for one GPU device.
// gs holds pre-sampled GRBM bit counts; pass nil when noPC is true.
func ReadDeviceSnapshot(dev *Device, card int, state *PollState, gs *GRBMSample, noPC bool) (*DeviceSnapshot, error) {
	mem, err := dev.MemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("MemoryInfo: %w", err)
	}
	const mib = 1 << 20

	// Refresh CPU Tctl every ~5 seconds to avoid hammering hwmon.
	if time.Since(state.cpuTctlAt) > 5*time.Second {
		state.cpuTctl, state.hasCPUTctl = ReadCPUTctl()
		state.cpuTctlAt = time.Now()
	}

	// Sensors. FCLK persistence: update stored value when pp_dpm_fclk has an
	// active marker; use the stored value when the marker is absent (DPM transition).
	sensors, fclkFromDPM := readSensors(card, state.cpuTctl, state.hasCPUTctl)
	if fclkFromDPM > 0 {
		state.LastFCLKMHz = fclkFromDPM
	} else if state.LastFCLKMHz > 0 {
		sensors["FCLK"] = SensorValue{"MHz", state.LastFCLKMHz}
	} else {
		delete(sensors, "FCLK")
	}

	gm := ParseGPUMetrics(card)

	// Activity percentages.
	gpuSysPct, memSysPct, vcnSysPct, hasGPUSys, hasMemSys, hasVCNSys := readGPUActivityPct(card)

	var gfxPct, memPct, mediaPct float64
	if noPC || gs == nil || gs.Samples == 0 {
		if v, ok := gm["average_gfx_activity"].(uint16); ok && v > 0 {
			gfxPct = float64(v)
		} else if hasGPUSys {
			gfxPct = gpuSysPct
		}
	} else {
		gfxPct = float64(gs.GRBMCounts[31]) / float64(gs.Samples) * 100
		_ = gpuSysPct
		_ = hasGPUSys
	}
	if hasMemSys {
		memPct = memSysPct
	} else if v, ok := gm["average_umc_activity"].(uint16); ok {
		memPct = float64(v)
	}
	if v, ok := gm["average_vcn_activity"].(uint16); ok {
		mediaPct = float64(v)
	} else if v, ok := gm["average_mm_activity"].(uint16); ok {
		mediaPct = float64(v)
	} else if hasVCNSys {
		mediaPct = vcnSysPct
	}

	// GRBM/GRBM2 percentages from sampled bit counts.
	grbm := make(map[string]SensorValue)
	grbm2 := make(map[string]SensorValue)
	if !noPC && gs != nil && gs.Samples > 0 {
		for _, e := range GRBMBits() {
			grbm[e.Name] = SensorValue{"%", float64(gs.GRBMCounts[e.Bit]) / float64(gs.Samples) * 100}
		}
		for _, e := range state.GRBM2Bits {
			grbm2[e.Name] = SensorValue{"%", float64(gs.GRBM2Counts[e.Bit]) / float64(gs.Samples) * 100}
		}
	}

	// Device Info map.
	var infoMap map[string]interface{}
	if di := state.DevInfo; di != nil {
		infoMap = buildDeviceInfoMap(card, di, mem, state.PCIDev, state.DRMVer)
	} else {
		name := ReadGPUName(card)
		infoMap = map[string]interface{}{"DeviceName": name, "ASIC Name": name}
	}

	// fdinfo delta computation.
	now := time.Now()
	currFdinfo := ScanFdinfo(state.PCIDev)
	var dt float64
	if !state.Fdinfo.PrevTime.IsZero() {
		dt = now.Sub(state.Fdinfo.PrevTime).Seconds()
	}
	seenEngines := collectSeenEngines(currFdinfo, state.Fdinfo.Prev)
	for k := range state.Fdinfo.HWEngines {
		seenEngines[k] = true
	}
	fdinfo, totalFdinfo := ComputeFdinfoDeltas(state.Fdinfo.Prev, currFdinfo, dt, seenEngines)
	state.Fdinfo.Prev = currFdinfo
	state.Fdinfo.PrevTime = now

	return &DeviceSnapshot{
		Info: infoMap,
		Activity: map[string]SensorValue{
			"GFX":         {"%", gfxPct},
			"Memory":      {"%", memPct},
			"MediaEngine": {"%", mediaPct},
		},
		VRAM: map[string]SensorValue{
			"Total VRAM Usage": {"MiB", float64(mem.VRAMHeapUsage) / mib},
			"Total VRAM":       {"MiB", float64(mem.VRAMTotalHeapSize) / mib},
			"Total GTT Usage":  {"MiB", float64(mem.GTTHeapUsage) / mib},
			"Total GTT":        {"MiB", float64(mem.GTTTotalHeapSize) / mib},
		},
		Sensors:     sensors,
		GPUMetrics:  gm,
		GRBM:        grbm,
		GRBM2:       grbm2,
		Fdinfo:      fdinfo,
		TotalFdinfo: totalFdinfo,
		NPUMetrics:  buildNPUMetrics(gm),
	}, nil
}
