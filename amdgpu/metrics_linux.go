//go:build linux

package amdgpu

import (
	"encoding/binary"
	"fmt"
	"os"
)

// ParseGPUMetrics reads the gpu_metrics binary sysfs file for the given card
// and returns a map suitable for JSON marshalling into the "gpu_metrics" field.
//
// Temperature values are in centi-Celsius (frontend divides by 100 for °C).
// Power values are in milliwatts (frontend divides by 1000 for watts).
// Clock frequencies are in MHz.
func ParseGPUMetrics(card int) map[string]interface{} {
	path := fmt.Sprintf("/sys/class/drm/renderD%d/device/gpu_metrics", 128+card)
	data, err := os.ReadFile(path)
	if err != nil || len(data) < 4 {
		return map[string]interface{}{}
	}
	formatRev := data[2]
	contentRev := data[3]
	switch formatRev {
	case 3:
		return parseMetricsV3(data, contentRev)
	case 2:
		return parseMetricsV2(data, contentRev)
	case 1:
		return parseMetricsV1(data, contentRev)
	}
	return map[string]interface{}{}
}

func gmu16(data []byte, off int) uint16 {
	if off+2 > len(data) {
		return 0
	}
	return binary.LittleEndian.Uint16(data[off:])
}

func gmu32(data []byte, off int) uint32 {
	if off+4 > len(data) {
		return 0
	}
	return binary.LittleEndian.Uint32(data[off:])
}

// parseMetricsV2 parses the APU gpu_metrics format (format_revision=2).
// All v2.x versions share the same base layout; newer content_revisions
// only append fields at the end.
// Verified from Linux kernel kgd_pp_interface.h struct gpu_metrics_v2_1.
func parseMetricsV2(data []byte, _ uint8) map[string]interface{} {
	if len(data) < 90 {
		return map[string]interface{}{}
	}
	m := make(map[string]interface{})

	m["temperature_gfx"] = gmu16(data, 4)
	m["temperature_soc"] = gmu16(data, 6)

	coreTemp := make([]uint16, 8)
	for j := range coreTemp {
		coreTemp[j] = gmu16(data, 8+j*2)
	}
	m["temperature_core"] = coreTemp

	m["average_gfx_activity"] = gmu16(data, 28)
	m["average_mm_activity"] = gmu16(data, 30)

	m["average_socket_power"] = gmu16(data, 40)
	m["average_all_core_power"] = gmu16(data, 42)

	corePwr := make([]uint16, 8)
	for j := range corePwr {
		corePwr[j] = gmu16(data, 48+j*2)
	}
	m["average_core_power"] = corePwr

	if f := gmu16(data, 64); f > 0 {
		m["average_gfxclk_frequency"] = f
	}
	if f := gmu16(data, 66); f > 0 {
		m["average_socclk_frequency"] = f
	}
	if f := gmu16(data, 68); f > 0 {
		m["average_uclk_frequency"] = f
	}
	if f := gmu16(data, 70); f > 0 {
		m["average_fclk_frequency"] = f
	}
	if f := gmu16(data, 72); f > 0 {
		m["average_vclk_frequency"] = f
	}

	if len(data) >= 104 {
		coreClk := make([]uint16, 8)
		for j := range coreClk {
			coreClk[j] = gmu16(data, 88+j*2)
		}
		m["current_coreclk"] = coreClk
	}
	return m
}

