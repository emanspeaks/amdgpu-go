//go:build linux

package amdgpu

/*
#cgo !cproto pkg-config: libdrm libdrm_amdgpu

#include <xf86drm.h>
#include <amdgpu_drm.h>
#include <amdgpu.h>
#include <fcntl.h>
#include <unistd.h>
#include <errno.h>
#include <string.h>
#include <stdint.h>

// Thin C wrappers to avoid CGO pointer issues with opaque handles
// and to provide non-variadic versions of variadic libc functions
// (cgo cannot bind variadic C functions directly).

static inline int open_wrapper(const char *path, int flags) {
	return open(path, flags);
}

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

static inline int query_info_wrapper(int fd,
                                      uint32_t query,
                                      void *return_pointer,
                                      uint32_t return_size) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = query;
	request.return_pointer = (uintptr_t)return_pointer;
	request.return_size = return_size;
	return drmCommandWriteRead(fd, DRM_AMDGPU_INFO, &request,
				   sizeof(struct drm_amdgpu_info));
}

static inline int query_fw_wrapper(int fd, uint32_t fw_type, uint32_t ip_instance,
                                    uint32_t index, void *return_pointer,
                                    uint32_t return_size) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = AMDGPU_INFO_FW_VERSION;
	request.return_pointer = (uintptr_t)return_pointer;
	request.return_size = return_size;
	request.query_fw.fw_type = fw_type;
	request.query_fw.ip_instance = ip_instance;
	request.query_fw.index = index;
	return drmCommandWriteRead(fd, DRM_AMDGPU_INFO, &request,
				   sizeof(struct drm_amdgpu_info));
}

static inline int query_hw_ip_wrapper(int fd, uint32_t type, uint32_t ip_instance,
                                       void *return_pointer, uint32_t return_size) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = AMDGPU_INFO_HW_IP_INFO;
	request.return_pointer = (uintptr_t)return_pointer;
	request.return_size = return_size;
	request.query_hw_ip.type = type;
	request.query_hw_ip.ip_instance = ip_instance;
	return drmCommandWriteRead(fd, DRM_AMDGPU_INFO, &request,
				   sizeof(struct drm_amdgpu_info));
}
*/
import "C"

