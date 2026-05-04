package amdgpu

import (
	"fmt"
	"unsafe"
)

// Device wraps an initialized amdgpu device handle.
type Device struct {
	handle unsafe.Pointer
	fd     int
}

// String returns a human-readable representation of the device.
func (d *Device) String() string {
	return fmt.Sprintf("amdgpu.Device{fd:%d}", d.fd)
}
