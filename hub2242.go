package insteon

import (
	"net"
)

type Hub2242 struct {
	address string
	conn    net.Conn
	*HubStreaming
}

func NewHub2242(address string) (Hub, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	streamHub, err := NewHubStreaming(conn)
	if err != nil {
		return nil, err
	}

	hub := &Hub2242{
		address:      address,
		conn:         conn,
		HubStreaming: streamHub,
	}

	return hub, nil
}
