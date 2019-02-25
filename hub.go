package insteon

type Hub interface {
	SendCommand(hostCmd byte, addr []byte, imCmd1 byte, imCmd2 byte) ([]byte, error)
	SendGroupCommand(hostCmd byte, group byte) error
	GetBuffer() (chan []byte, chan error)
	ClearBuffer() error
	waitForAck(cmd []byte) error
}
