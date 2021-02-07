package insteon

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/pkg/errors"
)

type HubStreaming struct {
	stream    io.ReadWriteCloser
	buffer    []byte
	errChan   chan error
	events    chan Event
	ackBuffer []expectAck
	listeners []EventListener
	logger    CommLogger
}

type expectAck struct {
	cmd    []byte
	length int
}

var ErrUnexpectedAckByte = errors.New("unexpected acknowledgement byte")

const (
	StreamingCommandPause = 200 * time.Millisecond
	ChannelBufferSize     = 10
)

func NewHubStreaming(stream io.ReadWriteCloser) (*HubStreaming, error) {
	hub := &HubStreaming{
		stream:  stream,
		buffer:  []byte{},
		errChan: make(chan error, ChannelBufferSize),
		events:  make(chan Event, ChannelBufferSize),
	}
	go hub.read()

	return hub, nil
}

func (hub *HubStreaming) GetInfo(ctx context.Context) (*ModemInfo, error) {
	cmd := []byte{serialStart, cmdHostGetInfo}

	rsp, err := hub.directIMCommand(ctx, cmd, 9)
	if err != nil {
		return nil, err
	}

	mi := &ModemInfo{}

	mi.fromBytes(rsp[2:])

	return mi, nil
}

func (hub *HubStreaming) GetModemConfig(ctx context.Context) (ModemConfiguration, error) {
	cmd := []byte{serialStart, cmdHostIMCfg}

	rsp, err := hub.directIMCommand(ctx, cmd, 6)
	if err != nil {
		return 0, err
	}

	return ModemConfiguration(rsp[2]), nil
}

