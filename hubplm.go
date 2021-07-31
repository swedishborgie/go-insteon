package insteon

import (
	"github.com/tarm/serial"
)

type HubPLM struct {
	port *serial.Port
	*HubStreaming
}

const PLMBaudRate = 19200

func NewHubPLM(dev string) (Hub, error) {
	port, err := serial.OpenPort(&serial.Config{Name: dev, Baud: PLMBaudRate})
	if err != nil {
		return nil, err
	}

	streamingHub, err := NewHubStreaming(port)
	if err != nil {
		return nil, err
	}

	hub := &HubPLM{
		port:         port,
		HubStreaming: streamingHub,
	}

	return hub, nil
}
