# amdgpu-go — Agent Instructions

## Repo at a glance

Go bindings for `libdrm_amdgpu` C.
Module: `github.com/emanspeaks/amdgpu-go`. Go 1.25.0.

## Commands

```bash
go vet ./...              # lint (CI)
go build -v -o atopweb .  # build check (CI) — NOTE: has no main.go, will fail
go test -v                # compile-only test, no GPU needed
go mod tidy               # update go.sum (Nix workflow)
```

**CI build is broken**: `.github/workflows/ci.yml` runs `go build -v -o atopweb .`, but there is no `package main` in this repo. The root package is a library only. Fix by either adding a `main.go` or changing the CI command to `go build -v ./...`.

## Build tags

- Base (`linux || windows`): `amdgpu.go`, `device.go`, `deviceinfo.go`, `registers.go`, `constants.go`, `enums.go`, `errors.go`, `version.go`, `amdgpu_test.go`
- Linux (`linux`): `amdgpu_linux.go` — single file containing all cgo bindings (preamble with `static inline` C wrappers + every Go-side function). Consolidated because cgo does NOT combine preambles across files in a package — each file's preamble is its own translation unit, so wrappers defined in one file are not visible to others. Splitting cgo across multiple files requires duplicating wrappers or using a separate `.c` file.
- Windows (`windows`): `amdgpu_windows.go` — single file with all stubs returning `ErrWindowsBackend`. Phase 6 replaces with real cgo impl (likely InpOut32 + SetupAPI + D3DKMT); can re-split by concern at that point if it gets unwieldy.
- Editor headers (`cproto` tag + mesa/drm clone): On Windows, gopls uses `-tags=cproto` (see workspace settings). `#cgo cproto CFLAGS: -DCPROTO_BUILD` injects the define, and `#ifdef CPROTO_BUILD` in the preamble selects `<drm/amdgpu_drm.h>` + `<amdgpu/amdgpu.h>` (mesa source paths) vs `<libdrm/amdgpu_drm.h>` + `<libdrm/amdgpu.h>` (real pkg-config installed paths on Linux). All headers come from the mesa/drm git clone at `/c/scratch/mesa/drm` (machine-local, not committed). The `zig-cc.cmd` wrapper adds the three mesa include dirs; `CGO_CFLAGS` in the workspace adds the same three. No fabricated stub files — all headers are real.

## CGO patterns

- Thin C `static inline` wrappers in `amdgpu_linux.go` avoid CGO pointer issues with opaque `amdgpu_device_handle` and provide non-variadic versions of variadic libc functions (`open_wrapper` — cgo cannot bind variadic C functions directly)
- `mapErr()` converts negative C return codes to named Go errors (EACCES, EINVAL, ENODEV, EIO)
- `ReadMMRegisters` allocates a `[]C.uint32_t`, passes pointer to C, copies back to Go slice
- `query_info_wrapper` builds `struct drm_amdgpu_info` manually with `memset` + direct field assignment

## Versioning

- `VERSION` file contains semver (currently `0.1.0`)
- `version.go` embeds it via `//go:embed VERSION` → `amdgpu.Version`
- **Every PR to main must bump VERSION** (enforced by CI `version-bump` job)
- Release workflow reads VERSION, creates git tag, builds linux/amd64 + linux/arm64 binaries with `CGO_ENABLED=0`

## CI

- `ci.yml`: exemption check → version-bump → build (go vet + go build)
- `release.yml`: bumps tag from VERSION, publishes linux binaries as GitHub release
- `gomod2nix.yml`: runs `go mod tidy` + `gomod2nix generate`, commits go.sum + gomod2nix.toml
- `.github/workflows/exempt.txt`: patterns that skip version-bump check (readme.md, assets/*)

## Editor setup

- `.editorconfig`: tabs for `.go`, spaces (2) for everything else
- `amdgpu-go.code-workspace` configures gopls for linux-primary editor analysis using a zig cross-toolchain:
  - `GOOS=linux`, `GOARCH=amd64`, `CGO_ENABLED=1`
  - `CC` set to `zig-cc.cmd` (not committed; wrapper: `zig.exe cc -target x86_64-linux-gnu -I<mesa>/drm -I<mesa>/drm/include -I<mesa>/drm/include/drm`; single-executable required by CGO on Windows)
  - `ZIG_LIB_DIR=C:/exe/zig-0.17.0/lib`
  - `CGO_CFLAGS=-I<mesa>/drm -I<mesa>/drm/include -I<mesa>/drm/include/drm` (three paths: root for `xf86drm.h`/`amdgpu/amdgpu.h`, `include/` for `drm/amdgpu_drm.h`, `include/drm/` for the flat `drm.h` that `xf86drm.h` includes)
  - `gopls.build.buildFlags=["-tags=cproto"]` (also `go.buildFlags` and `go.testTags`)
  - This combo lets gopls type-check `amdgpu_linux.go` end-to-end on a Windows host: zig provides Linux glibc headers for `runtime/cgo`; mesa/drm clone provides all libdrm and kernel DRM headers. `amdgpu_windows.go` shows "no packages found" when opened (unavoidable single-context limitation).
  - When focus shifts to Windows backend work, swap `GOOS` to `windows` and drop the `-tags=cproto` flag.
- `.cspell.jsonc` + `.markdownlint-cli2.jsonc` for spelling/linting

## Development roadmap

See `TODO.md` for phase-based plan (Phases 0-6). Key upcoming work:

- Phase 1.5: cross-platform refactor (file realignment) — **done**
- Phase 2.5: `DRMVersion()` method — **done**
- Phase 3: sensor reads (sysfs or `amdgpu_sensor_get_value`)
- Phase 4: atopweb integration (`amd/gpu.go` wrapper)
- Phase 6: real Windows backend (InpOut32 + SetupAPI + D3DKMT)

## Gotchas

- `DeviceInfo()` returns clocks in 10kHz units
- `MemoryInfo()` resizable_bar heuristic: `usable_heap_size < total_heap_size`
- `Device.handle` uses `unsafe.Pointer` — cast back to `C.amdgpu_device_handle` at every cgo call site in `amdgpu_linux.go`
- `go.sum` does not exist yet (no Go dependencies — only CGO system libs)
- `gomod2nix.toml` does not exist yet (Nix workflow generates it)