func (hub *HubStreaming) SetModemConfig(ctx context.Context, cfg ModemConfiguration) error {
	cmd := []byte{serialStart, cmdHostSetIMCFG, byte(cfg)}

	_, err := hub.directIMCommand(ctx, cmd, 4)
	if err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) StartAllLink(ctx context.Context, code LinkCode, group byte) (*AllLinkCompleted, error) {
	cmd := []byte{serialStart, cmdHostStartAllLink, byte(code), group}

	_, err := hub.directIMCommand(ctx, cmd, 5)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case evt := <-hub.events:
			if cevt, ok := evt.(*AllLinkCompleted); ok {
				return cevt, nil
			}
		case err := <-hub.errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (hub *HubStreaming) CancelAllLink(ctx context.Context) error {
	cmd := []byte{serialStart, cmdHostCancelAllLink}

	_, err := hub.directIMCommand(ctx, cmd, 3)
	if err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) Beep(ctx context.Context) error {
	cmd := []byte{serialStart, cmdHostBeep}

	_, err := hub.directIMCommand(ctx, cmd, 3)
	if err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) GetAllLinkDatabase(ctx context.Context) ([]*AllLinkRecord, error) {
	cmd := []byte{serialStart, cmdHostFirstAllLinkRecord}

	if _, err := hub.directIMCommand(ctx, cmd, 3); err != nil {
		return nil, err
	}

	var records []*AllLinkRecord

	for {
		select {
		case evt := <-hub.events:
			if alEvt, ok := evt.(*AllLinkRecord); ok {
				records = append(records, alEvt)

				cmd := []byte{serialStart, cmdHostGetNextAllLinkRecord}
				if _, err := hub.directIMCommand(context.Background(), cmd, 3); err != nil {
					if errors.Is(err, ErrNotReady) {
						// All set!
						return records, nil
					}

					return nil, err
				}
			}
		case err := <-hub.errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (hub *HubStreaming) manageAllLinkRecord(alCmd ManageAllLinkCommand, flags AllLinkRecordFlags, group byte, addr Address, data [3]byte) error {
	cmd := []byte{serialStart, cmdHostMngAllLink, byte(alCmd), byte(flags), group, addr[0], addr[1], addr[2], data[0], data[1], data[2]}
	if _, err := hub.directIMCommand(context.Background(), cmd, 12); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) SendCommand(ctx context.Context, addr Address, imCmd1 byte, imCmd2 byte) (*StdCommandResponse, error) {
	cmd := buildPlmCommand(addr, imCmd1, imCmd2)
	if _, err := hub.directIMCommand(ctx, cmd, len(cmd)+1); err != nil {
		return nil, err
	}

	return hub.waitForResponse()
}

func (hub *HubStreaming) SendExtendedCommand(ctx context.Context, addr Address, imCmd1, imCmd2 byte, userData [14]byte) (*StdCommandResponse, error) {
	cmd := buildExtPlmCommand(addr, imCmd1, imCmd2, userData)
	if _, err := hub.directIMCommand(ctx, cmd, len(cmd)+1); err != nil {
		return nil, err
	}

	return hub.waitForResponse()
}

func (hub *HubStreaming) SendX10(ctx context.Context, raw X10Raw, flags X10Flags) error {
	cmd := []byte{serialStart, cmdHostSendX10, byte(raw), byte(flags)}

	if _, err := hub.directIMCommand(ctx, cmd, 5); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) SendGroupCommand(ctx context.Context, cmd1 byte, group byte) error {
	cmd := buildGroupPlmCommand(group, cmd1, 0)

	if _, err := hub.directIMCommand(ctx, cmd, len(cmd)+1); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) SetDeviceCategory(ctx context.Context, cat Category, sub SubCategory, fw byte) error {
	cmd := []byte{serialStart, cmdHostDeviceCategory, byte(cat), byte(sub), fw}

	if _, err := hub.directIMCommand(ctx, cmd, 6); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) Sleep(ctx context.Context) error {
	cmd := []byte{serialStart, cmdHostRFSleep}

	if _, err := hub.directIMCommand(ctx, cmd, 3); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) Reset(ctx context.Context) error {
	cmd := []byte{serialStart, cmdHostResetIM}

	if _, err := hub.directIMCommand(ctx, cmd, 3); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) ReadDB(ctx context.Context, addr uint16) (*DatabaseRecord, error) {
	// The address must be aligned.
	if addr&0xF != 0 && addr&0xF != 0x8 {
		return nil, errors.New("address must be aligned to an 8 byte boundary")
	}

	cmd := []byte{serialStart, cmdHostReadDB, byte(addr & 0xFF00 >> 8), byte(addr & 0xFF)}

	_, err := hub.directIMCommand(ctx, cmd, 5)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case evt := <-hub.events:
			if db, ok := evt.(*DatabaseRecord); ok {
				return db, nil
			}
		case err := <-hub.errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (hub *HubStreaming) WriteDB(ctx context.Context, addr uint16, rec *AllLinkRecord) error {
	// The address must be aligned.
	if addr&0xF != 0 && addr&0xF != 0x8 {
		return errors.New("address must be aligned to an 8 byte boundary")
	}

	cmd := []byte{serialStart, cmdHostWriteDB, byte(addr & 0xFF00 >> 8), byte(addr & 0xFF)}
	cmd = append(cmd, rec.toBytes()...)

	_, err := hub.directIMCommand(ctx, cmd, len(cmd)+1)
	if err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) SetLED(ctx context.Context, on bool) error {
	cmd := []byte{serialStart, cmdHostLEDOff}
	if on {
		cmd[1] = cmdHostLEDOn
	}

	if _, err := hub.directIMCommand(ctx, cmd, len(cmd)+1); err != nil {
		return err
	}

	return nil
}

func (hub *HubStreaming) GetLastSender(ctx context.Context) (*AllLinkRecord, error) {
	cmd := []byte{serialStart, cmdHostAllLinkRecordSender}

	if _, err := hub.directIMCommand(ctx, cmd, len(cmd)+1); err != nil {
		return nil, err
	}

	for {
		select {
		case evt := <-hub.events:
			if db, ok := evt.(*AllLinkRecord); ok {
				return db, nil
			}
		case err := <-hub.errChan:
			return nil, err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (hub *HubStreaming) AddEventListener(listener EventListener) {
	hub.listeners = append(hub.listeners, listener)
}

func (hub *HubStreaming) RemoveEventListener(listener EventListener) {
	for idx := len(hub.listeners) - 1; idx >= 0; idx-- {
		if &hub.listeners[idx] == &listener {
			hub.listeners = append(hub.listeners[:idx], hub.listeners[idx+1:]...)
		}
	}
}

func (hub *HubStreaming) SetCommLogger(logger CommLogger) {
	hub.logger = logger
}

func (hub *HubStreaming) directIMCommand(ctx context.Context, cmd []byte, expect int) ([]byte, error) {
	hub.ackBuffer = append(hub.ackBuffer, expectAck{cmd: cmd, length: expect})

	if hub.logger != nil {
		hub.logger(CommDirectionHostToIM, cmd)
	}

	if _, err := hub.stream.Write(cmd); err != nil {
		return nil, err
	}

	for {
		select {
		case err := <-hub.errChan:
			return nil, err
		case evt := <-hub.events:
			if ack, ok := evt.(*Ack); ok {
				switch ack.Type {
				case serialNAK:
					return nil, ErrNotReady
				case serialACK:
					return ack.Response, nil
				default:
					return nil, errors.Wrapf(ErrUnexpectedAckByte, "byte: %x", ack.Type)
				}
			}
		case <-ctx.Done():
			return nil, ErrAckTimeout
		}
	}
}

func (hub *HubStreaming) waitForResponse() (*StdCommandResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	for {
		select {
		case err := <-hub.errChan:
			return nil, err
		case evt := <-hub.events:
			if stdEvt, ok := evt.(*StdCommandResponse); ok {
				// We apparently have to wait here for a bit, otherwise sending another command quickly will cause
				// the PLM to freak out and reply with two NAKs.
				time.Sleep(StreamingCommandPause)

				return stdEvt, nil
			}
		case <-ctx.Done():
			return nil, ErrAckTimeout
		}
	}
}

func (hub *HubStreaming) read() {
	buf := make([]byte, 255)

	for {
		cnt, err := hub.stream.Read(buf)
		if err != nil {
			hub.errChan <- err

			// Notify listeners
			for _, l := range hub.listeners {
				go l(nil, err)
			}

			return
		}

		if hub.logger != nil {
			hub.logger(CommDirectionIMToHost, buf[0:cnt])
		}

		hub.buffer = append(hub.buffer, buf[0:cnt]...)

		hub.parseBuffer()
	}
}

func (hub *HubStreaming) handleACK() bool {
	// First we need to check to see if we're waiting for any acks.
	if len(hub.ackBuffer) > 0 {
		// NAK's don't always echo commands, if our buffer starts with NAK, explode.
		if len(hub.buffer) > 0 && hub.buffer[0] == serialNAK {
			hub.ackBuffer = hub.ackBuffer[1:]
			ack := &Ack{Response: []byte{serialNAK}, Type: serialNAK}
			hub.buffer = hub.buffer[1:]
			hub.events <- ack

			return true
		}

		expected := hub.ackBuffer[0]

		idx := bytes.Index(hub.buffer, expected.cmd)
		if idx < 0 {
			return true
		} else if idx+expected.length > len(hub.buffer) {
			return true
		}

		// Got it
		hub.ackBuffer = hub.ackBuffer[1:]
		ack := &Ack{}
		ack.fromBytes(hub.buffer[idx : idx+expected.length])
		hub.buffer = hub.buffer[idx+expected.length:]
		hub.events <- ack

		hub.parseBuffer()
	}

	return false
}

func (hub *HubStreaming) parseBuffer() {
	// We need to handle ACKs first.
	if hub.handleACK() {
		return
	}

	// We'll scan the buffer looking for a start of serial command, followed by a known IM-to-Host command.
	idx := bytes.Index(hub.buffer, []byte{serialStart})
	if idx < 0 {
		// No start of serial command in buffer, nothing to do.
		return
	} else if idx+1 >= len(hub.buffer) {
		// Found a start of serial command, but there's nothing after it, abort.
		return
	}

	imCmd := imCommand(hub.buffer[idx+1])
	if imCmd == nil {
		// We found a start of serial, but it's not followed by a valid command, lets truncate the buffer to this point.
		hub.buffer = hub.buffer[idx+1:]

		return
	} else if imCmd.Length()+idx > len(hub.buffer) {
		// We don't have the expected length yet, we need to wait.
		return
	}

	imCmd.fromBytes(hub.buffer[idx : idx+imCmd.Length()])

	hub.events <- imCmd

	// Notify listeners
	for _, l := range hub.listeners {
		go l(imCmd, nil)
	}

	// Move buffer forward
	hub.buffer = hub.buffer[idx+imCmd.Length():]

	// Call this function again to make sure we didn't miss anything.
	hub.parseBuffer()
}

func imCommand(cmd byte) Event {
	cmdList := []Event{
		&StdCommandResponse{},
		&ExtCommandResponse{},
		&X10Response{},
		&AllLinkCompleted{},
		&ButtonEvent{},
		&UserReset{},
		&AllLinkCleanupFailure{},
		&AllLinkRecord{},
		&AllLinkCleanup{},
		&DatabaseRecord{},
	}

	for _, c := range cmdList {
		if c.ID() == cmd {
			return c
		}
	}

	return nil
}
