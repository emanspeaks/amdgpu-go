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
// On Fedora: dnf install libdrm-devel libdrm-amdgpu
// On Arch: pacman -S libdrm
//
// On Windows, only stub implementations are available that return ErrWindowsBackend.

package amdgpu
