// Package amdgpu provides Go bindings for the libdrm_amdgpu C library.
//
// This package wraps the low-level AMDGPU DRM ioctls to provide access to
// GPU device information, memory stats, sensor readings, and MMIO register
// reads (used for GRBM/GRBM2 performance counters).
//
// # Build Requirements
//
// This package requires CGO and the following system libraries:
//   - libdrm (pkg-config: libdrm)
//   - libdrm_amdgpu (pkg-config: libdrm_amdgpu)
//
// On Debian/Ubuntu: sudo apt install libdrm-dev libdrm-amdgpu1
// On Fedora: sudo dnf install libdrm-devel libdrm-amdgpu
// On Arch: sudo pacman -S libdrm
package amdgpu

/*
#cgo pkg-config: libdrm libdrm_amdgpu

#include <xf86drm.h>
#include <amdgpu.h>
#include <stdint.h>

// Thin C wrappers to avoid CGO pointer issues with opaque handles

static inline int init_wrapper(int fd, uint32_t *major, uint32_t *minor,
                               amdgpu_device_handle *handle) {
	return amdgpu_device_initialize(fd, major, minor, handle);
}

static inline void deinit_wrapper(amdgpu_device_handle handle) {
	amdgpu_device_deinitialize(handle);
}

static inline int read_mm_wrapper(amdgpu_device_handle handle,
                                  unsigned offset, unsigned count,
                                  uint32_t instance, uint32_t flags,
                                  uint32_t *values) {
	return amdgpu_read_mm_registers(handle, offset, count, instance, flags, values);
}

static inline int query_info_wrapper(amdgpu_device_handle handle,
                                     uint32_t query,
                                     void *return_pointer,
                                     uint32_t return_size) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = query;
	request.return_pointer = (uintptr_t)return_pointer;
	request.return_size = return_size;
	return drmCommandWriteRead(handle->fd, DRM_AMDGPU_INFO, &request,
				   sizeof(struct drm_amdgpu_info));
}

static inline int sensor_wrapper(amdgpu_device_handle handle,
				 uint32_t sensor_type, uint32_t *value) {
	return amdgpu_sensor_get_value(handle, sensor_type, value);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// mapErr converts a negative C error code to a Go error.
func mapErr(ret C.int) error {
	if ret >= 0 {
		return nil
	}
	switch -int(ret) {
	case int(C.EACCES):
		return ErrPermissionDenied
	case int(C.EINVAL):
		return ErrInvalidArg
	case int(C.ENODEV):
		return ErrNoDevice
	case int(C.EIO):
		return ErrIO
	default:
		return fmt.Errorf("drm error %d", -int(ret))
	}
}
