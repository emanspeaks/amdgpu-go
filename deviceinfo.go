package amdgpu

// DeviceInfo contains GPU identification and capability data.
type DeviceInfo struct {
	ASICName       ASIC_NAME
	ChipClass      CHIP_CLASS
	IsApu          bool
	MaxEngineClock uint32 // in 10kHz units
	MaxMemoryClock uint32 // in 10kHz units
	L1CacheSize    uint32 // bytes
	GL1CacheSize   uint32 // bytes
	L2CacheSize    uint32 // bytes
	L3CacheSizeMB  uint32 // MiB
	NUMTCCBlocks   uint32
}

// MemoryInfo contains VRAM/GTT heap information.
type MemoryInfo struct {
	VRAMHeapUsage      uint64
	VRAMTotalHeapSize  uint64
	VRAMUsableHeapSize uint64
	GTTHeapUsage       uint64
	GTTTotalHeapSize   uint64
	GTTUsableHeapSize  uint64
	ResizableBar       bool
}

// VramGttInfo contains usable VRAM/GTT heap sizes.
type VramGttInfo struct {
	VRAMTotal  uint64
	VRAMUsable uint64
	GTTTotal   uint64
	GTTUsable  uint64
}

// DRMVersion contains DRM driver version information.
type DRMVersion struct {
	Name        string // e.g., "amdpgu"
	Version     string // e.g., "5.18.0"
	Date        string // e.g., "20220622"
	Description string // e.g., "Linux AMDGPU"
}
