//go:build windows

package amdgpu

import "fmt"

// Device is a placeholder for a future native Windows backend.
type Device struct {
	fd int
}

// Open returns a Windows-specific not-implemented error for now.
func Open(card int) (*Device, error) {
	return nil, fmt.Errorf("open card %d: %w", card, ErrWindowsBackend)
}

// Close is a no-op for the Windows stub backend.
func (d *Device) Close() error {
	return nil
}

// FD returns 0 for the Windows stub backend.
func (d *Device) FD() int {
	return 0
}

// String returns a human-readable representation of the device.
func (d *Device) String() string {
	return "amdgpu.Device{windows-stub}"
}
