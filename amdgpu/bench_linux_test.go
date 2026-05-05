//go:build linux && cgo

package amdgpu

import (
	"testing"
)

// BenchmarkReadGRBM measures the round-trip latency of a single GRBM_STATUS
// register read via amdgpu_read_mm_registers. Requires a real AMD GPU at
// renderD128; the benchmark is skipped when no device is available.
func BenchmarkReadGRBM(b *testing.B) {
	dev, err := Open(0)
	if err != nil {
		b.Skipf("no GPU at renderD128: %v", err)
	}
	defer dev.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := dev.ReadGRBM(); err != nil {
			b.Fatalf("ReadGRBM: %v", err)
		}
	}
}

// BenchmarkReadMMRegisters measures the latency of ReadMMRegisters for a
// single register read, isolating the CGO ioctl overhead.
func BenchmarkReadMMRegisters(b *testing.B) {
	dev, err := Open(0)
	if err != nil {
		b.Skipf("no GPU at renderD128: %v", err)
	}
	defer dev.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := dev.ReadMMRegisters(GRBMOFFSET, 1); err != nil {
			b.Fatalf("ReadMMRegisters: %v", err)
		}
	}
}
