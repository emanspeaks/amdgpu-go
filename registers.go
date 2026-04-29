//go:build linux

package amdgpu

/*
#include <libdrm/amdgpu.h>
#include <stdint.h>
*/
import "C"

import (
	"fmt"
)

// ReadMMRegisters reads one or more MMIO registers at the given dword offset.
// Returns the register values as a slice of uint32.
//
// The offset is in dwords (4-byte units), matching the amdgpu kernel driver's
// register addressing. For GRBM performance counters, use GRBMOFFSET.
// For GRBM2, use GRBM2OFFSET.
func (d *Device) ReadMMRegisters(offset, count uint32) ([]uint32, error) {
	if count == 0 {
		return nil, fmt.Errorf("count must be > 0")
	}

	values := make([]C.uint32_t, count)
	ret := C.read_mm_wrapper(
		d.handle,
		C.uint(offset),
		C.uint(count),
		C.uint32_t(GRBM_INSTANCE),
		0, // flags
		&values[0],
	)
	if ret < 0 {
		return nil, fmt.Errorf("amdgpu_read_mm_registers: %w", mapErr(ret))
	}

	result := make([]uint32, count)
	for i := range values {
		result[i] = uint32(values[i])
	}
	return result, nil
}

// ReadGRBM reads the GRBM_STATUS register and returns the 32-bit value.
func (d *Device) ReadGRBM() (uint32, error) {
	values, err := d.ReadMMRegisters(GRBMOFFSET, 1)
	if err != nil {
		return 0, err
	}
	return values[0], nil
}

// ReadGRBM2 reads the GRBM2_STATUS2 register and returns the 32-bit value.
func (d *Device) ReadGRBM2() (uint32, error) {
	values, err := d.ReadMMRegisters(GRBM2OFFSET, 1)
	if err != nil {
		return 0, err
	}
	return values[0], nil
}
