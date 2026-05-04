# amdgpu-go Go cgo bindings — TODO List

## Phase 0: Repo scaffolding

- [x] **0.1** Initialize `amdgpu-go` git submodule at `./amdgpu_go`
- [x] **0.2** Create `amdgpu-go/go.mod` with module `github.com/emanspeaks/amdgpu-go`
- [x] **0.3** Add `.gitignore` for `amdgpu-go/` (build artifacts, etc.)
- [x] **0.4** Verify `go build ./...` passes in the empty module
- [x] **0.5** Add `amd/amdgpu` placeholder import in atopweb's `go.mod` replace directive

## Phase 1: Core bindings (GRBM/GRBM2 reads)

- [x] **1.1** Research C header constants — `GRBM_OFFSET = 0x2004`, `GRBM2_OFFSET = 0x2002`
- [x] **1.2** Create `amdgpu.go` — CGO preamble with `#cgo pkg-config: libdrm libdrm_amdgpu`, include headers, thin C wrapper functions for `init_wrapper`, `deinit_wrapper`, `read_mm_wrapper`
- [x] **1.3** Create `device.go` — `Open(card int)` that opens `/dev/dri/renderD128 + card`, calls `init_wrapper`, returns `*Device`; `Close()` that calls `deinit_wrapper`; `FD()` accessor
- [x] **1.4** Create `registers.go` — `ReadMMRegisters(offset, count uint32) ([]uint32, error)` that allocates a C uint32 array, calls `read_mm_wrapper`, converts to Go slice
- [x] **1.5** Create `enums.go` — `CHIP_CLASS`, `ASIC_NAME`, `HW_IP_TYPE` enum constants (copy from amdgpu_top's Rust definitions)
- [x] **1.6** Create `constants.go` — `GRBM_OFFSET = 0x2004`, `GRBM2_OFFSET` (verified in 1.1), `GRBM_INSTANCE = 0xffffffff`
- [x] **1.7** Create `amdgpu_test.go` — basic build test (compile-only, no GPU needed); GPU-dependent test deferred to Linux CI
- [x] **1.8** Verify `go build ./...` — passes without CGO; will pass on Linux with `libdrm-dev` + `libdrm-amdgpu-dev` installed

## Phase 1.5: Cross-platform refactor (file realignment)

- [x] **1.5.1** Create `amdgpu_linux.go`
- [x] **1.5.2** Create `device_linux.go`
- [x] **1.5.3** Create `deviceinfo_linux.go`
- [x] **1.5.4** Create `registers_linux.go`
- [x] **1.5.5** Create `device_win.go`
- [x] **1.5.6** Create `deviceinfo_win.go`
- [x] **1.5.7** Create `registers_win.go`
- [x] **1.5.8** Update `amdgpu.go`
- [x] **1.5.9** Update `device.go`
- [x] **1.5.10** Update `deviceinfo.go`
- [x] **1.5.11** Update `registers.go`
- [x] **1.5.12** Update build tags on cross-platform files
- [x] **1.5.13** Update `amdgpu_test.go` build tag
- [x] **1.5.14** Delete old `*_windows.go` files
- [x] **1.5.15** Verify `go vet ./...` passes

## Phase 2: Device info + memory

- [x] **2.1** Research C struct layouts — fields mapped from amdgpu.h structs
- [x] **2.2** Add `DeviceInfo` Go struct + `DeviceInfo()` method — wraps `AMDGPU_INFO_DEV_INFO`, extracts: asic_name, chip_class, is_apu, max_engine_clock, max_memory_clock
- [x] **2.3** Add `MemoryInfo` Go struct + `MemoryInfo()` method — wraps `AMDGPU_INFO_MEMORY`, extracts: VRAM/GTT heap_usage, total_heap_size, usable_heap_size, resizable_bar
- [x] **2.4** Add `VramGttInfo` Go struct + `VramGttInfo()` method — wraps `AMDGPU_INFO_VRAM_GTT`
- [x] **2.5** Add `DRMVersion()` method — wraps `drmGetVersion(fd)`
- [x] **2.5.1** Add Windows alternative for `DRMVersion()` — WMI or registry query
- [x] **2.6** Add error mapping — `mapErr()` converts negative C return codes to named Go errors (`ErrPermissionDenied`, `ErrInvalidArg`, `ErrNoDevice`, `ErrIO`)
- [x] **2.7** Verify `go build ./...` — will pass on Linux (fails on Windows as expected)

## Phase 3: Sensors + firmware (P2 features)

- [x] **3.1** Add `SensorType` enum — `GFX_SCLK`, `GFX_MCLK`, `VDDNB`, `VDDGFX`, temperatures (merged into enums.go as `SENSOR_TYPE`)
- [x] **3.2** Add `SensorValue(t SensorType) (uint32, error)` — wraps `amdgpu_sensor_get_value` or reads via sysfs (amdgpu_sensor_get_value not available in installed libdrm)
- [x] **3.3** Add firmware version query — `FirmwareVersion(fwType) (uint32, error)` wrapping `amdgpu_query_info(AMDGPU_INFO_FW_*)`
- [x] **3.4** Add HW IP info — `HWIPInfo(ipType) (*HWIPInfo, error)` wrapping `amdgpu_query_info(AMDGPU_INFO_HW_IP)`
- [x] **3.5** Verify `go build ./...`

## Phase 4: atopweb integration

- [ ] **4.1** Create package that imports `amdgpu-go`, builds `GPUStats` struct mirroring what atopweb currently parses from amdgpu_top JSON
- [ ] **4.2** Implement GRBM/GRBM2 reader — uses `ReadMMRegisters` + bit-to-metric mapping from `amdgpu_top/crates/libamdgpu_top/src/stat/mod.rs`
- [ ] **4.3** Implement device enumeration — find all `/dev/dri/renderD*`, map to card numbers, open each
- [ ] **4.4** Wire into atopweb's existing data pipeline — replace amdgpu_top JSON parsing with direct calls to `amdgpu-go`
- [ ] **4.5** Add `--no-pc` equivalent flag — skip GRBM reads if configured
- [ ] **4.6** Add graceful degradation — if `amdgpu-go` fails to open a device, fall back to amdgpu_top JSON for that device
- [ ] **4.7** Integration smoke test on real hardware (CI runner)

## Phase 5: Hardening + cleanup

- [ ] **5.1** Add `go doc` comments for all exported types/functions
- [ ] **5.2** Add `ExampleOpen` doc test (with `//go:build ignore`)
- [ ] **5.3** Add `Makefile` or `justfile` with `build`, `test`, `lint` targets
- [ ] **5.4** Add `.github/workflows/go.yml` — `go build`, `go vet`, `staticcheck` (no GPU needed)
- [ ] **5.4.1** Add Windows build verification to CI — `go build ./...` on Windows runner
- [ ] **5.4.2** Add Windows-specific documentation — InpOut32 setup, driver signing, feature parity table
- [ ] **5.4.3** Add feature parity checklist between Linux and Windows backends
- [ ] **5.5** Audit CGO memory management — ensure no leaks on rapid open/close cycles
- [ ] **5.6** Benchmark GRBM read latency — verify sub-millisecond per-read
- [ ] **5.7** Remove amdgpu_top dependency from atopweb (or keep as fallback)

## Phase 6: Windows backend (InpOut32 + SetupAPI + D3DKMT)

- [ ] **6.1** Add `ErrNotImplemented` to `errors.go` (`//go:build linux || windows`)
- [ ] **6.2** Download InpOut32 binary (`inpoutx64.dll`) from highrez.co.uk
- [ ] **6.3** Download InpOut32 source from highrez.co.uk (in case we need to rebuild)
- [ ] **6.4** Add `drivers/LICENSES/InpOut32.txt` — MIT license text
- [ ] **6.5** Flesh out `device_windows.go` — replace stubs with real InpOut32 + SetupAPI:
  - `Open(card int)`: Load InpOut32 driver → SetupAPI PCI scan → find GPU BAR5 → verify MMIO accessible
  - `Close()`: Deinitialize InpOut32
  - `FD()`: Return D3D device handle (opaque int)
  - `String()`: Return device description
- [ ] **6.6** Flesh out `registers_windows.go` — replace stubs with real MMIO reads:
  - `ReadMMRegisters(offset, count uint32)`: `InpOut32.GetPhysLong(BAR5_phys + offset*4)` for each dword
  - `ReadGRBM()`: Read BAR5 + 0x2004 via InpOut32
  - `ReadGRBM2()`: Read BAR5 + 0x2002 via InpOut32
- [ ] **6.7** Flesh out `deviceinfo_windows.go` — replace stubs with real PCI + SMN + D3DKMT:
  - `DeviceInfo()`: PCI vendor/device ID → ASIC mapping + SMN via PCI indirect (probe 0xE0/0xE4, fallback to 0x38/0x3C, then 0x60/0x64)
  - `MemoryInfo()`: Total from PCI BAR0 size + usage from D3DKMTQueryStatistics
  - `VramGttInfo()`: Total from PCI BAR0, usable/usage = 0, GTT = `ErrNotImplemented`
- [ ] **6.8** Update `readme.md` — cross-platform docs, Windows setup requirements, feature parity table
- [ ] **6.9** Verify `go build ./...` on Windows — InpOut32-only, no CGO needed for Windows
- [ ] **6.9.1** Test Windows stubs return correct `ErrWindowsBackend` errors

---

## Task dependencies

```text
Phase 0 ──► Phase 1.8 ──► Phase 2.7 ──► Phase 3.5 ──► Phase 4 ──► Phase 5
```

Each phase gates on `go build ./...` succeeding. No task in a phase can start until the previous phase's build verification passes.

## Key risks to watch

1. **C struct layout mismatch** — Go `struct` fields must match C struct field order and alignment exactly. We'll need `unsafe.Sizeof`/`unsafe.Offsetof` checks or `//go:linkname` to verify.

2. **pkg-config on CI runner** — The CI runner needs `libdrm-dev` and `libdrm-amdgpu-dev` installed. We should document this in a `README.md` for the amdgpu-go submodule.

3. **Pointer passing across CGO boundary** — `ReadMMRegisters` allocates a C array, passes pointer to C, then copies back to Go. Must ensure the C array lives long enough and is freed properly.

4. **Multiple concurrent devices** — If atopweb monitors multiple GPUs, each needs its own `Device`. We should test that concurrent `ReadMMRegisters` calls don't interfere.
5. **InpOut32 driver requirements** — Windows driver installation, elevation/admin rights, driver signing
6. **PCI BAR access on Windows** — permission levels, UAC, potential conflicts with other drivers
7. **D3DKMT availability** — not available on all Windows editions (e.g., Home vs Pro)
