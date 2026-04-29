package amdgpu

import "errors"

// Error values returned from C functions.
var (
	ErrPermissionDenied = errors.New("permission denied (EACCES)")
	ErrInvalidArg       = errors.New("invalid argument (EINVAL)")
	ErrNoDevice         = errors.New("no such device (ENODEV)")
	ErrIO               = errors.New("I/O error (EIO)")
)
