package insteon

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

var (
	ErrNotReady   = errors.New("device not ready")
	ErrAckTimeout = errors.New("ack timeout")
)

type ModemConfiguration byte

func (mc ModemConfiguration) AutoLink() bool {
	return !(mc&0x80 > 0)
}

func (mc ModemConfiguration) Monitor() bool {
	return mc&0x40 > 0
}

func (mc ModemConfiguration) AutoLED() bool {
	return !(mc&0x20 > 0)
}

func (mc ModemConfiguration) DeadMan() bool {
	return !(mc&0x10 > 0)
}

func (mc ModemConfiguration) String() string {
	return fmt.Sprintf("AutoLink=%t, Monitor=%t, AutoLED=%t, DeadMan=%t", mc.AutoLink(), mc.Monitor(), mc.AutoLED(), mc.DeadMan())
}

type Hub interface {
	SendCommand(ctx context.Context, addr Address, imCmd1 byte, imCmd2 byte) (*StdCommandResponse, error)
	SendExtendedCommand(ctx context.Context, addr Address, imCmd1, imCmd2 byte, userData [14]byte) (*StdCommandResponse, error)
	SendGroupCommand(ctx context.Context, hostCmd byte, group byte) error
	AddEventListener(EventListener)
	RemoveEventListener(EventListener)
	SetCommLogger(CommLogger)

	GetInfo() (*ModemInfo, error)
	GetModemConfig() (ModemConfiguration, error)
	SetModemConfig(cfg ModemConfiguration) error
	StartAllLink(LinkCode, byte) (*AllLinkCompleted, error)
	CancelAllLink() error
	GetAllLinkDatabase() ([]*AllLinkRecord, error)
	Beep() error
}

type EventListener func(Event, error)

type CommDirection int

const (
	CommDirectionHostToIM = iota
	CommDirectionIMToHost
	CommDirectionUnknown
)

func (cd CommDirection) String() string {
	switch cd {
	case CommDirectionIMToHost:
		return "IM to Host"
	case CommDirectionHostToIM:
		return "Host to IM"
	case CommDirectionUnknown:
		fallthrough
	default:
		return "Unknown"
	}
}

type CommLogger func(CommDirection, []byte)

// buildPlmCommand builds a power line modem command that gets sent by the
// Insteon hub.
func buildPlmCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte) []byte {
	return []byte{
		serialStart,
		hostCmd,
		addr[0], addr[1], addr[2],
		CommandFlagHopsLeftThree | CommandFlagRetransmitMax,
		imCmd1, imCmd2,
	}
}

func buildExtPlmCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte, userData [14]byte) []byte {
	return []byte{
		serialStart,
		hostCmd,
		addr[0], addr[1], addr[2],
		CommandFlagExtended | CommandFlagAck | CommandFlagHopsLeftThree | CommandFlagRetransmitMax,
		imCmd1, imCmd2,
		userData[0], userData[1], userData[2], userData[3], userData[4], userData[5], userData[6], userData[7],
		userData[8], userData[9], userData[10], userData[11], userData[12], userData[13],
	}
}

type ModemInfo struct {
	Address         Address
	Category        Category
	SubCategory     SubCategory
	FirmwareVersion byte
}

func (m *ModemInfo) fromBytes(b []byte) {
	m.Address = [3]byte{b[0], b[1], b[2]}
	m.Category = Category(b[3])
	m.SubCategory = SubCategory(b[4])
	m.FirmwareVersion = b[5]
}
