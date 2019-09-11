package insteon

import "github.com/pkg/errors"

var (
	ErrNotReady   = errors.New("device not ready")
	ErrAckTimeout = errors.New("ack timeout")
)

type CommandResponse struct {
	From  Address
	To    Address
	Flags byte
	Cmd1  byte
	Cmd2  byte
}

func (cr *CommandResponse) fromBytes(buffer []byte) {
	copy(cr.From[:], buffer[2:5])
	copy(cr.To[:], buffer[5:8])
	cr.Flags = buffer[8]
	cr.Cmd1 = buffer[9]
	cr.Cmd2 = buffer[10]
}

type Hub interface {
	SendCommand(hostCmd byte, addr Address, imCmd1 byte, imCmd2 byte) (*CommandResponse, error)
	SendGroupCommand(hostCmd byte, group byte) error
	GetBuffer() ([]byte, error)
	ClearBuffer() error
	waitForAck(cmd []byte) error
	waitForResponse() (*CommandResponse, error)
}
