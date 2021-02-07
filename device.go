package insteon

import (
	"context"
	"fmt"
)

type Address [3]byte

func (a Address) String() string {
	return fmt.Sprintf("%02X:%02X:%02X", a[0], a[1], a[2])
}

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
func (d *Device) TurnOnLevel(ctx context.Context, ramp bool, level byte) error {
	ctlCmd := cmdControlOn
	if !ramp {
		ctlCmd = cmdControlFastOn
	}

	_, err := d.hub.SendCommand(ctx, d.address, ctlCmd, level)

	return err
}

// TurnOn turns on a device.
func (d *Device) TurnOn(ctx context.Context) error {
	return d.TurnOnRamp(ctx, false)
}

// TurnOnRamp turns on a device with optional ramp. Ramp only works if the
// device is dimmable.
func (d *Device) TurnOnRamp(ctx context.Context, ramp bool) error {
	return d.TurnOnLevel(ctx, ramp, 0xFF)
}

// TurnOff turns off a device.
func (d *Device) TurnOff(ctx context.Context) error {
	return d.TurnOffRamp(ctx, false)
}

// TurnOffRamp turns off a device with optional ramp.
func (d *Device) TurnOffRamp(ctx context.Context, ramp bool) error {
	ctlCmd := cmdControlOff
	if !ramp {
		ctlCmd = cmdControlFastOff
	}

	_, err := d.hub.SendCommand(ctx, d.address, ctlCmd, 0)

	return err
}

func (d *Device) SetFanLevel(ctx context.Context, level byte) error {
	_, err := d.hub.SendExtendedCommand(ctx, d.address, cmdControlOn, level, [14]byte{2})

	return err
}

type DeviceIdentification struct {
	Category    byte
	SubCategory byte
	Firmware    byte
}

func (d *Device) Ping(ctx context.Context) error {
	_, err := d.hub.SendCommand(ctx, d.address, cmdControlPing, 0)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) GetProductData(ctx context.Context) error {
	data := [14]byte{}
	_, err := d.hub.SendExtendedCommand(ctx, d.address, cmdControlProduct, 2, data)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) GetDatabase(ctx context.Context) ([]*AllLinkRecord, error) {
	data := [14]byte{0x01, 0x00, 0x0f, 0xff}

	_, err := d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0, data)
	if err != nil {
		return nil, err
	}

	doneCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var db []*AllLinkRecord
	var lastErr error
	var lastAddr uint16

	listener := func(event Event, err error) {
		if err != nil {
			lastErr = err
			cancel()
		}

		if evt, ok := event.(*ExtCommandResponse); ok {
			addr := uint16(evt.Data[2])<<8 + uint16(evt.Data[3])
			dbEntry := &AllLinkRecord{}
			dbEntry.fromBytes(evt.Data[3:])

			if dbEntry.Flags.Last() {
				cancel()
			} else if addr != lastAddr {
				db = append(db, dbEntry)
			}

			lastAddr = addr
		}
	}

	d.hub.AddEventListener(listener)
	defer d.hub.RemoveEventListener(listener)

	select {
	case <-doneCtx.Done():
		return db, lastErr
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetStatus gets the current power status of the device.
func (d *Device) GetStatus(ctx context.Context) (*DeviceStatus, error) {
	return d.GetStatusChannel(ctx, 0)
}

// GetStatusChannel gets the current power status of the device.
func (d *Device) GetStatusChannel(ctx context.Context, channel byte) (*DeviceStatus, error) {
	rsp, err := d.hub.SendCommand(ctx, d.address, cmdQueryStatusRequest, channel)
	if err != nil {
		return nil, err
	}

	return &DeviceStatus{
		DeviceAddr: rsp.From,
		ModemAddr:  rsp.To,
		HopCount:   byte(rsp.Flags),
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
