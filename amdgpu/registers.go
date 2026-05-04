package amdgpu

// ReadGRBM reads the GRBM_STATUS register and re
func (d *Device) ReadGRBM() (uint32, error) {
	values, err := d.ReadMMRegisters(GRBMOFFSET, 1)
	if err != nil {
		return 0, err
	}
	return values[0], nil
}

// ReadGRBM2 reads the GRBM2_STATUS2 register a
func (d *Device) ReadGRBM2() (uint32, error) {
	values, err := d.ReadMMRegisters(GRBM2OFFSET, 1)
	if err != nil {
		return 0, err
	}
	return values[0], nil
}
