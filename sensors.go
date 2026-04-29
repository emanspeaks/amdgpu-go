package amdgpu

/*
#include <amdgpu.h>
#include <stdint.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// SensorValue reads a sensor value via amdgpu_sensor_get_value.
// Returns the raw sensor value (units depend on sensor_type).
//
// Common sensor types:
//   - SENSOR_TYPE_GFX_SCLK: GPU engine clock in kHz
//   - SENSOR_TYPE_GFX_MCLK: GPU memory clock in kHz
//   - SENSOR_TYPE_EDGE_TEMPERATURE: Die temperature in millidegrees C
//   - SENSOR_TYPE_VDDGFX: GFX voltage in millivolts
//   - SENSOR_TYPE_VDDNB: NB voltage in millivolts
func (d *Device) SensorValue(stype SENSOR_TYPE) (uint32, error) {
	var value C.uint32_t
	ret := C.sensor_wrapper(d.handle, C.uint32_t(stype), &value)
	if ret < 0 {
		return 0, fmt.Errorf("sensor query %d: %w", stype, mapErr(ret))
	}
	return uint32(value), nil
}
