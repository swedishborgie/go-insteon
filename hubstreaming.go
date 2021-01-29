package insteon

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
)

type HubStreaming struct {
	stream   io.ReadWriteCloser
	buffer   []byte
	readChan chan []byte
	errChan  chan error
}

var ErrUnexpectedAckByte = errors.New("unexpected acknowledgement byte")

const (
	StreamingResponseTimeout = 5 * time.Second
	StreamingCommandPause    = 200 * time.Millisecond
)

func NewHubStreaming(stream io.ReadWriteCloser) (Hub, error) {
	hub := &HubStreaming{
		stream:   stream,
		buffer:   []byte{},
		readChan: make(chan []byte),
		errChan:  make(chan error),
	}
	go hub.read()

	return hub, nil
}

func (hub *HubStreaming) SendCommand(hostCmd byte, addr Address, imCmd1 byte, imCmd2 byte) (*CommandResponse, error) {
	if err := hub.ClearBuffer(); err != nil {
		return nil, err
	}

	cmd := buildPlmCommand(hostCmd, addr, imCmd1, imCmd2)

	if _, err := hub.stream.Write(cmd); err != nil {
		return nil, err
	}

	if err := hub.waitForAck(cmd); err != nil {
		return nil, err
	}

	return hub.waitForResponse()
}

func (hub *HubStreaming) SendExtendedCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte, userData [14]byte) (*CommandResponse, error) {
	if err := hub.ClearBuffer(); err != nil {
		return nil, err
	}

	cmd := buildExtPlmCommand(hostCmd, addr, imCmd1, imCmd2, userData)

	if _, err := hub.stream.Write(cmd); err != nil {
		return nil, err
	}

	if err := hub.waitForAck(cmd); err != nil {
		return nil, err
	}

	return hub.waitForResponse()
}

func (hub *HubStreaming) SendGroupCommand(byte, byte) error {
	return nil
}

func (hub *HubStreaming) GetBuffer() ([]byte, error) {
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

func (hub *HubStreaming) ClearBuffer() error {
	hub.buffer = []byte{}

	return nil
}

func (hub *HubStreaming) waitForResponse() (*CommandResponse, error) {
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

				rspBytes := hub.buffer[idx : idx+11]

				rsp.fromBytes(rspBytes)
				// Move buffer forward
				hub.buffer = hub.buffer[idx+11:]
				// We apparently have to wait here for a bit, otherwise sending another command quickly will cause
				// the PLM to freak out and reply with two NAKs.
				time.Sleep(StreamingCommandPause)

				return rsp, nil
			}
		case <-ctx.Done():
			return nil, ErrAckTimeout
		}
	}
}

func (hub *HubStreaming) waitForAck(cmd []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), StreamingResponseTimeout)
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
					// Move buffer forward
					hub.buffer = hub.buffer[idx+len(cmd):]

					return nil
				case serialNAK:
					return ErrNotReady
				default:
					return errors.Wrapf(ErrUnexpectedAckByte, "%d", ack)
				}
			}
		case <-ctx.Done():
			return ErrAckTimeout
		}
	}
}

func (hub *HubStreaming) read() {
	buf := make([]byte, 255)

	for {
		cnt, err := hub.stream.Read(buf)
		if err != nil {
			hub.errChan <- err

			return
		}

		hub.buffer = append(hub.buffer, buf[0:cnt]...)
		hub.readChan <- buf
	}
}
