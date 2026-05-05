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

static inline const char* get_marketing_name_wrapper(amdgpu_device_handle handle) {
	return amdgpu_get_marketing_name(handle);
}

static inline int query_hw_ip_count_wrapper(int fd, uint32_t type, uint32_t *count) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = AMDGPU_INFO_HW_IP_COUNT;
	request.return_pointer = (uintptr_t)count;
	request.return_size = sizeof(*count);
	request.query_hw_ip.type = type;
	return drmCommandWriteRead(fd, DRM_AMDGPU_INFO, &request,
				   sizeof(struct drm_amdgpu_info));
}

static inline int query_vbios_str_wrapper(int fd, uint32_t type, char *buf, uint32_t bufsize) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = AMDGPU_INFO_VBIOS;
	request.return_pointer = (uintptr_t)buf;
	request.return_size = bufsize;
	request.vbios_info.type = type;
	return drmCommandWriteRead(fd, DRM_AMDGPU_INFO, &request,
				   sizeof(struct drm_amdgpu_info));
}

static inline int query_video_caps_wrapper(int fd, uint32_t type,
                                            void *return_pointer, uint32_t return_size) {
	struct drm_amdgpu_info request;
	memset(&request, 0, sizeof(request));
	request.query = AMDGPU_INFO_VIDEO_CAPS;
	request.return_pointer = (uintptr_t)return_pointer;
	request.return_size = return_size;
	request.sensor_info.type = type; // overlaps video_cap.type
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

	numCUPerSH := uint32(info.num_cu_per_sh)
	var marketingName string
	if s := C.get_marketing_name_wrapper((C.amdgpu_device_handle)(d.handle)); s != nil {
		marketingName = C.GoString(s)
	}
	return &DeviceInfo{
		Family:                   uint32(info.family),
		ExternalRev:              uint32(info.external_rev),
		MarketingName:            marketingName,
		IsApu:                    (uint64(info.ids_flags) & uint64(C.AMDGPU_IDS_FLAGS_FUSION)) != 0,
		MaxEngineClock:           uint64(info.max_engine_clock),
		MaxMemoryClock:           uint64(info.max_memory_clock),
		NUMTCCBlocks:             uint32(info.num_tcc_blocks),
		NumShaderEngines:         uint32(info.num_shader_engines),
		NumShaderArraysPerEngine: uint32(info.num_shader_arrays_per_engine),
		NumCUPerSH:               numCUPerSH,
		GL0CacheSize:             uint32(info.tcp_cache_size),
		GL1CacheSize:             uint32(info.gl1c_cache_size),
		GL2CacheSize:             uint32(info.gl2c_cache_size),
		MallSize:                 uint64(info.mall_size),
		VRAMType:                 uint32(info.vram_type),
		VRAMBitWidth:             uint32(info.vram_bit_width),
		NumRBPipes:               uint32(info.num_rb_pipes),
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
	rings := uint32(info.available_rings)
	return &HWIPInfo{
		Type:           ipType,
		Instance:       0,
		Major:          uint32(info.hw_ip_version_major),
		Minor:          uint32(info.hw_ip_version_minor),
		Enabled:        rings != 0,
		AvailableRings: rings,
		RevMajor:       (ipDisc >> 16) & 0xFF,
		RevMinor:       (ipDisc >> 8) & 0xFF,
	}, nil
}

// HWIPCount returns the number of instances of the given hardware IP type.
func (d *Device) HWIPCount(ipType HW_IP_TYPE) uint32 {
	var count C.uint32_t
	if ret := C.query_hw_ip_count_wrapper(C.int(d.fd), C.uint32_t(ipType), &count); ret < 0 {
		return 0
	}
	return uint32(count)
}

// VBIOSInfo returns VBIOS identification strings via AMDGPU_INFO_VBIOS_INFO ioctl.
// The kernel returns a drm_amdgpu_info_vbios struct:
//   name[64] + vbios_pn[64] + version(4) + pad(4) + vbios_ver_str[32] + date[32] = 200 bytes
func (d *Device) VBIOSInfo() (*VBIOSInfo, error) {
	const vbiosInfo C.uint32_t = 3 // AMDGPU_INFO_VBIOS_INFO
	var buf [200]C.char
	ret := C.query_vbios_str_wrapper(C.int(d.fd), vbiosInfo, &buf[0], C.uint32_t(len(buf)))
	if ret < 0 {
		return nil, fmt.Errorf("VBIOS info: %w", mapErr(ret))
	}
	cstr := func(start, end int) string {
		b := (*[1 << 20]byte)(unsafe.Pointer(&buf[0]))[start:end]
		for i, c := range b {
			if c == 0 {
				return string(b[:i])
			}
		}
		return string(b)
	}
	name := cstr(0, 64)
	pn := cstr(64, 128)
	if name == "" && pn == "" {
		return nil, fmt.Errorf("VBIOS info not available")
	}
	return &VBIOSInfo{
		Name:   name,
		PN:     pn,
		VerStr: cstr(136, 168),
		Date:   cstr(168, 200),
	}, nil
}

// VideoCaps returns video encode or decode capability dimensions per codec.
// pass 0 for decode, 1 for encode (AMDGPU_INFO_VIDEO_CAPS_DECODE/ENCODE).
// drm_amdgpu_info_video_codec_info: valid(4)+max_width(4)+max_height(4)+pixels(4)+level(4)+pad(4) = 24 bytes each
func (d *Device) VideoCaps(capType uint32) (*VideoCapsInfo, error) {
	var raw [8 * 6]C.uint32_t // 8 codecs × 6 uint32 fields each
	ret := C.query_video_caps_wrapper(
		C.int(d.fd),
		C.uint32_t(capType),
		unsafe.Pointer(&raw[0]),
		C.uint32_t(unsafe.Sizeof(raw)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_VIDEO_CAPS: %w", mapErr(ret))
	}
	var out VideoCapsInfo
	for i := 0; i < 8; i++ {
		base := i * 6
		if raw[base] == 0 { // valid == 0: codec not supported
			continue
		}
		entry := VideoCapEntry{
			MaxWidth:  uint32(raw[base+1]),
			MaxHeight: uint32(raw[base+2]),
		}
		if capType == 0 {
			out.Decode[i] = entry
		} else {
			out.Encode[i] = entry
		}
	}
	return &out, nil
}
