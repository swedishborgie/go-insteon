# go-insteon
A Go library for interfacing with Insteon Hubs. Currently this library supports both
the original 2242-222 (HUB1) and the 2245-222 (HUB2). This library is currently a work in
progress but is currently of capable of supporting most typical use cases.

Usage Example:

    func toggleLight() error {
        hub, err := insteon.NewHub2242("<address>:9761")
        // or
        // hub, err := insteon.NewHub2245("<address>", "<username>", "<password>")
        if err != nil {
            return err
        }
    
        device, err := insteon.NewDevice(hub, insteon.Address{0x12, 0x34, 0x56})
        if err != nil {
            return err
        }
    
        status, err := device.GetStatus()
        if err != nil {
            return err
        }
    
        if status.Level > 0 {
            return device.TurnOff()
        } else {
            return device.TurnOn()
        }
    }
    

Documentation for the PLM commands can be found here:
* [INSTEON Modem Developer's Guide](https://cache.insteon.com/pdf/INSTEON_Modem_Developer%27s_Guide_20071012a.pdf)