package amdgpu

import "encoding/json"

// BuildFrame assembles a Frame from device snapshots with the given sampling
// period and backend version string. The returned Frame has the same JSON
// shape as amdgpu_top -J output.
func BuildFrame(snaps []*DeviceSnapshot, periodMs int, backendVersion string) *Frame {
	devs := make([]DeviceSnapshot, len(snaps))
	for i, s := range snaps {
		devs[i] = *s
	}
	return &Frame{
		Devices:             devs,
		DevicesLen:          len(devs),
		Period:              PeriodInfo{Duration: periodMs, DurationUnit: "ms"},
		BackendVersion:      backendVersion,
		Title:               "amdgpu_top",
		SuspendedDevices:    []interface{}{},
		SuspendedDevicesLen: 0,
	}
}

// MarshalFrameJSON builds and marshals a complete amdgpu_top-compatible JSON
// frame. It is a convenience wrapper around BuildFrame + json.Marshal.
func MarshalFrameJSON(snaps []*DeviceSnapshot, periodMs int, backendVersion string) ([]byte, error) {
	return json.Marshal(BuildFrame(snaps, periodMs, backendVersion))
}
