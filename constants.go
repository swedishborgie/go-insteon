package insteon

const (
	serialStart byte = 0x02
	serialACK   byte = 0x06
	serialNAK   byte = 0x15
)

const (
	cmdTypeShort byte = 0
	cmdTypeFull  byte = 3
)

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
	cmdIMDBRecord         byte = 0x59

	cmdHostGetInfo            byte = 0x60
	cmdHostAllLink            byte = 0x61
	cmdHostSendMsg            byte = 0x62
	cmdHostSendX10            byte = 0x63
	cmdHostStartAllLink       byte = 0x64
	cmdHostCancelAllLink      byte = 0x65
	cmdHostDeviceCategory     byte = 0x66
	cmdHostResetIM            byte = 0x67
	cmdHostACKMsgByte         byte = 0x68
	cmdHostFirstAllLinkRecord byte = 0x69
	cmdHostLastAllLinkRecord  byte = 0x6A
	cmdHostSetIMCFG           byte = 0x6B
	cmdHostAllLinkRecord      byte = 0x6C
	cmdHostLEDOn              byte = 0x6D
	cmdHostLEDOff             byte = 0x6E
	cmdHostMngAllLink         byte = 0x6F
	cmdHostSetNAKByte         byte = 0x70
	cmdHostSetACKBytes        byte = 0x71
	cmdHostRFSleep            byte = 0x72
	cmdHostIMCfg              byte = 0x73
	cmdHostCancelCleanup      byte = 0x74
	cmdHostReadDB             byte = 0x75
	cmdHostWriteDB            byte = 0x76
	cmdHostBeep               byte = 0x77
	cmdHostSetStatus          byte = 0x78

	cmdControlOn       byte = 0x11
	cmdControlFastOn   byte = 0x12
	cmdControlOff      byte = 0x13
	cmdControlFastOff  byte = 0x14
	cmdControlBright   byte = 0x15
	cmdControlDim      byte = 0x16
	cmdControlStartDim byte = 0x17
	cmdcontrolStopDim  byte = 0x18

	cmdQueryIDRequest     byte = 0x10
	cmdQueryStatusRequest byte = 0x19
)
