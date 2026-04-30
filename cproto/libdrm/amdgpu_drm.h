/*
 * Editor-only stub for <libdrm/amdgpu_drm.h>.
 *
 * Used by gopls when analyzing this package on hosts without real libdrm.
 * Real builds use the system header via pkg-config.
 * Only declares the symbols and struct fields this package actually
 * references. Field types match the Go-side conversions; layouts do not
 * need to be ABI-correct because this header is never used in real builds.
 */
#ifndef AMDGPUGO_CPROTO_LIBDRM_AMDGPU_DRM_H
#define AMDGPUGO_CPROTO_LIBDRM_AMDGPU_DRM_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

#define DRM_AMDGPU_INFO 0x05

#define AMDGPU_INFO_VRAM_GTT 0x14
#define AMDGPU_INFO_GET_INFO 0x16
#define AMDGPU_INFO_MEMORY   0x19

struct drm_amdgpu_info {
    uint32_t query;
    uintptr_t return_pointer;
    uint32_t return_size;
};

struct drm_amdgpu_info_device {
    uint32_t asic_name;
    uint32_t chip_class;
    uint32_t is_apu;
    uint32_t max_engine_clock;
    uint32_t max_memory_clock;
};

struct drm_amdgpu_heap_info {
    uint64_t total_heap_size;
    uint64_t usable_heap_size;
    uint64_t heap_usage;
};

struct drm_amdgpu_memory_info {
    struct drm_amdgpu_heap_info vram;
    struct drm_amdgpu_heap_info gtt;
};

struct drm_amdgpu_info_vram_gtt {
    uint64_t vram_total;
    uint64_t vram_usable;
    uint64_t gtt_total;
    uint64_t gtt_usable;
};

#ifdef __cplusplus
}
#endif

#endif