// parseMetricsV3 parses the latest APU gpu_metrics format (format_revision=3).
// Used on Phoenix/Hawk Point APUs.
// Verified from Linux kernel kgd_pp_interface.h struct gpu_metrics_v3_0.
func parseMetricsV3(data []byte, _ uint8) map[string]interface{} {
	if len(data) < 190 {
		return map[string]interface{}{}
	}
	m := make(map[string]interface{})

	m["temperature_gfx"] = gmu16(data, 4)
	m["temperature_soc"] = gmu16(data, 6)
	m["temperature_skin"] = gmu16(data, 40)

	coreTemp := make([]uint16, 16)
	for j := range coreTemp {
		coreTemp[j] = gmu16(data, 8+j*2)
	}
	m["temperature_core"] = coreTemp

	m["average_gfx_activity"] = gmu16(data, 42)
	m["average_vcn_activity"] = gmu16(data, 44)

	ipuAct := make([]uint16, 8)
	for j := range ipuAct {
		ipuAct[j] = gmu16(data, 46+j*2)
	}
	m["average_ipu_activity"] = ipuAct

	m["average_dram_reads"] = gmu16(data, 94)
	m["average_dram_writes"] = gmu16(data, 96)
	m["average_ipu_reads"] = gmu16(data, 98)
	m["average_ipu_writes"] = gmu16(data, 100)

	m["average_socket_power"] = gmu32(data, 112)
	m["average_ipu_power"] = gmu16(data, 116)
	m["average_apu_power"] = gmu32(data, 120)
	m["average_gfx_power"] = gmu32(data, 124)
	m["average_all_core_power"] = gmu32(data, 132)

	corePwr := make([]uint16, 16)
	for j := range corePwr {
		corePwr[j] = gmu16(data, 136+j*2)
	}
	m["average_core_power"] = corePwr

	m["average_sys_power"] = gmu16(data, 168)

	if f := gmu16(data, 174); f > 0 {
		m["average_gfxclk_frequency"] = f
	}
	if f := gmu16(data, 176); f > 0 {
		m["average_socclk_frequency"] = f
	}
	if f := gmu16(data, 178); f > 0 {
		m["average_vpeclk_frequency"] = f
	}
	if f := gmu16(data, 180); f > 0 {
		m["average_ipuclk_frequency"] = f
	}
	if f := gmu16(data, 182); f > 0 {
		m["average_fclk_frequency"] = f
	}
	if f := gmu16(data, 184); f > 0 {
		m["average_vclk_frequency"] = f
	}
	if f := gmu16(data, 186); f > 0 {
		m["average_uclk_frequency"] = f
	}
	if f := gmu16(data, 188); f > 0 {
		m["average_mpipu_frequency"] = f
	}

	if len(data) >= 222 {
		coreClk := make([]uint16, 16)
		for j := range coreClk {
			coreClk[j] = gmu16(data, 190+j*2)
		}
		m["current_coreclk"] = coreClk
	}
	return m
}

// parseMetricsV1 parses the discrete GPU gpu_metrics format (format_revision=1).
// Handles v1.1+ (v1.0 skipped due to alignment issues).
// Verified from Linux kernel kgd_pp_interface.h struct gpu_metrics_v1_1.
func parseMetricsV1(data []byte, contentRev uint8) map[string]interface{} {
	if contentRev == 0 || len(data) < 48 {
		return map[string]interface{}{}
	}
	m := make(map[string]interface{})

	m["temperature_gfx"] = gmu16(data, 4) // edge temp mapped to temperature_gfx
	m["temperature_hotspot"] = gmu16(data, 6)
	m["temperature_mem"] = gmu16(data, 8)

	m["average_gfx_activity"] = gmu16(data, 16)
	m["average_umc_activity"] = gmu16(data, 18)
	m["average_mm_activity"] = gmu16(data, 20)

	m["average_socket_power"] = gmu16(data, 22)

	if f := gmu16(data, 40); f > 0 {
		m["average_gfxclk_frequency"] = f
	}
	if f := gmu16(data, 42); f > 0 {
		m["average_socclk_frequency"] = f
	}
	if f := gmu16(data, 46); f > 0 {
		m["average_vclk_frequency"] = f
	}
	return m
}

// buildNPUMetrics extracts NPU (IPU) metrics from the gpu_metrics map,
// returning a map matching the npu_metrics shape in amdgpu_top JSON output.
// All scalar values are wrapped in {unit, value} objects; npu_busy is an array.
func buildNPUMetrics(gm map[string]interface{}) map[string]interface{} {
	busy := make([]uint16, 8)
	if arr, ok := gm["average_ipu_activity"].([]uint16); ok && len(arr) == 8 {
		copy(busy, arr)
	}
	getU16 := func(key string) uint16 {
		if v, ok := gm[key].(uint16); ok {
			return v
		}
		return 0
	}
	wrap := func(unit string, v interface{}) map[string]interface{} {
		return map[string]interface{}{"unit": unit, "value": v}
	}
	return map[string]interface{}{
		"npu_busy":      wrap("%", busy),
		"npu_power":     wrap("mW", getU16("average_ipu_power")),
		"npu_reads":     wrap("MB/s", getU16("average_ipu_reads")),
		"npu_writes":    wrap("MB/s", getU16("average_ipu_writes")),
		"npuclk_freq":   wrap("MHz", getU16("average_ipuclk_frequency")),
		"mpnpuclk_freq": wrap("MHz", getU16("average_mpipu_frequency")),
	}
}
