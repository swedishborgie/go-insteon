package insteon

import (
	"github.com/tarm/serial"
)

type HubPLM struct {
	port *serial.Port
	Hub
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
		port: port,
		Hub:  streamingHub,
	}

	return hub, nil
}
