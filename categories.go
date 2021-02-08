package insteon

type Category byte

const (
	CategoryGeneralController    Category = 0x00
	CategoryDimmableLighting     Category = 0x01
	CategorySwitchedLighting     Category = 0x02
	CategoryNetworkBridge        Category = 0x03
	CategoryIrrigation           Category = 0x04
	CategoryClimate              Category = 0x05
	CategoryPoolAndSpa           Category = 0x06
	CategorySensorsAndActuators  Category = 0x07
	CategoryHomeEntertainment    Category = 0x08
	CategoryEnergyManagement     Category = 0x09
	CategoryAppliance            Category = 0x0A
	CategoryPlumbing             Category = 0x0B
	CategoryCommunication        Category = 0x0C
	CategoryComputer             Category = 0x0D
	CategoryWindowCovering       Category = 0x0E
	CategoryAccess               Category = 0x0F
	CategorySecurityHealthSafety Category = 0x10
	CategorySurveillance         Category = 0x11
	CategoryAutomotive           Category = 0x12
	CategoryPetCare              Category = 0x13
	CategoryToys                 Category = 0x14
	CategoryTimekeeping          Category = 0x15
	CategoryHoliday              Category = 0x16
	CategoryReserved             Category = 0x17
	CategoryUnassigned           Category = 0xFF
)

func (c Category) String() string {
	switch c {
	case CategoryGeneralController:
		return "Generalized Controllers"
	case CategoryDimmableLighting:
		return "Dimmable Lighting Control"
	case CategorySwitchedLighting:
		return "Switched Lighting Control"
	case CategoryNetworkBridge:
		return "Network Bridges"
	case CategoryIrrigation:
		return "Irrigation Control"
	case CategoryClimate:
		return "Climate Control"
	case CategoryPoolAndSpa:
		return "Pool and Spa Control"
	case CategorySensorsAndActuators:
		return "Sensors and Actuators"
	case CategoryHomeEntertainment:
		return "Home Entertainment"
	case CategoryEnergyManagement:
		return "Energy Management"
	case CategoryAppliance:
		return "Built-In Appliance Control"
	case CategoryPlumbing:
		return "Plumbing"
	case CategoryCommunication:
		return "Communication"
	case CategoryComputer:
		return "Computer Control"
	case CategoryWindowCovering:
		return "Window Coverings"
	case CategoryAccess:
		return "Access Control"
	case CategorySecurityHealthSafety:
		return "Security, Health, Safety"
	case CategorySurveillance:
		return "Surveillance"
	case CategoryAutomotive:
		return "Automotive"
	case CategoryPetCare:
		return "Pet Care"
	case CategoryToys:
		return "Toys"
	case CategoryTimekeeping:
		return "Timekeeping"
	case CategoryHoliday:
		return "Holiday"
	case CategoryUnassigned:
		return "Unassigned"
	case CategoryReserved:
		fallthrough
	default:
		return "Reserved"
	}
}

type SubCategory byte

type Product struct {
	Category    Category
	SubCategory SubCategory
	ProductKey  uint
	Description string
}

func GetProductDesc(cat Category, sub SubCategory) string {
	subCats, ok := products[cat]
	if !ok {
		return ""
	}

	prd, ok := subCats[sub]
	if !ok {
		return ""
	}

	return prd.Description
}

