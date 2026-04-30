/*
 * Editor-only stub for <xf86drm.h>.
 *
 * Used by gopls when analyzing this package on hosts without real libdrm
 * (e.g. Windows). Real builds use the system header via pkg-config.
 * Only declares the symbols this package actually references.
 */
#ifndef AMDGPUGO_CPROTO_XF86DRM_H
#define AMDGPUGO_CPROTO_XF86DRM_H

#ifdef __cplusplus
extern "C" {
#endif

extern int drmCommandWriteRead(int fd, unsigned long drmCommandIndex,
                                void *data, unsigned long size);

typedef struct drm_version {
    int   version_major;
    int   version_minor;
    int   version_patchlevel;
    int   name_len;
    char *name;
    int   date_len;
    char *date;
    int   desc_len;
    char *desc;
} drmVersion, *drmVersionPtr;

extern drmVersionPtr drmGetVersion(int fd);
extern void          drmFreeVersion(drmVersionPtr v);

#ifdef __cplusplus
}
#endif

#endif
