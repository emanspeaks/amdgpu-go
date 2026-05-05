//go:build linux

package amdgpu

import (
	"fmt"
	"strings"
)

// buildDeviceInfoMap constructs the Info map for a GPU device from all available
// sources: DRM ioctl results, sysfs, and derived values.
func buildDeviceInfoMap(card int, di *DeviceInfo, mi *MemoryInfo, state *PollState) map[string]interface{} {
	meta := ChipMetaForFamily(di.Family)

	gpuType := "Discrete"
	if di.IsApu {
		gpuType = "APU"
	}

	devID, _ := readPCIDeviceID(card)
	vramType := readVRAMType(card)
	vramVendor := readVRAMVendor(card)

	var drmMajor, drmMinor, drmPatch int
	fmt.Sscanf(state.DRMVer, "%d.%d.%d", &drmMajor, &drmMinor, &drmPatch)

	deviceName := di.MarketingName
	if deviceName == "" {
		deviceName = ReadGPUName(card)
	}

	// ASIC Name first part is the gfx target version uppercased (e.g. "GFX1151"),
	// matching amdgpu_top's get_asic_name() output. ChipClass uses the enum form
	// with underscores (e.g. "GFX11_5") and is emitted separately.
	asicNameFirst := strings.ToUpper(meta.GFXTarget) // "gfx1151" → "GFX1151"

	base := renderDevPath(card)
	sclkMin, _ := readMinMaxClockMHz(base + "/pp_dpm_sclk")
	mclkMin, _ := readMinMaxClockMHz(base + "/pp_dpm_mclk")

	m := map[string]interface{}{
		"DeviceName":         deviceName,
		"ASIC Name":          asicNameFirst + "/" + meta.Name,
		"Chip Class":         meta.ChipClass,
		"DeviceID":           int(devID),
		"RevisionID":         int(di.ExternalRev),
		"PCI":                state.PCIDev,
		"DevicePath": map[string]interface{}{
			"DeviceID":   int(devID),
			"DeviceName": deviceName,
			"RevisionID": int(di.ExternalRev),
			"card":       fmt.Sprintf("/dev/dri/card%d", card),
			"pci":        state.PCIDev,
			"render":     fmt.Sprintf("/dev/dri/renderD%d", 128+card),
		},
		"GPU Family":         meta.GPUFamily,
		"GPU Type":           gpuType,
		"gfx_target_version": meta.GFXTarget,
		"GPU Clock": map[string]interface{}{
			"max": int(di.MaxEngineClock / 1000), // kHz → MHz
			"min": int(sclkMin),
		},
		"Memory Clock": map[string]interface{}{
			"max": int(di.MaxMemoryClock / 1000),
			"min": int(mclkMin),
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
	var totalCU uint32
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
		totalCU = di.NumShaderEngines * di.NumShaderArraysPerEngine * di.NumCUPerSH
		if totalCU > 0 {
			m["num_cu"] = int(totalCU)
			m["Total Compute Unit"] = int(totalCU)
			m["Compute Unit"] = int(totalCU) // frontend reads this key for CU display
		}
	}

	// Peak FP32: numCU × FLOPsPerCUPerMHz × MaxEngineClockMHz / 1000 = GFLOPS
	if totalCU > 0 && di.MaxEngineClock > 0 {
		maxClkMHz := float64(di.MaxEngineClock) / 1000
		peakGFLOPS := float64(totalCU) * float64(FLOPsPerCUPerMHz(state.Gen)) * maxClkMHz / 1000
		m["Peak FP32"] = map[string]interface{}{"unit": "GFLOPS", "value": peakGFLOPS}
	}

	// NPU: prefer the XDNA driver's vbnv string (e.g. "RyzenAI-npu5") over the
	// hardcoded chip lookup table name.
	npuName := meta.NPUName
	if state.Xdna.AccelDev != "" {
		if vbnv := readXDNADeviceName(state.Xdna.AccelDev); vbnv != "" {
			npuName = vbnv
		}
	}
	if npuName != "" {
		m["NPU Name"] = npuName
		m["NPU"] = npuName
	}

	// ROCm version.
	if rocm := ReadROCmVersion(); rocm != "" {
		m["ROCm Version"] = rocm
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
	// VRAM type: sysfs first, fall back to ioctl enum (sysfs is empty on some APUs).
	if vramType == "" {
		vramType = VRAMTypeName(di.VRAMType)
	}
	if vramType != "" {
		m["VRAM Type"] = vramType
	}
	if vramVendor != "" {
		m["VRAM Vendor"] = vramVendor
	}
	if di.VRAMBitWidth > 0 {
		m["VRAM Bit width"] = int(di.VRAMBitWidth)
	}

	// Render backends and derived metrics.
	if di.NumRBPipes > 0 {
		ropsPerRB := 4
		rbType := "RB"
		if RBPlusForGen(state.Gen) {
			ropsPerRB = 8
			rbType = "RB Plus"
		}
		totalROP := int(di.NumRBPipes) * ropsPerRB
		m["RenderBackend"] = int(di.NumRBPipes)
		m["RenderBackend Type"] = rbType
		m["Total ROP"] = totalROP

		if di.MaxEngineClock > 0 {
			maxClkGHz := float64(di.MaxEngineClock) / 1e6 // kHz → GHz
			m["Peak Pixel Fill-Rate"] = map[string]interface{}{
				"unit":  "GP/s",
				"value": float64(totalROP) * maxClkGHz,
			}
		}
	}

	// Peak memory bandwidth: MaxMemClkMHz × VRAMBitWidth × 2 (DDR) / 8 / 1000.
	// For LPDDR5 the kernel reports half the effective clock; multiply by 8 not 2
	// to match amdgpu_top's output. We detect this via vram_type == 12 (LPDDR5).
	if di.VRAMBitWidth > 0 && di.MaxMemoryClock > 0 {
		maxMemMHz := float64(di.MaxMemoryClock) / 1000 // kHz → MHz
		var factor float64
		if di.VRAMType == 12 { // LPDDR5
			factor = 8
		} else {
			factor = 2
		}
		bwGBs := maxMemMHz * float64(di.VRAMBitWidth) * factor / 8 / 1000
		m["Peak Memory Bandwidth"] = map[string]interface{}{"unit": "GB/s", "value": bwGBs}
	}

	// Static hardware lists cached at init.
	if len(state.FirmwareInfo) > 0 {
		m["Firmware info"] = state.FirmwareInfo
	}
	if len(state.HWIPList) > 0 {
		m["Hardware IP info"] = state.HWIPList
	}
	if state.VBIOSData != nil {
		m["VBIOS"] = map[string]interface{}{
			"name":    state.VBIOSData.Name,
			"pn":      state.VBIOSData.PN,
			"ver_str": state.VBIOSData.VerStr,
			"date":    state.VBIOSData.Date,
		}
	}
	if state.VideoCaps != nil {
		m["Video Caps"] = state.VideoCaps
	}
	if features := ReadPPFeatureMask(); features != nil {
		m["pp_feature_mask"] = features
	}

	return m
}
