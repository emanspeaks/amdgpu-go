package amdgpu

import (
	"testing"
)

// TestBuild verifies the package compiles and basic types are usable.
// This test does not require a GPU and runs in CI.
func TestBuild(t *testing.T) {
	// Verify constants are defined
	if GRBMOFFSET != 0x2004 {
		t.Errorf("GRBMOFFSET = %x, want 0x2004", GRBMOFFSET)
	}
	if GRBM2OFFSET != 0x2002 {
		t.Errorf("GRBM2OFFSET = %x, want 0x2002", GRBM2OFFSET)
	}

	// Verify types exist and are usable
	var _ CHIP_CLASS = GFX10
	var _ ASIC_NAME = ASIC_UNKNOWN
	var _ HW_IP_TYPE = HW_IP_TYPE_GFX
	var _ SENSOR_TYPE = SENSOR_TYPE_GFX_SCLK

	// Verify error types exist
	var _ error = ErrPermissionDenied
	var _ error = ErrInvalidArg
	var _ error = ErrNoDevice
	var _ error = ErrIO
}


