# Insteon Hub Library for Go
A Go library for interfacing with Insteon Hubs. Currently, this library supports: the original 2242-222 (HUB1), the
2245-222 (HUB2), and the PLM Modems (2413S/2413U).

The library supports most of the Insteon PLM protocol and is capable of communicating with most devices and managing the
All-Link databases on the modems and devices.

Usage Example:

```go
    package example

    import(
    	"context"
    	
    	"github.com/swedishborgie/go-insteon"
    )
    
    func toggleLight() error {
        hub, err := insteon.NewHub2242("<address>:9761")
        // or
        // hub, err := insteon.NewHub2245("<address>", "<username>", "<password>")
        // or
        // hub, err := insteon.NewHubPLM("<serial port path>")
        if err != nil {
            return err
        }
    
        device, err := insteon.NewDevice(hub, insteon.Address{0x12, 0x34, 0x56})
        if err != nil {
            return err
        }

        ctx := context.Background()
    
        status, err := device.GetStatus(ctx)
        if err != nil {
            return err
        }
    
        if status.Level > 0 {
            return device.TurnOff(ctx)
        } else {
            return device.TurnOn(ctx)
        }
    }
```

Documentation for the Insteon devices can be found here:
* [INSTEON Modem Developer's Guide](https://cache.insteon.com/pdf/INSTEON_Modem_Developer%27s_Guide_20071012a.pdf)
* [INSTEON Hub: Developer's Guide](http://cache.insteon.com/developer/2242-222dev-062013-en.pdf)
* [INSTEON Developer's Guide: 2nd Edition](http://cache.insteon.com/pdf/INSTEON_Developers_Guide_20070816a.pdf)

A huge thank you to the [insteon-mqtt](https://github.com/TD22057/insteon-mqtt) project for documenting some of the
more arcane idiosyncrasies associated with the protocol and devices.