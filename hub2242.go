package insteon

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"time"
)

type Hub2242 struct {
	address  string
	conn     net.Conn
	buffer   []byte
	readChan chan []byte
	errChan  chan error
}

func NewHub2242(address string) (Hub, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	hub := &Hub2242{
		address:  address,
		conn:     conn,
		buffer:   []byte{},
		readChan: make(chan []byte),
		errChan:  make(chan error),
	}
	go hub.read()

	return hub, nil
}

func (hub *Hub2242) SendCommand(hostCmd byte, addr Address, imCmd1 byte, imCmd2 byte) (*CommandResponse, error) {
	if err := hub.ClearBuffer(); err != nil {
		return nil, err
	}
	cmd := buildPlmCommand(hostCmd, addr, imCmd1, imCmd2)
	_, err := hub.conn.Write(cmd)
	if err != nil {
		return nil, err
	}

	if err := hub.waitForAck(cmd); err != nil {
		return nil, err
	}
	return hub.waitForResponse()
}
func (hub *Hub2242) SendGroupCommand(hostCmd byte, group byte) error {
	return nil
}
func (hub *Hub2242) GetBuffer() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-hub.readChan:
		return hub.buffer, nil
	case err := <-hub.errChan:
		return nil, err
	case <-ctx.Done():
		return hub.buffer, nil
	}
}
func (hub *Hub2242) ClearBuffer() error {
	hub.buffer = []byte{}
	return nil
}

func (hub *Hub2242) waitForResponse() (*CommandResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for {
		select {
		case err := <-hub.errChan:
			return nil, err
		case <-hub.readChan:
			idx := bytes.Index(hub.buffer, []byte{0x02, 0x50})
			if idx >= 0 && idx+11 <= len(hub.buffer) {
				rsp := &CommandResponse{}
				rsp.fromBytes(hub.buffer[idx : idx+11])
				// We apparently have to wait here for a bit, otherwise sending another command quickly will cause
				// the PLM to freak out and reply with two NAKs.
				time.Sleep(200 * time.Millisecond)
				return rsp, nil
			}
		case <-ctx.Done():
			return nil, ErrAckTimeout
		}
	}
}

func (hub *Hub2242) waitForAck(cmd []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for {
		select {
		case err := <-hub.errChan:
			return err
		case <-hub.readChan:
			idx := bytes.Index(hub.buffer, cmd)
			if idx >= 0 && idx+len(cmd)+1 <= len(hub.buffer) {
				ack := hub.buffer[idx+len(cmd)]
				switch ack {
				case serialACK:
					//Move buffer forward
					hub.buffer = hub.buffer[idx+len(cmd):]
					return nil
				case serialNAK:
					return ErrNotReady
				default:
					return fmt.Errorf("unexpected acknowledgement byte: %d", ack)
				}
			}
		case <-ctx.Done():
			return ErrAckTimeout
		}
	}
}

func (hub *Hub2242) read() {
	buf := make([]byte, 255)
	for {
		cnt, err := hub.conn.Read(buf)
		if err != nil {
			hub.errChan <- err
			return
		} else {
			hub.buffer = append(hub.buffer, buf[0:cnt]...)
			hub.readChan <- buf
		}
	}
}
