package insteon

import "fmt"

type Event interface {
	fromBytes([]byte)
	ID() byte
	Length() int
}

type Ack struct {
	Response []byte
	Type     byte
}

func (a *Ack) fromBytes(buffer []byte) {
	a.Response = buffer
	a.Type = buffer[len(buffer)-1]
}

func (a *Ack) ID() byte {
	return a.Type
}

func (a *Ack) Length() int {
	return len(a.Response)
}

type StdCommandResponse struct {
	From  Address
	To    Address
	Flags CommandResponseFlags
	Cmd1  byte
	Cmd2  byte
}

func (cr *StdCommandResponse) fromBytes(buffer []byte) {
	copy(cr.From[:], buffer[2:5])
	copy(cr.To[:], buffer[5:8])
	cr.Flags = CommandResponseFlags(buffer[8])
	cr.Cmd1 = buffer[9]
	cr.Cmd2 = buffer[10]
}

func (cr *StdCommandResponse) ID() byte {
	return cmdIMStd
}

func (cr *StdCommandResponse) Length() int {
	return 11
}

type ExtCommandResponse struct {
	From  Address
	To    Address
	Flags CommandResponseFlags
	Cmd1  byte
	Cmd2  byte
	Data  [14]byte
}

func (cr *ExtCommandResponse) fromBytes(buffer []byte) {
	copy(cr.From[:], buffer[2:5])
	copy(cr.To[:], buffer[5:8])
	cr.Flags = CommandResponseFlags(buffer[8])
	cr.Cmd1 = buffer[9]
	cr.Cmd2 = buffer[10]
	copy(cr.Data[:], buffer[11:24])
}

func (cr *ExtCommandResponse) ID() byte {
	return cmdIMExt
}

func (cr *ExtCommandResponse) Length() int {
	return 25
}

type X10Response struct {
	raw   X10Raw
	flags X10Flags
}

func (cr *X10Response) fromBytes(buffer []byte) {
	cr.raw = X10Raw(buffer[2])
	cr.flags = X10Flags(buffer[3])
}

func (cr *X10Response) ID() byte {
	return cmdIMX10
}

func (cr *X10Response) Length() int {
	return 4
}

type AllLinkCompleted struct {
	LinkCode    LinkCode
	Group       byte
	Address     Address
	Category    Category
	SubCategory SubCategory
	Firmware    byte
}

func (cr *AllLinkCompleted) fromBytes(buffer []byte) {
	cr.LinkCode = LinkCode(buffer[2])
	cr.Group = buffer[3]
	copy(cr.Address[:], buffer[4:7])
	cr.Category = Category(buffer[7])
	cr.SubCategory = SubCategory(buffer[8])
	cr.Firmware = buffer[9]
}

func (cr *AllLinkCompleted) ID() byte {
	return cmdIMAllLinkComplete
}

func (cr *AllLinkCompleted) Length() int {
	return 10
}

type ButtonEvent struct {
	Event ButtonEventType
}

func (cr *ButtonEvent) fromBytes(buffer []byte) {
	cr.Event = ButtonEventType(buffer[2])
}

func (cr *ButtonEvent) ID() byte {
	return cmdIMBtnEvt
}

func (cr *ButtonEvent) Length() int {
	return 3
}

type UserReset struct{}

func (cr *UserReset) fromBytes([]byte) {
}

func (cr *UserReset) ID() byte {
	return cmdIMRst
}

func (cr *UserReset) Length() int {
	return 2
}

type AllLinkCleanupFailure struct {
	Group   byte
	Address Address
}

func (cr *AllLinkCleanupFailure) fromBytes(buffer []byte) {
	cr.Group = buffer[3]
	copy(cr.Address[:], buffer[4:7])
}

func (cr *AllLinkCleanupFailure) ID() byte {
	return cmdIMAllLinkCleanFail
}

func (cr *AllLinkCleanupFailure) Length() int {
	return 7
}

type AllLinkRecord struct {
	Flags   AllLinkRecordFlags
	Group   byte
	Address Address
	Data    [3]byte
}

func (cr *AllLinkRecord) fromBytes(buffer []byte) {
	cr.Flags = AllLinkRecordFlags(buffer[2])
	cr.Group = buffer[3]
	copy(cr.Address[:], buffer[4:7])
	copy(cr.Data[:], buffer[7:10])
}

func (cr *AllLinkRecord) toBytes() []byte {
	return []byte{byte(cr.Flags), cr.Group, cr.Address[0], cr.Address[1], cr.Address[2], cr.Data[0], cr.Data[1], cr.Data[2]}
}

func (cr *AllLinkRecord) ID() byte {
	return cmdIMAllLinkRecord
}

func (cr *AllLinkRecord) Length() int {
	return 10
}

type AllLinkCleanup struct {
	Status AllLinkCleanupStatus
}

func (cr *AllLinkCleanup) fromBytes(buffer []byte) {
	cr.Status = AllLinkCleanupStatus(buffer[2])
}

func (cr *AllLinkCleanup) ID() byte {
	return cmdIMAllLinkCleanup
}

func (cr *AllLinkCleanup) Length() int {
	return 3
}

type DatabaseRecord struct {
	MemAddr uint16
	Record  *AllLinkRecord
}

func (cr *DatabaseRecord) fromBytes(buffer []byte) {
	cr.MemAddr = uint16(buffer[2])<<8 + uint16(buffer[3])
	cr.Record = &AllLinkRecord{}
	cr.Record.fromBytes(buffer[2:])
}

func (cr *DatabaseRecord) ID() byte {
	return cmdIMDatabaseRecord
}

func (cr *DatabaseRecord) Length() int {
	return 12
}

type CommandResponseFlags byte

func (flags CommandResponseFlags) BroadcastNAK() bool {
	return flags&0x80 > 0
}

func (flags CommandResponseFlags) AllLink() bool {
	return flags&0x40 > 0
}

func (flags CommandResponseFlags) Acknowledgement() bool {
	return flags&0x20 > 0
}

func (flags CommandResponseFlags) Extended() bool {
	return flags&0x10 > 0
}

func (flags CommandResponseFlags) HopsLeft() int {
	return int((flags & 0xC) >> 2)
}

func (flags CommandResponseFlags) MaxHops() int {
	return int(flags & 0x3)
}

func (flags CommandResponseFlags) String() string {
	return fmt.Sprintf("BroadcastNAK=%t, AllLink=%t, Acknowledgement=%t, Extended=%t, HopsLeft=%d, MaxHops=%d",
		flags.BroadcastNAK(), flags.AllLink(), flags.Acknowledgement(), flags.Extended(), flags.HopsLeft(), flags.MaxHops())
}
