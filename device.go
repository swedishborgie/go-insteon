package insteon

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pkg/errors"
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

func (d *Device) Address() Address {
	return d.address
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

func (d *Device) GetProductData(ctx context.Context) (*Product, error) {
	_, err := d.hub.SendCommand(ctx, d.address, cmdControlProduct, 0)
	if err != nil {
		return nil, err
	}

	evt, err := d.hub.Expect(ctx, &ExtCommandResponse{})
	if err != nil {
		return nil, err
	}

	data := evt.(*ExtCommandResponse).Data()
	prd := &Product{}
	prd.ProductKey = uint(data[1])<<16 + uint(data[2])<<8 + uint(data[3])
	prd.Category = Category(data[4])
	prd.SubCategory = SubCategory(data[5])
	prd.Description = GetProductDesc(prd.Category, prd.SubCategory)

	return prd, nil
}

func (d *Device) GetName(ctx context.Context) (string, error) {
	_, err := d.hub.SendCommand(ctx, d.address, cmdControlProduct, 2)
	if err != nil {
		return "", err
	}

	evt, err := d.hub.Expect(ctx, &ExtCommandResponse{})
	if err != nil {
		return "", err
	}

	data := string(evt.(*ExtCommandResponse).Data())

	return strings.Trim(data, "\x00"), nil
}

func (d *Device) SetName(ctx context.Context, name string) error {
	if len(name) > 14 {
		name = name[0:14]
	}

	data := [14]byte{}

	for idx := 0; idx < len(name); idx++ {
		data[idx] = name[idx]
	}

	_, err := d.hub.SendExtendedCommand(ctx, d.address, cmdControlProduct, 2, data)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) GetDatabase(ctx context.Context) (map[uint16]*AllLinkRecord, error) {
	data := [14]byte{}
	_, err := d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0, data)
	if err != nil {
		return nil, err
	}

	db := make(map[uint16]*AllLinkRecord)

	for {
		rsp, err := d.hub.Expect(ctx, &ExtCommandResponse{})
		if err != nil {
			return nil, err
		}

		evt := rsp.(*ExtCommandResponse)
		addr := uint16(evt.data[2])<<8 + uint16(evt.data[3])
		dbEntry := &AllLinkRecord{}
		dbEntry.fromBytes(evt.data[3:])

		if dbEntry.Flags.Last() {
			return db, nil
		}

		db[addr] = dbEntry
	}
}

// GetStatus gets the current power status of the device.
func (d *Device) GetStatus(ctx context.Context) (*DeviceStatus, error) {
	return d.GetStatusChannel(ctx, 0)
}

func (d *Device) StartAllLink(ctx context.Context, group byte) error {
	_, err := d.hub.SendCommand(ctx, d.address, cmdControlLink, group)
	if err != nil {
		return err
	}

	return nil
}

func (d *Device) ModifyAllLinkEntry(ctx context.Context, alCmd ManageAllLinkCommand, flags AllLinkRecordFlags, group byte, addr Address, data [3]byte) error {
	db, err := d.GetDatabase(ctx)
	if err != nil {
		return err
	}

	var foundEntry *AllLinkRecord
	var foundAddr uint16

	for memAddr, db := range db {
		if bytes.Equal(db.Address[:], addr[:]) {
			foundEntry = db
			foundAddr = memAddr
		}
	}

	log.Printf("found device %+v at %d", foundEntry, foundAddr)

	switch alCmd {
	case ManageAllLinkAddResponder:
	case ManageAllLinkAddController:
	case ManageAllLinkDelete:
	case ManageAllLinkUpdate:

	case ManageAllLinkFindFirst:
		fallthrough
	case ManageAllLinkFindNext:
		fallthrough
	default:
		return errors.New("unsupported modification type")
	}

	return nil
}

type DeviceOpFlags byte

func (f DeviceOpFlags) ProgramLock() bool {
	return f&0x1 > 0
}

func (f DeviceOpFlags) LEDTransmit() bool {
	return f&0x2 > 0
}

func (f DeviceOpFlags) ResumeDim() bool {
	return f&0x4 > 0
}

func (f DeviceOpFlags) LED() bool {
	return f&0x10 == 0
}

func (f DeviceOpFlags) LoadSense() bool {
	return f&0x20 == 0
}

func (f DeviceOpFlags) String() string {
	return fmt.Sprintf("ProgramLock=%t, LEDTransmit=%t, ResumeDim=%t, LED=%t, LoadSense=%t", f.ProgramLock(), f.LEDTransmit(), f.ResumeDim(), f.LED(), f.LoadSense())
}

func (d *Device) GetOperatingFlags(ctx context.Context) (DeviceOpFlags, error) {
	rsp, err := d.hub.SendCommand(ctx, d.address, cmdControlGetOpFlags, 0)
	if err != nil {
		return 0, err
	}

	return DeviceOpFlags(rsp.Cmd2()), nil
}

// GetStatusChannel gets the current power status of the device.
func (d *Device) GetStatusChannel(ctx context.Context, channel byte) (*DeviceStatus, error) {
	rsp, err := d.hub.SendCommand(ctx, d.address, cmdQueryStatusRequest, channel)
	if err != nil {
		return nil, err
	}

	return &DeviceStatus{
		DeviceAddr: rsp.From(),
		ModemAddr:  rsp.To(),
		HopCount:   byte(rsp.Flags()),
		Delta:      rsp.Cmd1(),
		Level:      rsp.Cmd2(),
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
