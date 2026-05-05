package amdgpu

// SensorValue is the {unit, value} object used in amdgpu_top JSON output.
type SensorValue struct {
	Unit  string  `json:"unit"`
	Value float64 `json:"value"`
}

// CPUCoreFreq holds frequency data for one CPU thread.
type CPUCoreFreq struct {
	ThreadID int `json:"thread_id"`
	CoreID   int `json:"core_id"`
	CurFreq  int `json:"cur_freq"` // MHz
	MinFreq  int `json:"min_freq"` // MHz
	MaxFreq  int `json:"max_freq"` // MHz
}

// GRBMSample holds raw bit-set counts from a GRBM/GRBM2 sampling window.
type GRBMSample struct {
	GRBMCounts  [32]int
	GRBM2Counts [32]int
	Samples     int
}

// DeviceSnapshot is a complete point-in-time reading for one GPU device.
// Field names mirror the amdgpu_top JSON device object keys exactly.
type DeviceSnapshot struct {
	Info        map[string]interface{} `json:"Info"`
	Activity    map[string]SensorValue `json:"gpu_activity"`
	VRAM        map[string]SensorValue `json:"VRAM"`
	Sensors     map[string]interface{} `json:"Sensors"` // SensorValue scalars + []CPUCoreFreq
	GPUMetrics  map[string]interface{} `json:"gpu_metrics"`
	GRBM        map[string]SensorValue `json:"GRBM"`
	GRBM2       map[string]SensorValue `json:"GRBM2"`
	Fdinfo      map[string]interface{} `json:"fdinfo"`
	TotalFdinfo interface{}            `json:"Total fdinfo"`
	NPUMetrics  map[string]interface{} `json:"npu_metrics"`
	XdnaFdinfo  map[string]interface{} `json:"xdna_fdinfo,omitempty"`
}

// Frame is the top-level amdgpu_top-compatible JSON structure.
type Frame struct {
	Devices             []DeviceSnapshot `json:"devices"`
	DevicesLen          int              `json:"devices_len"`
	Period              PeriodInfo       `json:"period"`
	BackendVersion      string           `json:"amdgpu_top_version"`
	Title               string           `json:"title,omitempty"`
	SuspendedDevices    []interface{}    `json:"suspended_devices"`
	SuspendedDevicesLen int              `json:"suspended_devices_len"`
}

// PeriodInfo describes the sampling period.
type PeriodInfo struct {
	Duration     int    `json:"duration"`
	DurationUnit string `json:"unit"`
}
