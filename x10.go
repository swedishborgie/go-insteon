package insteon

type X10Raw byte

func (x X10Raw) HouseCode() X10HouseCode {
	return X10HouseCode(x & 0xF0 >> 4)
}

func (x X10Raw) UnitCode() byte {
	return byte(x & 0xF)
}

func (x X10Raw) Command() X10Command {
	return X10Command(x & 0xF)
}

type X10Flags byte

func (x X10Flags) Command() bool {
	return x&0x80 > 0
}

func (x X10Flags) UnitCode() bool {
	return !x.Command()
}

type X10HouseCode byte

const (
	X10HouseCodeA X10HouseCode = 0x6
	X10HouseCodeB X10HouseCode = 0xE
	X10HouseCodeC X10HouseCode = 0x2
	X10HouseCodeD X10HouseCode = 0xA
	X10HouseCodeE X10HouseCode = 0x1
	X10HouseCodeF X10HouseCode = 0x9
	X10HouseCodeG X10HouseCode = 0x5
	X10HouseCodeH X10HouseCode = 0xD
	X10HouseCodeI X10HouseCode = 0x7
	X10HouseCodeJ X10HouseCode = 0xF
	X10HouseCodeK X10HouseCode = 0x3
	X10HouseCodeL X10HouseCode = 0xB
	X10HouseCodeM X10HouseCode = 0x0
	X10HouseCodeN X10HouseCode = 0x8
	X10HouseCodeO X10HouseCode = 0x4
	X10HouseCodeP X10HouseCode = 0xC
)

type X10Command byte

const (
	X10CommandAllLightsOff X10Command = 0x0
	X10CommandStatusOff    X10Command = 0x1
	X10CommandOn           X10Command = 0x2
	X10CommandPresetDim1   X10Command = 0x3
	X10CommandAllLightsOn  X10Command = 0x4
	X10CommandHailAck      X10Command = 0x5
	X10CommandBright       X10Command = 0x6
	X10CommandStatusOn     X10Command = 0x7
	X10CommandExtendedCode X10Command = 0x8
	X10CommandStatusReq    X10Command = 0x9
	X10CommandOff          X10Command = 0xA
	X10CommandPresetDim2   X10Command = 0xB
	X10CommandAllUnitsOff  X10Command = 0xC
	X10CommandHailReq      X10Command = 0xD
	X10CommandDim          X10Command = 0xE
	X10CommandExtAnalog    X10Command = 0xF
)
