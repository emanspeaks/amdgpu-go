package amdgpu

// GRBMBit maps a GRBM/GRBM2 register bit position to a display name.
type GRBMBit struct {
	Name string
	Bit  uint
}

// GRBMBitsForGen returns the GRBM_STATUS bit table for the given chip generation.
// Bit positions sourced from drivers/gpu/drm/amd/include/asic_reg/gc/gc_*_sh_mask.h
//
// The entry at bit 21 differs by generation: GFX9 uses WD_BUSY (Work Distributor),
// while GFX10+ replaced VGT+WD+IA with a unified GE block (GE_BUSY).
func GRBMBitsForGen(gen Generation) []GRBMBit {
	base := []GRBMBit{
		{"Graphics Pipe", 31},
		{"Texture Pipe", 14},
		{"Shader Export", 20},
		{"Shader Processor Interpolator", 22},
		{"Primitive Assembly", 25},
		{"Depth Block", 26},
		{"Color Block", 30},
	}
	if gen == GenGFX9 {
		return append(base, GRBMBit{"Work Distributor", 21})
	}
	return append(base, GRBMBit{"Geometry Engine", 21})
}

// GRBM2BitsForGen returns the GRBM2_STATUS bit table for the given chip generation.
// Bit positions sourced from the Linux kernel amdgpu driver register headers:
// drivers/gpu/drm/amd/include/asic_reg/gc/gc_*_sh_mask.h
func GRBM2BitsForGen(gen Generation) []GRBMBit {
	base := []GRBMBit{
		{"Unified Translation Cache Level-2", 15},
		{"Efficiency Arbiter", 16},
		{"Command Processor -  Fetcher", 28},
		{"Command Processor -  Compute", 29},
		{"Command Processor - Graphics", 30},
	}
	switch gen {
	case GenGFX9:
		return append([]GRBMBit{
			{"RunList Controller", 24},
			{"Texture Cache per Pipe", 25},
			{"Render Backend Memory Interface", 17},
		}, base...)
	case GenGFX10:
		return append([]GRBMBit{
			{"RunList Controller", 24},
			{"Texture Cache per Pipe", 25},
			{"Render Backend Memory Interface", 17},
			{"SDMA", 21},
		}, base...)
	case GenGFX12:
		return append([]GRBMBit{
			{"RunList Controller", 26},
			{"Texture Cache per Pipe", 27},
			{"SDMA", 21},
		}, base...)
	default: // GFX10.3 and GFX11 share the same layout
		return append([]GRBMBit{
			{"RunList Controller", 26},
			{"Texture Cache per Pipe", 27},
			{"Render Backend Memory Interface", 17},
			{"SDMA", 21},
		}, base...)
	}
}
