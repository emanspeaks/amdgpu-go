# amdgpu-go

Go bindings for the `libdrm_amdgpu` C library. Provides low-level access to AMD GPU devices via DRM ioctls.

## Features

- Device initialization and cleanup
- MMIO register reads (GRBM/GRBM2 performance counters)
- Device info queries (ASIC name, chip class, clocks)
- Memory info queries (VRAM/GTT heap usage)
- Sensor reads (SCLK, MCLK, voltages, temperatures)

## Build Requirements

This package requires CGO and the following system libraries:

- `libdrm` (pkg-config: `libdrm`)
- `libdrm_amdgpu` (pkg-config: `libdrm_amdgpu`)

### Install dependencies

```bash
# Debian/Ubuntu
sudo apt install libdrm-dev libdrm-amdgpu1

# Fedora
sudo dnf install libdrm-devel libdrm-amdgpu

# Arch
sudo pacman -S libdrm
```

## Usage

```go
import "github.com/emanspeaks/amdgpu-go"

dev, err := amdgpu.Open(0) // opens /dev/dri/renderD128
if err != nil {
    log.Fatal(err)
}
defer dev.Close()

// Read GRBM performance counter
val, err := dev.ReadGRBM()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("GRBM_STATUS = 0x%08x\n", val)

// Read device info
info, err := dev.DeviceInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("ASIC: %v, Class: %v\n", info.ASICName, info.ChipClass)

// Read sensor
sclk, err := dev.SensorValue(amdgpu.SENSOR_TYPE_GFX_SCLK)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("SCLK: %d kHz\n", sclk)
```

## License

MIT
