package amdgpu

// GRBMBit maps a GRBM/GRBM2 register bit position to a display name.
type GRBMBit struct {
	Name string
	Bit  uint
}

// GRBMBits returns the GRBM_STATUS bit table used across GFX9+ hardware.
func GRBMBits() []GRBMBit {
	return []GRBMBit{
		{"Graphics Pipe", 31},
		{"Texture Pipe", 14},
		{"Shader Export", 20},
		{"Shader Processor Interpolator", 22},
		{"Primitive Assembly", 25},
		{"Depth Block", 26},
		{"Color Block", 30},
		{"Geometry Engine", 21},
	}
}

// GRBM2BitsForGen returns the GRBM2_STATUS bit table for the given chip generation.
// Bit positions sourced from amdgpu_top libamdgpu_top/src/stat/mod.rs.
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
