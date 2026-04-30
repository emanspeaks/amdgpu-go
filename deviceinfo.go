package amdgpu

// DeviceInfo contains GPU identification and capability data from AMDGPU_INFO_DEV_INFO.
type DeviceInfo struct {
	Family         uint32 // AMDGPU_FAMILY_* constant (e.g. 143 = FAMILY_NV)
	ExternalRev    uint32 // external chip revision; identifies specific ASIC variant
	IsApu          bool   // true when AMDGPU_IDS_FLAGS_FUSION is set (integrated GPU)
	MaxEngineClock uint64 // kHz
	MaxMemoryClock uint64 // kHz
	NUMTCCBlocks   uint32 // number of texture channel caches (L2 slices)
	// GFX11+ only; zero on older hardware
	GL0CacheSize uint32 // L0 VMEM cache per CU (tcp_cache_size), bytes
	GL1CacheSize uint32 // L1 cache per shader array (gl1c_cache_size), bytes
	GL2CacheSize uint32 // L2 cache (gl2c_cache_size), bytes
	MallSize     uint64 // L3 infinity cache (mall_size), bytes
}

// MemoryInfo contains VRAM/GTT heap information from AMDGPU_INFO_MEMORY.
type MemoryInfo struct {
	VRAMTotalHeapSize       uint64
	VRAMUsableHeapSize      uint64
	VRAMHeapUsage           uint64
	CPUAccessibleTotalSize  uint64 // CPU-visible VRAM; equals VRAM total when rebar is enabled
	CPUAccessibleUsableSize uint64
	CPUAccessibleHeapUsage  uint64
	GTTTotalHeapSize        uint64
	GTTUsableHeapSize       uint64
	GTTHeapUsage            uint64
	ResizableBar            bool // true when CPU can access all VRAM (rebar enabled)
}

// VramGttInfo contains VRAM/GTT size info from AMDGPU_INFO_VRAM_GTT.
type VramGttInfo struct {
	VRAMSize          uint64
	VRAMCpuAccessible uint64 // CPU-visible VRAM; < VRAMSize when rebar is disabled
	GTTSize           uint64
}

// DRMVersion contains DRM driver version information.
type DRMVersion struct {
	Name        string // e.g., "amdgpu"
	Version     string // e.g., "5.18.0"
	Date        string // e.g., "20220622"
	Description string // e.g., "AMD GPU"
}
