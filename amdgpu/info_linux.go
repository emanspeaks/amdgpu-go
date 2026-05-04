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

	deviceName := di.MarketingName
	if deviceName == "" {
		deviceName = ReadGPUName(card)
	}

	m := map[string]interface{}{
		"DeviceName":         deviceName,
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

	// Shader topology — emit both new structured names and amdgpu_top compat names.
	if di.NumShaderEngines > 0 {
		m["NumShaderEngines"] = int(di.NumShaderEngines)
		m["Shader Engine"] = int(di.NumShaderEngines)
		m["NumShaderArraysPerEngine"] = int(di.NumShaderArraysPerEngine)
		m["Shader Array per Shader Engine"] = int(di.NumShaderArraysPerEngine)
		m["NumCUPerSH"] = int(di.NumCUPerSH)
		m["CU per Shader Array"] = map[string]interface{}{
			"max": int(di.NumCUPerSH),
			"min": int(di.NumCUPerSH),
		}
		totalCU := di.NumShaderEngines * di.NumShaderArraysPerEngine * di.NumCUPerSH
		if totalCU > 0 {
			m["num_cu"] = int(totalCU)
			m["Total Compute Unit"] = int(totalCU)
		}
	}

	if meta.NPUName != "" {
		m["NPU Name"] = meta.NPUName
		m["NPU"] = meta.NPUName
	}

	// Cache sizes — kernel reports GL0/GL1/GL2 in KiB; convert to bytes.
	if di.GL0CacheSize > 0 {
		b := int64(di.GL0CacheSize) * 1024
		m["GL0 Cache Size"] = b
		m["L1 Cache per CU"] = b
	}
	if di.GL1CacheSize > 0 {
		b := int64(di.GL1CacheSize) * 1024
		m["GL1 Cache Size"] = b
		m["GL1 Cache per Shader Array"] = b
	}
	if di.GL2CacheSize > 0 {
		b := int64(di.GL2CacheSize) * 1024
		m["GL2 Cache Size"] = b
		m["L2 Cache"] = b
	}
	if di.MallSize > 0 {
		m["L3 Cache Size"] = int64(di.MallSize)
		m["L3 Cache"] = int64(di.MallSize)
	}
	if vramType != "" {
		m["VRAM Type"] = vramType
	}
	if vramVendor != "" {
		m["VRAM Vendor"] = vramVendor
	}
	return m
}
