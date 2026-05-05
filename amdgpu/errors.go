package amdgpu

import "errors"

var (
	// ErrPermissionDenied is returned when the DRM ioctl fails with EACCES.
	ErrPermissionDenied = errors.New("permission denied (EACCES)")
	// ErrInvalidArg is returned when the DRM ioctl fails with EINVAL.
	ErrInvalidArg = errors.New("invalid argument (EINVAL)")
	// ErrNoDevice is returned when the DRM ioctl fails with ENODEV.
	ErrNoDevice = errors.New("no such device (ENODEV)")
	// ErrIO is returned when the DRM ioctl fails with EIO.
	ErrIO = errors.New("I/O error (EIO)")
	// ErrWindowsBackend is returned by all Device methods on Windows until the
	// InpOut32 backend (Phase 6) is implemented.
	ErrWindowsBackend = errors.New("windows backend not implemented")
)
