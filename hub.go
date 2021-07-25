package insteon

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

var (
	// ErrNotReady indicates the Hub received the command but isn't in a state where it's capable of processing it.
	ErrNotReady = errors.New("device not ready")
	// ErrAckTimeout indicates we didn't receive an acknowledgement in an appropriate amount of time.
	ErrAckTimeout = errors.New("ack timeout")
)

// Hub is an interface that represents functionality that can be performed by any of the Insteon Hub's.
// A generic implementation is provided by HubStreaming.
// The underlying serial communication protocol is described in these documents:
// https://cache.insteon.com/pdf/INSTEON_Modem_Developer%27s_Guide_20071012a.pdf
// http://cache.insteon.com/developer/2242-222dev-062013-en.pdf
type Hub interface {
	// SendMessage sends a standard length message to a remote device connected to the same network as this Hub.
	SendMessage(ctx context.Context, addr Address, imCmd1 byte, imCmd2 byte) (CommandResponse, error)
	// SendExtendedMessage sends an extended length message to a remote device connected to the same network as this
	// Hub.
	SendExtendedMessage(ctx context.Context, addr Address, imCmd1, imCmd2 byte, userData [14]byte) (CommandResponse, error)
	// Expect indicates that you're interested in waiting for a particular type of event from the Hub. The first event
	// with a matching ID will be returned.
	Expect(ctx context.Context, evt Event) (Event, error)
	// SendX10 sends an X10 message to the network this Hub is connected to.
	SendX10(context.Context, X10Raw, X10Flags) error
	// SendGroupCommand sends a group command to the network this Hub is connected to.
	SendGroupCommand(ctx context.Context, hostCmd byte, group byte) error
	// AddEventListener registers a listener interested in events coming from this Hub.
	AddEventListener(EventListener)
	// RemoveEventListener removes a previously registered EventListener.
	RemoveEventListener(EventListener)
	// SetCommLogger can be used to intercept the underlying serial communications between the host and this Hub.
	SetCommLogger(CommLogger)

	// GetInfo gets information about the Hub we're currently communicating with such as the Address, Category,
	// SubCategory, and firmware version.
	GetInfo(context.Context) (*ModemInfo, error)
	// SetDeviceCategory changes the Category, SubCategory, and Firmware Version of this Hub.
	SetDeviceCategory(context.Context, Category, SubCategory, byte) error
	// Sleep tells the Hub to go into low power mode.
	Sleep(context.Context) error
	// Reset clears the All-Link database, Modem Configuration, and Hub Information.
	Reset(context.Context) error
	// GetModemConfig gets the current modem configuration for the Hub we're communicating with.
	GetModemConfig(context.Context) (ModemConfiguration, error)
	// SetModemConfig sets the modem configuration for the current Hub.
	SetModemConfig(context.Context, ModemConfiguration) error
	// StartAllLink puts the modem into All-Link mode without requiring the user to press the button on the Hub.
	StartAllLink(context.Context, LinkCode, byte) (*AllLinkCompleted, error)
	// CancelAllLink removes the modem from All-Link mode.
	CancelAllLink(context.Context) error
	// ModifyAllLinkEntry modifies a single entry in the Hub's All-Link database.
	ModifyAllLinkEntry(context.Context, ManageAllLinkCommand, AllLinkRecordFlags, byte, Address, [3]byte) error
	// GetAllLinkDatabase retrieves all entries from the Hub's All-Link database.
	GetAllLinkDatabase(context.Context) ([]*AllLinkRecord, error)
	// GetLastSender gets the All-Link record of the last Insteon device to communicate with the Hub.
	GetLastSender(context.Context) (*AllLinkRecord, error)
	// Beep makes the Hub audibly beep.
	Beep(context.Context) error
	// ReadDB performs a raw read of the Hub's All-Link database.
	ReadDB(context.Context, uint16) (*DatabaseRecord, error)
	// WriteDB performs a raw write of the Hub's All-Link database.
	WriteDB(context.Context, uint16, *AllLinkRecord) error
	// SetLED sets the status of the Hub's LED.
	SetLED(context.Context, bool) error
}