var products = map[Category]map[SubCategory]Product{
	CategoryGeneralController: {
		0x04: Product{ProductKey: 0x000000, Description: "ControLinc [2430]"},
		0x05: Product{ProductKey: 0x000034, Description: "RemoteLinc [2440]"},
		0x06: Product{ProductKey: 0x000000, Description: "Icon Tabletop Controller [2830]"},
		0x08: Product{ProductKey: 0x00003D, Description: "EZBridge/EZServer"},
		0x09: Product{ProductKey: 0x000000, Description: "SignaLinc RF Signal Enhancer [2442]"},
		0x0A: Product{ProductKey: 0x000007, Description: "Balboa Instrument’s Poolux LCD Controller"},
		0x0B: Product{ProductKey: 0x000022, Description: "Access Point [2443]"},
		0x0C: Product{ProductKey: 0x000028, Description: "IES Color Touchscreen"},
		0x0D: Product{ProductKey: 0x00004D, Description: "SmartLabs KeyFOB [????]"},
	},
	CategoryDimmableLighting: {
		0x00: Product{ProductKey: 0x000000, Description: "LampLinc V2 [2456D3]"},
		0x01: Product{ProductKey: 0x000000, Description: "SwitchLinc V2 Dimmer 600W [2476D]"},
		0x02: Product{ProductKey: 0x000000, Description: "In-LineLinc Dimmer [2475D]"},
		0x03: Product{ProductKey: 0x000000, Description: "Icon Switch Dimmer [2876D]"},
		0x04: Product{ProductKey: 0x000000, Description: "SwitchLinc V2 Dimmer 1000W [2476DH]"},
		0x05: Product{ProductKey: 0x000041, Description: "KeypadLinc Dimmer Countdown Timer [2484DWH8]"},
		0x06: Product{ProductKey: 0x000000, Description: "LampLinc 2-Pin [2456D2]"},
		0x07: Product{ProductKey: 0x000000, Description: "Icon LampLinc V2 2-Pin [2856D2]"},
		0x08: Product{ProductKey: 0x000040, Description: "SwitchLinc Dimmer Count-down Timer [2484DWH8]"},
		0x09: Product{ProductKey: 0x000037, Description: "KeypadLinc Dimmer [2486D]"},
		0x0A: Product{ProductKey: 0x000000, Description: "Icon In-Wall Controller [2886D]"},
		0x0B: Product{ProductKey: 0x00001c, Description: "Access Point LampLinc [2458D3]"},
		0x0C: Product{ProductKey: 0x00001d, Description: "KeypadLinc Dimmer – 8-Button defaulted mode [2486DWH8]"},
		0x0D: Product{ProductKey: 0x00001e, Description: "SocketLinc [2454D]"},
		0x0E: Product{ProductKey: 0x00004b, Description: "LampLinc Dimmer, Dual-Band [2457D3]"},
		0x13: Product{ProductKey: 0x000032, Description: "ICON SwitchLinc Dimmer for Lixar/Bell Canada [2676D-B]"},
		0x17: Product{ProductKey: 0x000000, Description: "ToggleLinc Dimmer [2466D]"},
		0x18: Product{ProductKey: 0x00003f, Description: "Icon SL Dimmer Inline Companion [2474D]"},
		0x19: Product{ProductKey: 0x00004e, Description: "SwitchLinc 800W"},
		0x1A: Product{ProductKey: 0x00004f, Description: "In-LineLinc Dimmer with Sense [2475D2]"},
		0x1B: Product{ProductKey: 0x000050, Description: "KeypadLinc 6-button Dimmer [2486DWH6]"},
		0x1C: Product{ProductKey: 0x000051, Description: "KeypadLinc 8-button Dimmer [2486DWH8]"},
		0x1D: Product{ProductKey: 0x000052, Description: "SwitchLinc Dimmer 1200W [2476D]"},
		0x3a: Product{ProductKey: 0x0, Description: "LED Bulb [2672-222]"},
	},
	CategorySwitchedLighting: {
		0x05: Product{ProductKey: 0x000042, Description: "KeypadLinc Relay – 8-Button defaulted mode [2486SWH8]"},
		0x06: Product{ProductKey: 0x000048, Description: "Outdoor ApplianceLinc [2456S3E]"},
		0x07: Product{ProductKey: 0x000029, Description: "TimerLinc [2456ST3]"},
		0x08: Product{ProductKey: 0x000023, Description: "OutletLinc [2473S]"},
		0x09: Product{ProductKey: 0x000000, Description: "ApplianceLinc [2456S3]"},
		0x0a: Product{ProductKey: 0x000000, Description: "SwitchLinc Relay [2476S]"},
		0x0b: Product{ProductKey: 0x000000, Description: "Icon On Off Switch [2876S]"},
		0x0c: Product{ProductKey: 0x000000, Description: "Icon Appliance Adapter [2856S3]"},
		0x0d: Product{ProductKey: 0x000000, Description: "ToggleLinc Relay [2466S]"},
		0x0e: Product{ProductKey: 0x000000, Description: "SwitchLinc Relay Countdown Timer [2476ST]"},
		0x0f: Product{ProductKey: 0x000036, Description: "KeypadLinc On/Off Switch [2486SWH6]"},
		0x10: Product{ProductKey: 0x00001b, Description: "In-LineLinc Relay [2475D]"},
		0x11: Product{ProductKey: 0x00003c, Description: "EZSwitch30 (240V, 30A load controller)"},
		0x12: Product{ProductKey: 0x00003e, Description: "Icon SL Relay Inline Companion"},
		0x13: Product{ProductKey: 0x000033, Description: "ICON SwitchLinc Relay for Lixar/Bell Canada [2676R-B]"},
		0x14: Product{ProductKey: 0x000045, Description: "In-LineLinc Relay with Sense [2475S2]"},
		0x15: Product{ProductKey: 0x000047, Description: "SwitchLinc Relay with Sense [2476S2]"},
	},
	CategoryNetworkBridge: {
		0x01: Product{ProductKey: 0x000000, Description: "PowerLinc Serial [2414S]"},
		0x02: Product{ProductKey: 0x000000, Description: "PowerLinc USB [2414U]"},
		0x03: Product{ProductKey: 0x000000, Description: "Icon PowerLinc Serial [2814 S]"},
		0x04: Product{ProductKey: 0x000000, Description: "Icon PowerLinc USB [2814U]"},
		0x05: Product{ProductKey: 0x00000c, Description: "SmartLabs PowerLinc Modem Serial [2412S]"},
		0x06: Product{ProductKey: 0x000016, Description: "SmartLabs IR to Insteon Interface [2411R]"},
		0x07: Product{ProductKey: 0x000017, Description: "SmartLabs IRLinc - IR Transmitter Interface [2411T]"},
		0x08: Product{ProductKey: 0x000018, Description: "SmartLabs Bi-Directional IR -Insteon Interface"},
		0x09: Product{ProductKey: 0x000019, Description: "SmartLabs RF Developer’s Board [2600RF]"},
		0x0a: Product{ProductKey: 0x00000d, Description: "SmartLabs PowerLinc Modem Ethernet [2412E]"},
		0x0b: Product{ProductKey: 0x000030, Description: "SmartLabs PowerLinc Modem USB [2412U]"},
		0x0c: Product{ProductKey: 0x000031, Description: "SmartLabs PLM Alert Serial"},
		0x0d: Product{ProductKey: 0x000036, Description: "SimpleHomeNet EZX10RF"},
		0x0e: Product{ProductKey: 0x00002c, Description: "X10 TW-523/PSC05 Translator"},
		0x0f: Product{ProductKey: 0x00003b, Description: "EZX10IR (X10 IR receiver, Insteon controller and IR"},
		0x10: Product{ProductKey: 0x000044, Description: "distribution hub)"},
		0x11: Product{ProductKey: 0x000045, Description: "SmartLinc 2412N INSTEON Central Controller"},
		0x12: Product{ProductKey: 0x00004c, Description: "PowerLinc - Serial (Dual Band) [2413S]"},
		0x13: Product{ProductKey: 0x000053, Description: "RF Modem Card [new smaller board]"},
		0x14: Product{ProductKey: 0x000054, Description: "PowerLinc USB – HouseLinc 2 enabled [2412UH]"},
		0x37: Product{ProductKey: 0x000000, Description: "Insteon Hub 1 [2242]"},
	},
	CategoryIrrigation: {
		0x00: Product{ProductKey: 0x000001, Description: "Compacta EZRain Sprinkler Controller"},
	},
	CategoryClimate: {
		0x00: Product{ProductKey: 0x000000, Description: "Broan SMSC080 Exhaust Fan"},
		0x01: Product{ProductKey: 0x000002, Description: "Compacta EZTherm"},
		0x02: Product{ProductKey: 0x000000, Description: "Broan SMSC110 Exhaust Fan"},
		0x03: Product{ProductKey: 0x00001f, Description: "INSTEON Thermostat Adapter [2441V]"},
		0x04: Product{ProductKey: 0x000024, Description: "Compacta EZThermx Thermostat"},
		0x05: Product{ProductKey: 0x000038, Description: "Broan, Venmar, BEST Rangehoods"},
		0x06: Product{ProductKey: 0x000043, Description: "Broan SmartSense Make-up Damper"},
	},
	CategoryPoolAndSpa: {
		0x00: Product{ProductKey: 0x000003, Description: "Compacta EZPool"},
		0x01: Product{ProductKey: 0x000008, Description: "Low-end pool controller (Temp. Eng. Project name)"},
		0x02: Product{ProductKey: 0x000009, Description: "Mid-Range pool controller (Temp. Eng. Project name)"},
		0x03: Product{ProductKey: 0x00000a, Description: "Next Generation pool controller (Temp. Eng. Project name)"},
	},
	CategorySensorsAndActuators: {
		0x00: Product{ProductKey: 0x00001a, Description: "IOLinc [2450]"},
		0x01: Product{ProductKey: 0x000004, Description: "Compacta EZSns1W Sensor Interface Module"},
		0x02: Product{ProductKey: 0x000012, Description: "Compacta EZIO8T I/O Module"},
		0x03: Product{ProductKey: 0x000005, Description: "Compacta EZIO2X4 #5010D"},
		0x04: Product{ProductKey: 0x000013, Description: "INSTEON / X10 Input/Output Module"},
		0x05: Product{ProductKey: 0x000014, Description: "Compacta EZIO8SA I/O Module"},
		0x06: Product{ProductKey: 0x000015, Description: "Compacta EZSnsRF #5010E RF Receiver Interface"},
		0x07: Product{ProductKey: 0x000039, Description: "Module for Dakota Alerts Products"},
		0x08: Product{ProductKey: 0x00003a, Description: "Compacta EZISnsRf Sensor Interface Module"},
	},
	CategoryEnergyManagement: {
		0x00: Product{ProductKey: 0x000006, Description: "Compacta EZEnergy"},
		0x01: Product{ProductKey: 0x000020, Description: "OnSitePro Leak Detector"},
		0x02: Product{ProductKey: 0x000021, Description: "OnSitePro Control Valve"},
		0x03: Product{ProductKey: 0x000025, Description: "Energy Inc. TED 5000 Single Phase Measuring Transmitting Unit (MTU)"},
		0x04: Product{ProductKey: 0x000026, Description: "Energy Inc. TED 5000 Gateway - USB"},
		0x05: Product{ProductKey: 0x00002a, Description: "Energy Inc. TED 5000 Gateway - Ethernet"},
		0x06: Product{ProductKey: 0x00002b, Description: "Energy Inc. TED 3000 Three Phase Measuring Transmitting Unit (MTU)"},
	},
	CategoryWindowCovering: {
		0x00: Product{ProductKey: 0x00000B, Description: "Somfy Drape Controller RF Bridge"},
	},
	CategoryAccess: {
		0x00: Product{ProductKey: 0x00000E, Description: "Weiland Doors’ Central Drive and Controller"},
		0x01: Product{ProductKey: 0x00000F, Description: "Weiland Doors’ Secondary Central Drive"},
		0x02: Product{ProductKey: 0x000010, Description: "Weiland Doors’ Assist Drive"},
		0x03: Product{ProductKey: 0x000011, Description: "Weiland Doors’ Elevation Drive"},
	},
	CategorySecurityHealthSafety: {
		0x00: Product{ProductKey: 0x000027, Description: "First Alert ONELink RF to Insteon Bridge"},
		0x01: Product{ProductKey: 0x00004a, Description: "Motion Sensor [2420M]"},
		0x02: Product{ProductKey: 0x000049, Description: "TriggerLinc - INSTEON Open / Close Sensor [2421]"},
	},
}
