//go:build linux

package amdgpu

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// xdnaRaw holds a single poll-cycle snapshot for one XDNA process.
type xdnaRaw struct {
	Name      string
	TotalKiB  uint64
	SharedKiB uint64
	ActiveKiB uint64
	EngineNS  uint64
}

// XdnaState persists XDNA fdinfo across poll cycles for delta computation.
type XdnaState struct {
	Prev     map[int]xdnaRaw
	PrevTime time.Time
	AccelDev string // e.g. "/dev/accel/accel0"
}

// findXDNAAccelDev returns the first /dev/accel/accelN path that hosts an
// amdxdna device (identified by the presence of a "vbnv" sysfs file).
// Returns "" if no XDNA device is found.
func findXDNAAccelDev() string {
	const accelMajor = 261
	for i := range 64 {
		accel := fmt.Sprintf("/dev/accel/accel%d", i)
		if _, err := os.Stat(accel); os.IsNotExist(err) {
			continue
		}
		// Confirm it's an XDNA device via sysfs.
		sysfs := fmt.Sprintf("/sys/dev/char/%d:%d/device", accelMajor, i)
		if _, err := os.Stat(filepath.Join(sysfs, "vbnv")); err == nil {
			return accel
		}
	}
	return ""
}

// InitXdnaState probes for the XDNA accel device and returns a ready state.
func InitXdnaState() XdnaState {
	return XdnaState{
		Prev:     make(map[int]xdnaRaw),
		AccelDev: findXDNAAccelDev(),
	}
}

// ScanXdnaFdinfo scans /proc/*/fdinfo/* for amdxdna_accel_driver entries and
// returns a snapshot map keyed by PID.
func ScanXdnaFdinfo() map[int]xdnaRaw {
	result := make(map[int]xdnaRaw)

	procDirs, _ := filepath.Glob("/proc/[0-9]*/fdinfo")
	for _, fdinfoDir := range procDirs {
		pidStr := filepath.Base(filepath.Dir(fdinfoDir))
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		fds, _ := filepath.Glob(filepath.Join(fdinfoDir, "*"))
		seenIDs := make(map[uint64]bool)
		var agg xdnaRaw
		found := false

		for _, fd := range fds {
			entry, clientID := parseXdnaFdinfoEntry(fd)
			if entry == nil {
				continue
			}
			if seenIDs[clientID] {
				continue // deduplicate by drm-client-id
			}
			seenIDs[clientID] = true

			if !found {
				found = true
				agg.Name = readProcComm(pid)
			}
			agg.TotalKiB += entry.TotalKiB
			agg.SharedKiB += entry.SharedKiB
			agg.ActiveKiB += entry.ActiveKiB
			agg.EngineNS += entry.EngineNS
		}

		if found {
			result[pid] = agg
		}
	}
	return result
}

// parseXdnaFdinfoEntry reads one fdinfo file and returns an xdnaRaw snapshot
// plus the drm-client-id (for dedup). Returns nil if the fd is not XDNA.
func parseXdnaFdinfoEntry(path string) (*xdnaRaw, uint64) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0
	}
	defer f.Close()

	var snap xdnaRaw
	var clientID uint64
	isXDNA := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		colon := strings.IndexByte(line, ':')
		if colon < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colon])
		val := strings.TrimSpace(line[colon+1:])

		switch key {
		case "drm-driver":
			if val != "amdxdna_accel_driver" {
				return nil, 0
			}
			isXDNA = true
		case "drm-client-id":
			clientID, _ = strconv.ParseUint(val, 10, 64)
		case "drm-total-memory":
			snap.TotalKiB = parseKiB(val)
		case "drm-shared-memory":
			snap.SharedKiB = parseKiB(val)
		case "drm-active-memory":
			snap.ActiveKiB = parseKiB(val)
		case "drm-engine-npu-amdxdna":
			// value is "<N> ns"
			if ns, ok := parseNS(val); ok {
				snap.EngineNS = ns
			}
		}
	}

	if !isXDNA {
		return nil, 0
	}
	return &snap, clientID
}

// parseKiB parses a memory value string like "4096 KiB" or "8 MiB" into KiB.
func parseKiB(s string) uint64 {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return 0
	}
	v, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return 0
	}
	if len(fields) >= 2 {
		switch strings.ToUpper(fields[1]) {
		case "MIB", "MB":
			v *= 1024
		case "GIB", "GB":
			v *= 1024 * 1024
		}
	}
	return v
}

// ComputeXdnaDeltas computes per-process NPU usage from XDNA fdinfo snapshots.
// Returns the xdna_fdinfo map keyed by PID string.
func ComputeXdnaDeltas(prev, curr map[int]xdnaRaw, dt float64) map[string]interface{} {
	out := make(map[string]interface{})

	for pid, cs := range curr {
		var npuPct float64
		if ps, ok := prev[pid]; ok && dt > 0 && cs.EngineNS >= ps.EngineNS {
			npuPct = float64(cs.EngineNS-ps.EngineNS) / (dt * 1e9) * 100
			if npuPct > 100 {
				npuPct = 100
			}
		}

		totalMiB := float64(cs.TotalKiB) / 1024
		sharedMiB := float64(cs.SharedKiB) / 1024
		activeMiB := float64(cs.ActiveKiB) / 1024

		usage := map[string]interface{}{
			"Total Memory Usage":  map[string]interface{}{"unit": "MiB", "value": totalMiB},
			"Shared Memory Usage": map[string]interface{}{"unit": "MiB", "value": sharedMiB},
			"Active Memory Usage": map[string]interface{}{"unit": "MiB", "value": activeMiB},
			"NPU":                 map[string]interface{}{"unit": "%", "value": npuPct},
		}
		out[strconv.Itoa(pid)] = map[string]interface{}{
			"name":  cs.Name,
			"usage": usage,
		}
	}

	return out
}