// ModemConfiguration is a bitfield describing the current configuration of a PLM.
type ModemConfiguration byte

const (
	// ModemConfigurationAutoLink enables automatic linking when the user pushes and holds the SET Button.
	ModemConfigurationAutoLink ModemConfiguration = 0x80
	// ModemConfigurationMonitor puts the modem into Monitor Mode. Monitor mode allows the modem to receive communication
	// from any device in the Modem's All-Link database regardless of whether or not the message is addressed to the
	// modem.
	ModemConfigurationMonitor ModemConfiguration = 0x40
	// ModemConfigurationAutoLED enables automatic LED operation. This must be off to use Hub.SetLED().
	ModemConfigurationAutoLED ModemConfiguration = 0x20
	// ModemConfigurationDeadMan enables dead-man detection for communication between the modem and ourselves. Dead-man
	// detection will automatically reset communication if commands aren't sent in 240ms.
	ModemConfigurationDeadMan ModemConfiguration = 0x10
)

// AutoLink indicates of the ModemConfigurationAutoLink flag is set on this ModemConfiguration.
func (mc ModemConfiguration) AutoLink() bool {
	return !(mc&ModemConfigurationAutoLink > 0)
}

// Monitor indicates of the ModemConfigurationMonitor flag is set on this ModemConfiguration.
func (mc ModemConfiguration) Monitor() bool {
	return mc&ModemConfigurationMonitor > 0
}

// AutoLED indicates of the ModemConfigurationAutoLED flag is set on this ModemConfiguration.
func (mc ModemConfiguration) AutoLED() bool {
	return !(mc&ModemConfigurationAutoLED > 0)
}

// DeadMan indicates of the ModemConfigurationDeadMan flag is set on this ModemConfiguration.
func (mc ModemConfiguration) DeadMan() bool {
	return !(mc&ModemConfigurationDeadMan > 0)
}

// String describes the state of the modem configuration in a human readable format.
func (mc ModemConfiguration) String() string {
	return fmt.Sprintf("AutoLink=%t, Monitor=%t, AutoLED=%t, DeadMan=%t",
		mc.AutoLink(), mc.Monitor(), mc.AutoLED(), mc.DeadMan())
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

// buildPlmCommand builds a power line modem command to send a remote device that gets sent by the Insteon hub.
func buildPlmCommand(addr Address, imCmd1, imCmd2 byte) []byte {
	return []byte{
		serialStart,
		cmdHostSendMsg,
		addr[0], addr[1], addr[2],
		CommandFlagHopsLeftThree | CommandFlagRetransmitMax,
		imCmd1, imCmd2,
	}
}

// buildGroupPlmCommand builds a group command that can be sent to groups of devices on the network.
func buildGroupPlmCommand(group byte, imCmd1, imCmd2 byte) []byte {
	return []byte{
		serialStart,
		cmdHostSendMsg,
		0, 0, group,
		CommandFlagBroadcast | CommandFlagGroup | CommandFlagHopsLeftThree | CommandFlagRetransmitMax,
		imCmd1, imCmd2,
	}
}

func buildExtPlmCommand(addr Address, imCmd1, imCmd2 byte, userData [14]byte) []byte {
	return []byte{
		serialStart,
		cmdHostSendMsg,
		addr[0], addr[1], addr[2],
		CommandFlagExtended | CommandFlagAck | CommandFlagHopsLeftThree | CommandFlagRetransmitMax,
		imCmd1, imCmd2,
		userData[0], userData[1], userData[2], userData[3], userData[4], userData[5], userData[6], userData[7],
		userData[8], userData[9], userData[10], userData[11], userData[12], userData[13],
	}
}

// ModemInfo describes an Insteon Power Line Modem, typically the Hub you're connected to.
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
