//go:build linux

package amdgpu

import "fmt"

// buildDeviceInfoMap constructs the Info map for a GPU device from all available
// sources: DRM ioctl results, sysfs, and derived values.
func buildDeviceInfoMap(card int, di *DeviceInfo, mi *MemoryInfo, pciDev, drmVersion string) map[string]interface{} {
	meta := ChipMetaForFamily(di.Family)

	gpuType := "Discrete"
	if di.IsApu {
		gpuType = "APU"
	}

	devID, _ := readPCIDeviceID(card)
	vramType := readVRAMType(card)
	vramVendor := readVRAMVendor(card)

	var drmMajor, drmMinor, drmPatch int
	fmt.Sscanf(drmVersion, "%d.%d.%d", &drmMajor, &drmMinor, &drmPatch)

	m := map[string]interface{}{
		"DeviceName":         ReadGPUName(card),
		"ASIC Name":          meta.ChipClass + "/" + meta.Name,
		"Chip Class":         meta.ChipClass,
		"DeviceID":           int(devID),
		"RevisionID":         int(di.ExternalRev),
		"PCI":                pciDev,
		"GPU Family":         meta.GPUFamily,
		"GPU Type":           gpuType,
		"gfx_target_version": meta.GFXTarget,
		"GPU Clock": map[string]interface{}{
			"max": int(di.MaxEngineClock / 1000), // kHz → MHz
			"min": 0,
		},
		"Memory Clock": map[string]interface{}{
			"max": int(di.MaxMemoryClock / 1000),
			"min": 0,
		},
		"VRAM Size":       int64(mi.VRAMTotalHeapSize),
		"VRAM Usage Size": int64(mi.VRAMHeapUsage),
		"GTT Size":        int64(mi.GTTTotalHeapSize),
		"GTT Usage Size":  int64(mi.GTTHeapUsage),
		"ResizableBAR":    mi.ResizableBar,
		"drm_version": map[string]interface{}{
			"major":      drmMajor,
			"minor":      drmMinor,
			"patchlevel": drmPatch,
		},
		"num_tcc_blocks": int(di.NUMTCCBlocks),
	}

	// Shader topology
	if di.NumShaderEngines > 0 {
		m["NumShaderEngines"] = int(di.NumShaderEngines)
		m["NumShaderArraysPerEngine"] = int(di.NumShaderArraysPerEngine)
		m["NumCUPerSH"] = int(di.NumCUPerSH)
		totalCU := di.NumShaderEngines * di.NumShaderArraysPerEngine * di.NumCUPerSH
		if totalCU > 0 {
			m["num_cu"] = int(totalCU)
		}
	}

	if meta.NPUName != "" {
		m["NPU Name"] = meta.NPUName
	}

	if di.GL0CacheSize > 0 {
		m["GL0 Cache Size"] = int64(di.GL0CacheSize)
	}
	if di.GL1CacheSize > 0 {
		m["GL1 Cache Size"] = int64(di.GL1CacheSize)
	}
	if di.GL2CacheSize > 0 {
		m["GL2 Cache Size"] = int64(di.GL2CacheSize)
	}
	if di.MallSize > 0 {
		m["L3 Cache Size"] = int64(di.MallSize)
	}
	if vramType != "" {
		m["VRAM Type"] = vramType
	}
	if vramVendor != "" {
		m["VRAM Vendor"] = vramVendor
	}
	return m
}
