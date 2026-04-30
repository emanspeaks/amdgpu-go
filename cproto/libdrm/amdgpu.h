/*
 * Editor-only stub for <libdrm/amdgpu.h>.
 *
 * Used by gopls when analyzing this package on hosts without real libdrm.
 * Real builds use the system header via pkg-config.
 * Only declares the symbols this package actually references.
 */
#ifndef AMDGPUGO_CPROTO_LIBDRM_AMDGPU_H
#define AMDGPUGO_CPROTO_LIBDRM_AMDGPU_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct amdgpu_device *amdgpu_device_handle;

extern int amdgpu_device_initialize(int fd,
                                     uint32_t *major_version,
                                     uint32_t *minor_version,
                                     amdgpu_device_handle *device_handle);

extern void amdgpu_device_deinitialize(amdgpu_device_handle device_handle);

extern int amdgpu_read_mm_registers(amdgpu_device_handle device_handle,
                                     unsigned dword_offset,
                                     unsigned count,
                                     uint32_t instance,
                                     uint32_t flags,
                                     uint32_t *values);

#ifdef __cplusplus
}
#endif

#endif