import (
	"fmt"
	"path/filepath"
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

// Open opens /dev/dri/renderD(128+card) and initializes the amdgpu device.
// card is the card index (0 for renderD128, 1 for renderD129, etc.).
func Open(card int) (*Device, error) {
	renderPath := filepath.Join("/dev", "dri", fmt.Sprintf("renderD%d", 128+card))

	fd := C.open_wrapper(C.CString(renderPath), C.O_RDWR)
	if fd < 0 {
		return nil, fmt.Errorf("open %s: file not found", renderPath)
	}

	var major, minor C.uint32_t
	var handle C.amdgpu_device_handle

	ret := C.init_wrapper(fd, &major, &minor, &handle)
	if ret < 0 {
		C.close(fd)
		return nil, fmt.Errorf("amdgpu_device_initialize: %w", mapErr(ret))
	}

	return &Device{
		handle: unsafe.Pointer(handle),
		fd:     int(fd),
	}, nil
}

// Close releases the device handle and closes the file descriptor.
func (d *Device) Close() error {
	C.deinit_wrapper((C.amdgpu_device_handle)(d.handle))
	C.close(C.int(d.fd))
	return nil
}

// FD returns the underlying file descriptor.
func (d *Device) FD() int {
	return d.fd
}

// DeviceInfo returns GPU identification data via AMDGPU_INFO_DEV_INFO.
func (d *Device) DeviceInfo() (*DeviceInfo, error) {
	var info C.struct_drm_amdgpu_info_device
	ret := C.query_info_wrapper(
		C.int(d.fd),
		C.AMDGPU_INFO_DEV_INFO,
		unsafe.Pointer(&info),
		C.uint(unsafe.Sizeof(info)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_DEV_INFO: %w", mapErr(ret))
	}

	return &DeviceInfo{
		Family:         uint32(info.family),
		ExternalRev:    uint32(info.external_rev),
		IsApu:          (uint64(info.ids_flags) & uint64(C.AMDGPU_IDS_FLAGS_FUSION)) != 0,
		MaxEngineClock: uint64(info.max_engine_clock),
		MaxMemoryClock: uint64(info.max_memory_clock),
		NUMTCCBlocks:   uint32(info.num_tcc_blocks),
		GL0CacheSize:   uint32(info.tcp_cache_size),
		GL1CacheSize:   uint32(info.gl1c_cache_size),
		GL2CacheSize:   uint32(info.gl2c_cache_size),
		MallSize:       uint64(info.mall_size),
	}, nil
}

// MemoryInfo returns VRAM/GTT heap info via AMDGPU_INFO_MEMORY.
func (d *Device) MemoryInfo() (*MemoryInfo, error) {
	var mem C.struct_drm_amdgpu_memory_info
	ret := C.query_info_wrapper(
		C.int(d.fd),
		C.AMDGPU_INFO_MEMORY,
		unsafe.Pointer(&mem),
		C.uint(unsafe.Sizeof(mem)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_MEMORY: %w", mapErr(ret))
	}

	return &MemoryInfo{
		VRAMTotalHeapSize:       uint64(mem.vram.total_heap_size),
		VRAMUsableHeapSize:      uint64(mem.vram.usable_heap_size),
		VRAMHeapUsage:           uint64(mem.vram.heap_usage),
		CPUAccessibleTotalSize:  uint64(mem.cpu_accessible_vram.total_heap_size),
		CPUAccessibleUsableSize: uint64(mem.cpu_accessible_vram.usable_heap_size),
		CPUAccessibleHeapUsage:  uint64(mem.cpu_accessible_vram.heap_usage),
		GTTTotalHeapSize:        uint64(mem.gtt.total_heap_size),
		GTTUsableHeapSize:       uint64(mem.gtt.usable_heap_size),
		GTTHeapUsage:            uint64(mem.gtt.heap_usage),
		ResizableBar:            mem.cpu_accessible_vram.total_heap_size >= mem.vram.total_heap_size,
	}, nil
}

// VramGttInfo returns VRAM/GTT size info via AMDGPU_INFO_VRAM_GTT.
func (d *Device) VramGttInfo() (*VramGttInfo, error) {
	var vg C.struct_drm_amdgpu_info_vram_gtt
	ret := C.query_info_wrapper(
		C.int(d.fd),
		C.AMDGPU_INFO_VRAM_GTT,
		unsafe.Pointer(&vg),
		C.uint(unsafe.Sizeof(vg)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_VRAM_GTT: %w", mapErr(ret))
	}

	return &VramGttInfo{
		VRAMSize:          uint64(vg.vram_size),
		VRAMCpuAccessible: uint64(vg.vram_cpu_accessible_size),
		GTTSize:           uint64(vg.gtt_size),
	}, nil
}

// ReadMMRegisters reads one or more MMIO registers at the given dword offset.
// The offset is in dwords (4-byte units), matching the amdgpu kernel driver's
// register addressing. For GRBM performance counters, use GRBMOFFSET.
// For GRBM2, use GRBM2OFFSET.
func (d *Device) ReadMMRegisters(offset, count uint32) ([]uint32, error) {
	if count == 0 {
		return nil, fmt.Errorf("count must be > 0")
	}

	values := make([]C.uint32_t, count)
	ret := C.read_mm_wrapper(
		(C.amdgpu_device_handle)(d.handle),
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

// DRMVersion returns DRM driver version info via drmGetVersion.
func (d *Device) DRMVersion() (*DRMVersion, error) {
	version := C.drmGetVersion(C.int(d.fd))
	if version == nil {
		return nil, fmt.Errorf("drmGetVersion: %w", ErrIO)
	}
	defer C.drmFreeVersion(version)

	return &DRMVersion{
		Name:        C.GoString(version.name),
		Version:     fmt.Sprintf("%d.%d.%d", int(version.version_major), int(version.version_minor), int(version.version_patchlevel)),
		Date:        C.GoString(version.date),
		Description: C.GoString(version.desc),
	}, nil
}

// SensorValue returns sensor readings via sysfs (/sys/class/drm/card*/device/hwmon/).
func (d *Device) SensorValue(t SENSOR_TYPE) (uint32, error) {
	return 0, fmt.Errorf("SensorValue: sysfs not yet implemented")
}

// FirmwareVersion returns the firmware version for the given firmware type.
func (d *Device) FirmwareVersion(fwType AMDGPU_INFO_FW) (*FirmwareVersion, error) {
	var info C.struct_drm_amdgpu_info_firmware
	ret := C.query_fw_wrapper(
		C.int(d.fd),
		C.uint32_t(fwType),
		0, 0,
		unsafe.Pointer(&info),
		C.uint32_t(unsafe.Sizeof(info)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_FW_VERSION: %w", mapErr(ret))
	}

	return &FirmwareVersion{
		Version: uint32(info.ver),
		Feature: uint32(info.feature),
	}, nil
}

// HWIPInfo returns hardware IP block info via AMDGPU_INFO_HW_IP_INFO.
func (d *Device) HWIPInfo(ipType HW_IP_TYPE) (*HWIPInfo, error) {
	var info C.struct_drm_amdgpu_info_hw_ip
	ret := C.query_hw_ip_wrapper(
		C.int(d.fd),
		C.uint32_t(ipType),
		0,
		unsafe.Pointer(&info),
		C.uint32_t(unsafe.Sizeof(info)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_HW_IP: %w", mapErr(ret))
	}

	ipDisc := uint32(info.ip_discovery_version)
	return &HWIPInfo{
		Type:     ipType,
		Instance: 0,
		Major:    uint32(info.hw_ip_version_major),
		Minor:    uint32(info.hw_ip_version_minor),
		Enabled:  uint32(info.available_rings) != 0,
		RevMajor: (ipDisc >> 16) & 0xFF,
		RevMinor: (ipDisc >> 8) & 0xFF,
	}, nil
}
