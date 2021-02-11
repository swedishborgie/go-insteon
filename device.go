package insteon

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrDBEntryNotFound      = errors.New("unable to find database entry")
	ErrDBEntryAlreadyExists = errors.New("database entry for this device already exists")
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
	const maxNameLength = 14
	if len(name) > maxNameLength {
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

	if _, err := d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0, data); err != nil {
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

func (d *Device) DeleteAllLink(ctx context.Context, addr Address, group byte, controller bool) error {
	db, err := d.GetDatabase(ctx)
	if err != nil {
		return err
	}

	memAddr, lastAddr := findAllLinkDBEntry(db, addr, group, controller)
	if memAddr == 0 {
		return errors.Wrapf(ErrDBEntryNotFound, "address: %s, group: %d, controller: %t", addr, group, controller)
	}

	secondLastAddr := uint16(0)

	for a := range db {
		if a > memAddr && a < lastAddr {
			secondLastAddr = a
		}
	}

	log.Printf("memAddr=%x, lastAddr=%x, secondLast=%x", memAddr, lastAddr, secondLastAddr)

	// If this isn't the last entry in the list.
	if memAddr != lastAddr {
		// We need to swap the entry we want to delete with the current last entry.
		keepEntry := db[lastAddr]

		keepFlags := keepEntry.Flags
		if secondLastAddr == memAddr {
			keepFlags |= AllLinkRecordFlagsLast
		} else {
			keepFlags &= 0xFD
		}

		log.Printf("performing swap: %x <-> %x", memAddr, lastAddr)

		if _, err = d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0,
			d.modifyDbCommand(memAddr, keepFlags, keepEntry.Group, keepEntry.Address, keepEntry.Data)); err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond)
	}

	// Mark the last entry as empty.
	log.Printf("marking empty: %x", lastAddr)

	if _, err = d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0,
		d.modifyDbCommand(lastAddr, 0, 0, [3]byte{}, [3]byte{})); err != nil {
		return err
	}

	// If we haven't already done this,
	if secondLastAddr != memAddr && secondLastAddr > 0 {
		time.Sleep(500 * time.Millisecond)

		// We need to re-write the new last entry to toggle the last flag.
		newLast := db[secondLastAddr]
		log.Printf("marking last: %x", secondLastAddr)

		if _, err = d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0,
			d.modifyDbCommand(secondLastAddr, newLast.Flags|AllLinkRecordFlagsLast,
				newLast.Group, newLast.Address, newLast.Data)); err != nil {
			return err
		}
	}

	return nil
}

func (d *Device) UpdateAllLink(ctx context.Context, addr Address, group byte, data [3]byte, controller bool) error {
	db, err := d.GetDatabase(ctx)
	if err != nil {
		return err
	}

	memAddr, lastAddr := findAllLinkDBEntry(db, addr, group, controller)
	if memAddr == 0 {
		return errors.Wrapf(ErrDBEntryNotFound, "address: %s, group: %d, controller: %t", addr, group, controller)
	}

	flags := AllLinkRecordFlagsInUse
	if memAddr == lastAddr {
		flags |= AllLinkRecordFlagsLast
	}

	if controller {
		flags |= AllLinkRecordFlagsContoller
	}

	_, err = d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0,
		d.modifyDbCommand(memAddr, flags, group, addr, data))

	return err
}

func (d *Device) AddAllLink(ctx context.Context, addr Address, group byte, data [3]byte, controller bool) error {
	db, err := d.GetDatabase(ctx)
	if err != nil {
		return err
	}

	memAddr, lastAddr := findAllLinkDBEntry(db, addr, group, controller)
	if memAddr != 0 {
		return errors.Wrapf(ErrDBEntryAlreadyExists, "address: %s, group: %d, controller: %t", addr, group, controller)
	}

	// We need to do two things here, we need to add our new entry and then update the previous last entry to make sure
	// the last entry flag is cleared.
	memAddr = lastAddr + 0x8

	flags := AllLinkRecordFlagsInUse | AllLinkRecordFlagsLast
	if controller {
		flags |= AllLinkRecordFlagsContoller
	}

	// Create new last entry.
	if _, err = d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0,
		d.modifyDbCommand(memAddr, flags, group, addr, data)); err != nil {
		return err
	}

	oldLast := db[lastAddr]
	if oldLast.Flags&AllLinkRecordFlagsLast > 0 {
		// We need to clear the last flag from the previous last entry.
		if _, err = d.hub.SendExtendedCommand(ctx, d.address, cmdControlAllLink, 0,
			d.modifyDbCommand(lastAddr, oldLast.Flags&0xFD, oldLast.Group, oldLast.Address, oldLast.Data)); err != nil {
			return err
		}
	}

	return nil
}

func findAllLinkDBEntry(db map[uint16]*AllLinkRecord, addr Address, group byte, controller bool) (uint16, uint16) {
	var rByte AllLinkRecordFlags
	if controller {
		rByte = AllLinkRecordFlagsContoller
	}

	foundAddr := uint16(0)

	lastAddr := uint16(0xFFFF)

	for memAddr, d := range db {
		if bytes.Equal(d.Address[:], addr[:]) && d.Flags&AllLinkRecordFlagsContoller == rByte && d.Group == group {
			foundAddr = memAddr
		}

		if memAddr < lastAddr {
			lastAddr = memAddr
		}
	}

	return foundAddr, lastAddr
}

func (d *Device) modifyDbCommand(memAddr uint16, flags AllLinkRecordFlags, group byte, addr Address, data [3]byte) [14]byte {
	cmd := [14]byte{
		0,                      // Unused
		0x2,                    // Modify
		byte(memAddr>>8) & 0xF, // Address High Nibble
		byte(memAddr),          // Address Low Byte
		0,                      // Unused
		byte(flags),            // Flags
		group,                  // Group
		addr[0],                // Address
		addr[1],                //
		addr[2],                //
		data[0],                // Data
		data[1],                //
		data[2],                //
		0,                      // CRC
	}

	cmd[13] = calculateCRC(append([]byte{cmdControlAllLink, 0}, cmd[:]...))

	return cmd
}

func calculateCRC(buf []byte) byte {
	ckSum := uint(1)

	for _, c := range buf {
		ckSum += uint(c)
	}

	return ^byte(ckSum)
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
