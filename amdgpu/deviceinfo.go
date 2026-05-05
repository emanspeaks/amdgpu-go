package amdgpu

// DeviceInfo contains GPU identification and capability data from AMDGPU_INFO_DEV_INFO.
type DeviceInfo struct {
	Family         uint32 // AMDGPU_FAMILY_* constant (e.g. 143 = FAMILY_NV)
	ExternalRev    uint32 // external chip revision; identifies specific ASIC variant
	MarketingName  string // product marketing name from amdgpu_get_marketing_name()
	IsApu          bool   // true when AMDGPU_IDS_FLAGS_FUSION is set (integrated GPU)
	MaxEngineClock uint64 // kHz
	MaxMemoryClock uint64 // kHz
	NUMTCCBlocks   uint32 // number of texture channel caches (L2 slices)
	// Shader topology — used to compute total CU count and peak FLOPS.
	NumShaderEngines         uint32 // number of top-level shader engines
	NumShaderArraysPerEngine uint32 // shader arrays per shader engine
	NumCUPerSH               uint32 // compute units per shader array
	// GFX11+ only; zero on older hardware
	GL0CacheSize uint32 // L0 VMEM cache per CU (tcp_cache_size), bytes
	GL1CacheSize uint32 // L1 cache per shader array (gl1c_cache_size), bytes
	GL2CacheSize uint32 // L2 cache (gl2c_cache_size), bytes
	MallSize     uint64 // L3 infinity cache (mall_size), bytes
	// Memory subsystem
	VRAMType     uint32 // AMDGPU_VRAM_TYPE_* (e.g. 12=LPDDR5, 9=GDDR6, 6=HBM)
	VRAMBitWidth uint32 // memory bus width in bits (e.g. 256)
	// Render backends
	NumRBPipes uint32 // number of render backend pipes (rb_pipes)
}

// VBIOSInfo contains VBIOS identification strings from AMDGPU_INFO_VBIOS_* ioctls.
type VBIOSInfo struct {
	Name   string // e.g. "AMD STRIX_HALO_GENERIC"
	PN     string // part number, e.g. "113-STRXLGEN-001"
	VerStr string // version string, e.g. "023.011.000.039.000001"
	Date   string // build date, e.g. "2024/06/17 02:08"
}

// VideoCapEntry holds the max encode/decode dimensions for one codec.
type VideoCapEntry struct {
	MaxWidth  uint32
	MaxHeight uint32
}

// VideoCapsInfo holds per-codec decode and encode capability dimensions.
type VideoCapsInfo struct {
	Decode [8]VideoCapEntry // indexed by AMDGPU_INFO_VIDEO_CAPS_CODEC_IDX_*
	Encode [8]VideoCapEntry
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

// FirmwareVersion contains firmware version information.
type FirmwareVersion struct {
	Name    string // e.g., "VCE", "UVD", "VCN"
	Version uint32
	Feature uint32
}

// HWIPInfo contains hardware IP block information.
type HWIPInfo struct {
	Type           HW_IP_TYPE
	Instance       uint32
	Major          uint32
	Minor          uint32
	Enabled        bool
	AvailableRings uint32 // bitmask; popcount = number of queues
	RevMajor       uint32
	RevMinor       uint32
}
