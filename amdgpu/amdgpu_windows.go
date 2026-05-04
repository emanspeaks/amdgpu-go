//go:build windows

package amdgpu

import "fmt"

// Open returns ErrWindowsBackend on Windows.
func Open(card int) (*Device, error) {
	return nil, fmt.Errorf("%w: Open", ErrWindowsBackend)
}

// Close is a stub on Windows.
func (d *Device) Close() error {
	return ErrWindowsBackend
}

// FD returns -1 on Windows.
func (d *Device) FD() int {
	return -1
}

// DeviceInfo returns ErrWindowsBackend on Windows.
func (d *Device) DeviceInfo() (*DeviceInfo, error) {
	return nil, fmt.Errorf("%w: DeviceInfo", ErrWindowsBackend)
}

// MemoryInfo returns ErrWindowsBackend on Windows.
func (d *Device) MemoryInfo() (*MemoryInfo, error) {
	return nil, fmt.Errorf("%w: MemoryInfo", ErrWindowsBackend)
}

// VramGttInfo returns ErrWindowsBackend on Windows.
func (d *Device) VramGttInfo() (*VramGttInfo, error) {
	return nil, fmt.Errorf("%w: VramGttInfo", ErrWindowsBackend)
}

// ReadMMRegisters returns ErrWindowsBackend on Windows.
func (d *Device) ReadMMRegisters(offset, count uint32) ([]uint32, error) {
	if count == 0 {
		return nil, fmt.Errorf("count must be > 0")
	}
	return nil, fmt.Errorf("ReadMMRegisters: %w", ErrWindowsBackend)
}

// DRMVersion returns ErrWindowsBackend on Windows.
func (d *Device) DRMVersion() (*DRMVersion, error) {
	return nil, fmt.Errorf("%w: DRMVersion", ErrWindowsBackend)
}

// SensorValue returns ErrWindowsBackend on Windows.
func (d *Device) SensorValue(t SENSOR_TYPE) (uint32, error) {
	return 0, fmt.Errorf("%w: SensorValue", ErrWindowsBackend)
}

// FirmwareVersion returns ErrWindowsBackend on Windows.
func (d *Device) FirmwareVersion(fwType AMDGPU_INFO_FW) (*FirmwareVersion, error) {
	return nil, fmt.Errorf("%w: FirmwareVersion", ErrWindowsBackend)
}

// HWIPInfo returns ErrWindowsBackend on Windows.
func (d *Device) HWIPInfo(ipType HW_IP_TYPE) (*HWIPInfo, error) {
	return nil, fmt.Errorf("%w: HWIPInfo", ErrWindowsBackend)
}
