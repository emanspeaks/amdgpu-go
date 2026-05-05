package amdgpu

import "fmt"

// Generation identifies the GPU shader generation, used to select register layouts.
type Generation int

const (
	GenUnknown Generation = iota
	GenGFX9
	GenGFX10
	GenGFX10_3
	GenGFX11
	GenGFX12
)

// ChipMeta holds chip-specific metadata derived from the AMDGPU family ID.
type ChipMeta struct {
	ChipClass string // e.g. "GFX1151"
	Name      string // e.g. "Strix Halo"
	GFXTarget string // e.g. "gfx1151"
	GPUFamily string // e.g. "GC 11.5.0"
	NPUName   string // e.g. "XDNA2"; empty if no integrated NPU
}

// ChipMetaForFamily returns metadata for the given AMDGPU_FAMILY_* ID.
// Family IDs are from include/uapi/drm/amdgpu_drm.h.
func ChipMetaForFamily(family uint32) ChipMeta {
	switch family {
	case 141:
		return ChipMeta{"GFX9", "Vega10", "gfx900", "GC 9.0.0", ""}
	case 142:
		return ChipMeta{"GFX9", "Raven", "gfx902", "GC 9.1.0", ""}
	case 143:
		return ChipMeta{"GFX10", "Navi10", "gfx1010", "GC 10.1.0", ""}
	case 144:
		return ChipMeta{"GFX10_3", "Van Gogh", "gfx1033", "GC 10.3.3", ""}
	case 145:
		return ChipMeta{"GFX11", "Navi31", "gfx1100", "GC 11.0.0", ""}
	case 146:
		return ChipMeta{"GFX10_3", "Yellow Carp", "gfx1035", "GC 10.3.5", ""}
	case 148:
		return ChipMeta{"GFX11", "Phoenix", "gfx1103", "GC 11.0.1", "XDNA"}
	case 149:
		return ChipMeta{"GFX10_3", "Raphael/Mendocino", "gfx1036", "GC 10.3.6", ""}
	case 150:
		return ChipMeta{"GFX11_5", "Strix Halo", "gfx1151", "GC 11.5.0", "XDNA2"}
	case 151:
		return ChipMeta{"GFX10_3", "GC 10.3.7", "gfx1037", "GC 10.3.7", ""}
	case 152:
		return ChipMeta{"GFX12", "Navi48", "gfx1200", "GC 12.0.0", ""}
	case 154:
		return ChipMeta{"GFX11_5", "GC 11.5.4", "gfx1154", "GC 11.5.4", ""}
	default:
		return ChipMeta{
			fmt.Sprintf("GFX?%d", family),
			fmt.Sprintf("Unknown (family %d)", family),
			fmt.Sprintf("gfxunknown%d", family),
			"",
			"",
		}
	}
}

// HWEnginesForGen returns the set of fdinfo engine keys expected to be present
// on hardware of the given generation. Used to emit value:0 instead of null
// for engines that exist on the hardware but produced no fdinfo entries.
func HWEnginesForGen(gen Generation) map[string]bool {
	engines := map[string]bool{
		"GFX":     true,
		"Compute": true,
		"DMA":     true,
	}
	switch gen {
	case GenGFX10, GenGFX10_3:
		engines["Decode"] = true
		engines["Encode"] = true
		engines["VCN_JPEG"] = true
	case GenGFX11, GenGFX12:
		engines["Media"] = true
		engines["VCN_JPEG"] = true
		engines["VCN_Unified"] = true
		engines["VPE"] = true
	}
	return engines
}

// RBPlusForGen returns true for generations that use RB+ (8 ROPs per RB pipe)
// rather than the older RB (4 ROPs per RB pipe). GFX9 and later all use RB+.
func RBPlusForGen(gen Generation) bool {
	return gen >= GenGFX9
}

// FLOPsPerCUPerMHz returns the peak FP32 FLOP count per compute unit per MHz.
// GFX11+ (RDNA3) doubles throughput via dual-issue; earlier generations use 128.
func FLOPsPerCUPerMHz(gen Generation) int {
	switch gen {
	case GenGFX11, GenGFX12:
		return 256
	default:
		return 128
	}
}

// DetectGeneration maps an AMDGPU family ID to a shader generation.
func DetectGeneration(family uint32) Generation {
	switch family {
	case 141, 142:
		return GenGFX9
	case 143:
		return GenGFX10
	case 144, 146, 149, 151:
		return GenGFX10_3
	case 145, 148, 150, 154:
		return GenGFX11
	case 152:
		return GenGFX12
	default:
		return GenUnknown
	}
}
