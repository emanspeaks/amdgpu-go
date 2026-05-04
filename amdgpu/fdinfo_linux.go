//go:build linux

package amdgpu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// fdinfoProcRaw holds a single poll-cycle snapshot for one process.
type fdinfoProcRaw struct {
	Name       string
	EngineNS   map[string]uint64 // accumulated engine nanoseconds per engine name
	MemBytes   map[string]uint64 // memory bytes per type (vram, gtt, cpu)
	CPUJiffies uint64            // utime+stime from /proc/PID/stat
}

// fdinfoEngineVal is the {unit, value} wrapper for fdinfo fields.
// Value is a pointer so it marshals as JSON null when nil (engine not present).
type fdinfoEngineVal struct {
	Unit  string   `json:"unit"`
	Value *float64 `json:"value"`
}

// fdinfoStdKeys lists the engine/memory keys always emitted per process,
// matching amdgpu_top's fixed fdinfo schema. Keys not seen in any fdinfo entry
// for this device will be emitted with a null value.
var fdinfoStdKeys = []struct{ key, unit string }{
	{"CPU", "%"},
	{"GFX", "%"},
	{"Compute", "%"},
	{"DMA", "%"},
	{"Decode", "%"},
	{"Encode", "%"},
	{"Media", "%"},
	{"VCN_JPEG", "%"},
	{"VCN_Unified", "%"},
	{"VPE", "%"},
	{"VRAM", "MiB"},
	{"GTT", "MiB"},
}

// ScanFdinfo scans /proc/*/fdinfo/* for amdgpu file descriptors on pciDev
// and returns a snapshot map keyed by PID.
func ScanFdinfo(pciDev string) map[int]fdinfoProcRaw {
	result := make(map[int]fdinfoProcRaw)

	procDirs, _ := filepath.Glob("/proc/[0-9]*/fdinfo")
	for _, fdinfoDir := range procDirs {
		pidStr := filepath.Base(filepath.Dir(fdinfoDir))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		fds, _ := filepath.Glob(filepath.Join(fdinfoDir, "*"))
		var found bool
		for _, fd := range fds {
			entry := parseFdinfoEntry(fd, pciDev)
			if entry == nil {
				continue
			}
			if !found {
				found = true
				result[pid] = fdinfoProcRaw{
					Name:     readProcComm(pid),
					EngineNS: make(map[string]uint64),
					MemBytes: make(map[string]uint64),
				}
			}
			existing := result[pid]
			for k, v := range entry.EngineNS {
				existing.EngineNS[k] += v
			}
			for k, v := range entry.MemBytes {
				if v > existing.MemBytes[k] {
					existing.MemBytes[k] = v // max across fds (same process)
				}
			}
			result[pid] = existing
		}

		if found {
			existing := result[pid]
			existing.CPUJiffies = readProcCPUJiffies(pid)
			result[pid] = existing
		}
	}
	return result
}

func parseFdinfoEntry(path, wantPCIDev string) *fdinfoProcRaw {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	snap := &fdinfoProcRaw{
		EngineNS: make(map[string]uint64),
		MemBytes: make(map[string]uint64),
	}
	isAMDGPU := false
	pciMatch := wantPCIDev == ""

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])

		switch {
		case key == "drm-driver":
			if val != "amdgpu" {
				return nil
			}
			isAMDGPU = true
		case key == "drm-pdev":
			if wantPCIDev != "" && val == wantPCIDev {
				pciMatch = true
			}
		case strings.HasPrefix(key, "drm-engine-"):
			eng := engineKey(strings.TrimPrefix(key, "drm-engine-"))
			if ns, ok := parseNS(val); ok {
				snap.EngineNS[eng] = ns
			}
		case strings.HasPrefix(key, "drm-memory-"):
			mem := strings.TrimPrefix(key, "drm-memory-")
			if b, ok := parseBytes(val); ok {
				snap.MemBytes[mem] = b
			}
		}
	}

	if !isAMDGPU || !pciMatch {
		return nil
	}
	return snap
}

// engineKey normalizes a drm-engine-* suffix to the amdgpu_top key name.
func engineKey(raw string) string {
	switch raw {
	case "gfx":
		return "GFX"
	case "compute":
		return "Compute"
	case "dma", "sdma":
		return "DMA"
	case "dec":
		return "Decode"
	case "enc":
		return "Encode"
	case "enc_1":
		return "Media"
	case "jpeg":
		return "VCN_JPEG"
	case "vcn":
		return "VCN_Unified"
	case "vpe":
		return "VPE"
	default:
		return raw
	}
}

func parseNS(s string) (uint64, bool) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.ParseUint(fields[0], 10, 64)
	return v, err == nil
}

func parseBytes(s string) (uint64, bool) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0, false
	}
	if len(fields) >= 2 {
		switch strings.ToUpper(fields[1]) {
		case "KIB", "KB":
			v *= 1024
		case "MIB", "MB":
			v *= 1024 * 1024
		case "GIB", "GB":
			v *= 1024 * 1024 * 1024
		}
	}
	return v, true
}

