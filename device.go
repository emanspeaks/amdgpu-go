package amdgpu

/*
#include <xf86drm.h>
#include <amdgpu.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

// Device wraps an initialized amdgpu device handle.
type Device struct {
	handle C.amdgpu_device_handle
	fd     C.int
}

// Open opens /dev/dri/renderD128 + card and initializes the amdgpu device.
// card is the card number (0 for card0/renderD128, 1 for card1/renderD130, etc.).
func Open(card int) (*Device, error) {
	renderName := fmt.Sprintf("renderD128%d", card)
	renderPath := filepath.Join("/dev", "dri", renderName)

	fd, err := C.open(C.CString(renderPath), C.O_RDWR, 0)
	if fd < 0 {
		return nil, fmt.Errorf("open %s: %w", renderPath, os.ErrNotExist)
	}

	var major, minor C.uint32_t
	var handle C.amdgpu_device_handle

	ret := C.init_wrapper(fd, &major, &minor, &handle)
	if ret < 0 {
		C.close(fd)
		return nil, fmt.Errorf("amdgpu_device_initialize: %w", mapErr(ret))
	}

	return &Device{
		handle: handle,
		fd:     fd,
	}, nil
}

// Close releases the device handle and closes the file descriptor.
func (d *Device) Close() error {
	C.deinit_wrapper(d.handle)
	C.close(d.fd)
	return nil
}

// FD returns the underlying file descriptor.
func (d *Device) FD() int {
	return int(d.fd)
}

// String returns a human-readable representation of the device.
func (d *Device) String() string {
	return fmt.Sprintf("amdgpu.Device{fd:%d}", d.fd)
}
