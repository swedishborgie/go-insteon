package insteon

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrNotReady   = errors.New("device not ready")
	ErrAckTimeout = errors.New("ack timeout")
)

type CommandResponse struct {
	From  Address
	To    Address
	Flags CommandResponseFlags
	Cmd1  byte
	Cmd2  byte
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

func (cr *CommandResponse) fromBytes(buffer []byte) {
	copy(cr.From[:], buffer[2:5])
	copy(cr.To[:], buffer[5:8])
	cr.Flags = CommandResponseFlags(buffer[8])
	cr.Cmd1 = buffer[9]
	cr.Cmd2 = buffer[10]
}

type Hub interface {
	SendCommand(hostCmd byte, addr Address, imCmd1 byte, imCmd2 byte) (*CommandResponse, error)
	SendExtendedCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte, userData [14]byte) (*CommandResponse, error)
	SendGroupCommand(hostCmd byte, group byte) error
	GetBuffer() ([]byte, error)
	ClearBuffer() error
	waitForAck(cmd []byte) error
	waitForResponse() (*CommandResponse, error)
}

// buildPlmCommand builds a power line modem command that gets sent by the
// Insteon hub.
func buildPlmCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte) []byte {
	return []byte{
		serialStart,
		hostCmd,
		addr[0], addr[1], addr[2],
		0x0f,
		imCmd1, imCmd2,
	}
}

func buildExtPlmCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte, userData [14]byte) []byte {
	return []byte{
		serialStart,
		hostCmd,
		addr[0], addr[1], addr[2],
		0x15,
		imCmd1, imCmd2,
		userData[0], userData[1], userData[2], userData[3], userData[4], userData[5], userData[6], userData[7],
		userData[8], userData[9], userData[10], userData[11], userData[12], userData[13],
	}
}