func readProcComm(pid int) string {
	comm := sysfsRead(fmt.Sprintf("/proc/%d/comm", pid))
	if comm != "" {
		return comm
	}
	return fmt.Sprintf("pid%d", pid)
}

func readProcCPUJiffies(pid int) uint64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0
	}
	s := string(data)
	// Skip past the closing ) of the comm field (which may contain spaces).
	rp := strings.LastIndex(s, ")")
	if rp < 0 {
		return 0
	}
	// Fields after ")": state(0) ppid(1) pgrp(2) session(3) tty_nr(4) tpgid(5)
	// flags(6) minflt(7) cminflt(8) majflt(9) cmajflt(10) utime(11) stime(12)
	fields := strings.Fields(strings.TrimSpace(s[rp+1:]))
	if len(fields) < 13 {
		return 0
	}
	utime, _ := strconv.ParseUint(fields[11], 10, 64)
	stime, _ := strconv.ParseUint(fields[12], 10, 64)
	return utime + stime
}

// collectSeenEngines returns the union of engine keys present across both the
// current and previous fdinfo snapshots. A key in either cycle is "available".
func collectSeenEngines(curr, prev map[int]fdinfoProcRaw) map[string]bool {
	seen := make(map[string]bool)
	for _, p := range curr {
		for k := range p.EngineNS {
			seen[k] = true
		}
	}
	for _, p := range prev {
		for k := range p.EngineNS {
			seen[k] = true
		}
	}
	return seen
}

// ComputeFdinfoDeltas computes per-process GPU usage by comparing current and
// previous fdinfo snapshots. dt is elapsed seconds between snapshots.
//
// Returns the fdinfo map (keyed by PID string, double-nested amdgpu_top format)
// and the "Total fdinfo" aggregate entry (same format, key "Total").
func ComputeFdinfoDeltas(
	prev, curr map[int]fdinfoProcRaw,
	dt float64,
	seenEngines map[string]bool,
) (map[string]interface{}, interface{}) {
	const clkTck = 100.0 // standard Linux jiffies/sec

	out := make(map[string]interface{})
	totals := make(map[string]float64)

	for pid, cs := range curr {
		var ps *fdinfoProcRaw
		if p, ok := prev[pid]; ok {
			ps = &p
		}

		var cpuJifDelta uint64
		if ps != nil && cs.CPUJiffies >= ps.CPUJiffies {
			cpuJifDelta = cs.CPUJiffies - ps.CPUJiffies
		}

		inner := make(map[string]interface{}, len(fdinfoStdKeys))
		for _, e := range fdinfoStdKeys {
			switch e.key {
			case "CPU":
				var pct float64
				if dt > 0 {
					pct = float64(cpuJifDelta) / (dt * clkTck) * 100
					if pct > 100 {
						pct = 100
					}
				}
				v := pct
				inner["CPU"] = fdinfoEngineVal{"%", &v}
				totals["CPU"] += pct

			case "VRAM":
				v := float64(cs.MemBytes["vram"]) / (1 << 20)
				inner["VRAM"] = fdinfoEngineVal{"MiB", &v}
				totals["VRAM"] += v

			case "GTT":
				v := float64(cs.MemBytes["gtt"]) / (1 << 20)
				inner["GTT"] = fdinfoEngineVal{"MiB", &v}
				totals["GTT"] += v

			default:
				if !seenEngines[e.key] {
					inner[e.key] = fdinfoEngineVal{e.unit, nil}
				} else {
					var pct float64
					if ps != nil && dt > 0 {
						curNS := cs.EngineNS[e.key]
						prevNS := ps.EngineNS[e.key]
						if curNS >= prevNS {
							pct = float64(curNS-prevNS) / (dt * 1e9) * 100
							if pct > 100 {
								pct = 100
							}
						}
					}
					v := pct
					inner[e.key] = fdinfoEngineVal{e.unit, &v}
					totals[e.key] += pct
				}
			}
		}

		// Double-nested: {name, usage: {name, usage: {...}}}
		innerUsage := map[string]interface{}{
			"name":  cs.Name,
			"usage": inner,
		}
		out[strconv.Itoa(pid)] = map[string]interface{}{
			"name":  cs.Name,
			"usage": innerUsage,
		}
	}

	// Total fdinfo aggregate — flat map matching amdgpu_top format (not double-nested).
	totalInner := make(map[string]interface{}, len(fdinfoStdKeys))
	for _, e := range fdinfoStdKeys {
		if e.key != "CPU" && e.key != "VRAM" && e.key != "GTT" && !seenEngines[e.key] {
			totalInner[e.key] = fdinfoEngineVal{e.unit, nil}
		} else {
			v := totals[e.key]
			totalInner[e.key] = fdinfoEngineVal{e.unit, &v}
		}
	}
	return out, totalInner
}
