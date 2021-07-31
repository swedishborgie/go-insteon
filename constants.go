package insteon

import (
	"fmt"
	"time"
)

const (
	StreamingCommandPause = 200 * time.Millisecond
	ChannelBufferSize     = 10
)

const (
	serialStart byte = 0x02
	serialACK   byte = 0x06
	serialNAK   byte = 0x15
)

// These are used for the newer hubs.
const (
	cmdTypeShort byte = 0
	cmdTypeFull  byte = 3
)

// These are commands the modem can send to the host.
const (
	cmdIMStd              byte = 0x50
	cmdIMExt              byte = 0x51
	cmdIMX10              byte = 0x52
	cmdIMAllLinkComplete  byte = 0x53
	cmdIMBtnEvt           byte = 0x54
	cmdIMRst              byte = 0x55
	cmdIMAllLinkCleanFail byte = 0x56
	cmdIMAllLinkRecord    byte = 0x57
	cmdIMAllLinkCleanup   byte = 0x58
	cmdIMDatabaseRecord   byte = 0x59
)

// These are commands the host can send to the modem.
const (
	cmdHostGetInfo              byte = 0x60
	cmdHostAllLink              byte = 0x61
	cmdHostSendMsg              byte = 0x62
	cmdHostSendX10              byte = 0x63
	cmdHostStartAllLink         byte = 0x64
	cmdHostCancelAllLink        byte = 0x65
	cmdHostDeviceCategory       byte = 0x66
	cmdHostResetIM              byte = 0x67
	cmdHostACKMsgByte           byte = 0x68
	cmdHostFirstAllLinkRecord   byte = 0x69
	cmdHostGetNextAllLinkRecord byte = 0x6A
	cmdHostSetIMCFG             byte = 0x6B
	cmdHostAllLinkRecordSender  byte = 0x6C
	cmdHostLEDOn                byte = 0x6D
	cmdHostLEDOff               byte = 0x6E
	cmdHostMngAllLink           byte = 0x6F
	cmdHostSetNAKByte           byte = 0x70
	cmdHostSetACKBytes          byte = 0x71
	cmdHostRFSleep              byte = 0x72
	cmdHostIMCfg                byte = 0x73
	cmdHostCancelCleanup        byte = 0x74
	cmdHostReadDB               byte = 0x75
	cmdHostWriteDB              byte = 0x76
	cmdHostBeep                 byte = 0x77
	cmdHostSetStatus            byte = 0x78
)

// These are commands the modem can send to other Insteon devices.
const (
	cmdControlProduct    byte = 0x03
	cmdControlLink       byte = 0x09
	cmdControlUnlink     byte = 0x0a
	cmdControlPing       byte = 0x0F
	cmdControlID         byte = 0x10
	cmdControlOn         byte = 0x11
	cmdControlFastOn     byte = 0x12
	cmdControlOff        byte = 0x13
	cmdControlFastOff    byte = 0x14
	cmdControlBright     byte = 0x15
	cmdControlDim        byte = 0x16
	cmdControlStartDim   byte = 0x17
	cmdControlStopDim    byte = 0x18
	cmdControlStatus     byte = 0x19
	cmdControlGetOpFlags byte = 0x1f
	cmdControlSetOpFlags byte = 0x20
	cmdControlAllLink    byte = 0x2F
	cmdControlBeep       byte = 0x30

	cmdQueryIDRequest     byte = 0x10
	cmdQueryStatusRequest byte = 0x19
)

type LinkCode byte

const (
	LinkCodeResponder  = LinkCode(0)
	LinkCodeController = LinkCode(1)
	LinkCodeAuto       = LinkCode(3)
	LinkCodeDeleted    = LinkCode(0xFF)
	LinkCodeUnknown    = LinkCode(0xFE)
)

type ButtonEventType byte

const (
	ButtonEventSetTapped    = ButtonEventType(0x02)
	ButtonEventSetHeld      = ButtonEventType(0x03)
	ButtonEventSetReleased  = ButtonEventType(0x04)
	ButtonEventBtn2Tapped   = ButtonEventType(0x12)
	ButtonEventBtn2Held     = ButtonEventType(0x13)
	ButtonEventBtn2Released = ButtonEventType(0x14)
	ButtonEventBtn3Tapped   = ButtonEventType(0x22)
	ButtonEventBtn3Held     = ButtonEventType(0x23)
	ButtonEventBtn3Released = ButtonEventType(0x24)
	ButtonEventUnknown      = ButtonEventType(0xFF)
)

type AllLinkRecordFlags byte

const (
	AllLinkRecordFlagsInUse     AllLinkRecordFlags = 0x80
	AllLinkRecordFlagsContoller AllLinkRecordFlags = 0x40
	AllLinkRecordFlagsLast      AllLinkRecordFlags = 0x2
)

func (al AllLinkRecordFlags) InUse() bool {
	return al&0x80 > 0
}

func (al AllLinkRecordFlags) Controller() bool {
	return al&0x40 > 0
}

func (al AllLinkRecordFlags) Last() bool {
	return !(al&0x2 == 0)
}

func (al AllLinkRecordFlags) String() string {
	return fmt.Sprintf("InUse=%t, Controller=%t, Last=%t", al.InUse(), al.Controller(), al.Last())
}

type AllLinkCleanupStatus byte

const (
	AllLinkCleanupStatusSuccess = AllLinkCleanupStatus(serialACK)
	AllLinkCleanupStatusFailure = AllLinkCleanupStatus(serialNAK)
	AllLinkCleanupStatusUnknown = AllLinkCleanupStatus(0xFF)
)

type CommandFlag byte

const (
	CommandFlagExtended        = 0x10
	CommandFlagAck             = 0x20
	CommandFlagGroup           = 0x40
	CommandFlagBroadcast       = 0x80
	CommandFlagRetransmitNever = 0x0
	CommandFlagRetransmitOnce  = 0x1
	CommandFlagRetransmitTwice = 0x2
	CommandFlagRetransmitMax   = 0x3
	CommandFlagHopsLeftNone    = 0x0
	CommandFlagHopsLeftOne     = 0x4
	CommandFlagHopsLeftTwo     = 0x8
	CommandFlagHopsLeftThree   = 0xC
)

type ManageAllLinkCommand byte

const (
	ManageAllLinkFindFirst     ManageAllLinkCommand = 0x00
	ManageAllLinkFindNext      ManageAllLinkCommand = 0x01
	ManageAllLinkUpdate        ManageAllLinkCommand = 0x20
	ManageAllLinkAddController ManageAllLinkCommand = 0x40
	ManageAllLinkAddResponder  ManageAllLinkCommand = 0x41
	ManageAllLinkDelete        ManageAllLinkCommand = 0x80
)
