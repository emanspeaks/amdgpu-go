//go:build windows

package amdgpu

import "fmt"

// DeviceInfo contains GPU identification and capability data.
type DeviceInfo struct {
	ASICName       ASIC_NAME
	ChipClass      CHIP_CLASS
	IsApu          bool
	MaxEngineClock uint32 // in 10kHz units
	MaxMemoryClock uint32 // in 10kHz units
	L1CacheSize    uint32 // bytes
	GL1CacheSize   uint32 // bytes
	L2CacheSize    uint32 // bytes
	L3CacheSizeMB  uint32 // MiB
	NUMTCCBlocks   uint32
}

// DeviceInfo returns a Windows-specific not-implemented error for now.
func (d *Device) DeviceInfo() (*DeviceInfo, error) {
	return nil, fmt.Errorf("DeviceInfo: %w", ErrWindowsBackend)
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

// MemoryInfo returns a Windows-specific not-implemented error for now.
func (d *Device) MemoryInfo() (*MemoryInfo, error) {
	return nil, fmt.Errorf("MemoryInfo: %w", ErrWindowsBackend)
}

// VramGttInfo contains usable VRAM/GTT heap sizes.
type VramGttInfo struct {
	VRAMTotal  uint64
	VRAMUsable uint64
	GTTTotal   uint64
	GTTUsable  uint64
}

// VramGttInfo returns a Windows-specific not-implemented error for now.
func (d *Device) VramGttInfo() (*VramGttInfo, error) {
	return nil, fmt.Errorf("VramGttInfo: %w", ErrWindowsBackend)
}
