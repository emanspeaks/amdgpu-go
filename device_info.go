package amdgpu

/*
#include <amdgpu.h>
#include <stdint.h>
#include <string.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// DeviceInfo contains GPU identification and capability data.
type DeviceInfo struct {
	ASICName         ASIC_NAME
	ChipClass        CHIP_CLASS
	IsApu            bool
	MaxEngineClock   uint32 // in 10kHz units
	MaxMemoryClock   uint32 // in 10kHz units
	L1CacheSize      uint32 // bytes
	GL1CacheSize     uint32 // bytes
	L2CacheSize      uint32 // bytes
	L3CacheSizeMB    uint32 // MiB
	NUMTCCBlocks     uint32
}

// DeviceInfo returns GPU identification data via AMDGPU_INFO_GET_INFO.
func (d *Device) DeviceInfo() (*DeviceInfo, error) {
	var info C.struct_drm_amdgpu_info_device
	ret := C.query_info_wrapper(
		C.int(d.fd),
		C.AMDGPU_INFO_GET_INFO,
		unsafe.Pointer(&info),
		C.uint(unsafe.Sizeof(info)),
	)
	if ret < 0 {
		return nil, fmt.Errorf("AMDGPU_INFO_GET_INFO: %w", mapErr(ret))
	}

	return &DeviceInfo{
		ASICName:       ASIC_NAME(info.asic_name),
		ChipClass:      CHIP_CLASS(info.chip_class),
		IsApu:          info.is_apu != 0,
		MaxEngineClock: uint32(info.max_engine_clock),
		MaxMemoryClock: uint32(info.max_memory_clock),
	}, nil
}

// MemoryInfo contains VRAM/GTT heap information.
type MemoryInfo struct {
	VRAMHeapUsage      uint64
	VRAMTotalHeapSize  uint64
	VRAMUsableHeapSize uint64
	GTTHeapUsage       uint64
	GTTTotalHeapSize   uint64
	GTTUsableHeapSize  uint64
	ResizableBar       bool
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
		VRAMHeapUsage:      uint64(mem.vram.heap_usage),
		VRAMTotalHeapSize:  uint64(mem.vram.total_heap_size),
		VRAMUsableHeapSize: uint64(mem.vram.usable_heap_size),
		GTTHeapUsage:       uint64(mem.gtt.heap_usage),
		GTTTotalHeapSize:   uint64(mem.gtt.total_heap_size),
		GTTUsableHeapSize:  uint64(mem.gtt.usable_heap_size),
		ResizableBar:       mem.vram.usable_heap_size < mem.vram.total_heap_size,
	}, nil
}

// VramGttInfo contains usable VRAM/GTT heap sizes.
type VramGttInfo struct {
	VRAMTotal       uint64
	VRAMUsable      uint64
	GTTTotal        uint64
	GTTUsable       uint64
}

// VramGttInfo returns usable VRAM/GTT sizes via AMDGPU_INFO_VRAM_GTT.
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
		VRAMTotal: uint64(vg.vram_total),
		VRAMUsable: uint64(vg.vram_usable),
		GTTTotal: uint64(vg.gtt_total),
		GTTUsable: uint64(vg.gtt_usable),
	}, nil
}
