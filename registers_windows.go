//go:build windows

package amdgpu

import "fmt"

// ReadMMRegisters returns a Windows-specific not-implemented error for now.
func (d *Device) ReadMMRegisters(offset, count uint32) ([]uint32, error) {
	if count == 0 {
		return nil, fmt.Errorf("count must be > 0")
	}
	return nil, fmt.Errorf("ReadMMRegisters: %w", ErrWindowsBackend)
}

// ReadGRBM returns a Windows-specific not-implemented error for now.
func (d *Device) ReadGRBM() (uint32, error) {
	return 0, fmt.Errorf("ReadGRBM: %w", ErrWindowsBackend)
}

// ReadGRBM2 returns a Windows-specific not-implemented error for now.
func (d *Device) ReadGRBM2() (uint32, error) {
	return 0, fmt.Errorf("ReadGRBM2: %w", ErrWindowsBackend)
}
