package insteon

type Address [3]byte

// Device represents an Insteon device.
type Device struct {
	address Address
	hub     Hub
}

// NewDevice creates a new device by raw address.
func NewDevice(hub Hub, addr Address) (*Device, error) {
	return &Device{
		address: addr,
		hub:     hub,
	}, nil
}

// TurnOnLevel turns on a dimmable device set to a specific level.
func (d *Device) TurnOnLevel(ramp bool, level byte) error {
	ctlCmd := cmdControlOn
	if !ramp {
		ctlCmd = cmdControlFastOn
	}
	_, err := d.hub.SendCommand(cmdHostSendMsg, d.address, ctlCmd, level)
	return err
}

// TurnOn turns on a device.
func (d *Device) TurnOn() error {
	return d.TurnOnRamp(false)
}

// TurnOnRamp turns on a device with optional ramp. Ramp only works if the
// device is dimmable.
func (d *Device) TurnOnRamp(ramp bool) error {
	return d.TurnOnLevel(ramp, 0xFF)
}

// TurnOff turns off a device.
func (d *Device) TurnOff() error {
	return d.TurnOffRamp(false)
}

// TurnOffRamp turns off a device with optional ramp.
func (d *Device) TurnOffRamp(ramp bool) error {
	ctlCmd := cmdControlOff
	if !ramp {
		ctlCmd = cmdControlFastOff
	}
	_, err := d.hub.SendCommand(cmdHostSendMsg, d.address, ctlCmd, 0)
	return err
}

// Beep causes the device to beep.
func (d *Device) Beep() error {
	_, err := d.hub.SendCommand(cmdHostBeep, d.address, 0, 0)
	return err
}

// LED turns the device LED on or off.
func (d *Device) LED(on bool) error {
	ctlCmd := cmdHostLEDOn
	if !on {
		ctlCmd = cmdHostLEDOff
	}
	_, err := d.hub.SendCommand(cmdHostSendMsg, d.address, ctlCmd, 0)
	return err
}

// GetStatus gets the current power status of the device.
func (d *Device) GetStatus() (*DeviceStatus, error) {
	err := d.hub.ClearBuffer()
	if err != nil {
		return nil, err
	}

	rsp, err := d.hub.SendCommand(cmdHostSendMsg, d.address, cmdQueryStatusRequest, 0)
	if err != nil {
		return nil, err
	}

	return &DeviceStatus{
		DeviceAddr: rsp.From,
		ModemAddr:  rsp.To,
		HopCount:   rsp.Flags,
		Delta:      rsp.Cmd1,
		Level:      rsp.Cmd2,
	}, nil
}

// DeviceStatus represents the status of a device.
type DeviceStatus struct {
	// DeviceAddr is the address of the device.
	DeviceAddr Address
	// ModemAddr is the address of the hub.
	ModemAddr Address
	// Hop count to the device.
	HopCount byte
	// Delta of the device.
	Delta byte
	// Current power level of the device.
	Level byte
}
