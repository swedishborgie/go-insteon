package insteon

import (
	"bytes"
	"fmt"
	"net"
	"time"
)

type Hub2242 struct {
	address    string
	conn       net.Conn
	buffer     []byte
	readChan   chan []byte
	readErrors chan error
}

func NewHub2242(address string) (Hub, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	hub := &Hub2242{
		address:    address,
		conn:       conn,
		buffer:     []byte{},
		readChan:   make(chan []byte),
		readErrors: make(chan error),
	}
	go hub.read()

	return hub, nil
}

func (hub *Hub2242) SendCommand(hostCmd byte, addr []byte, imCmd1 byte, imCmd2 byte) ([]byte, error) {
	hub.ClearBuffer()
	cmd := buildPlmCommand(hostCmd, addr, imCmd1, imCmd2)
	cnt, err := hub.conn.Write(cmd)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Wrote %d bytes: ", cnt)
	for i := 0; i < cnt; i++ {
		fmt.Printf("%02x ", cmd[i])
	}
	fmt.Println()
	time.Sleep(1 * time.Second)
	return cmd, hub.waitForAck(cmd)
}
func (hub *Hub2242) SendGroupCommand(hostCmd byte, group byte) error {
	return nil
}
func (hub *Hub2242) GetBuffer() (chan []byte, chan error) {
	return hub.readChan, hub.readErrors
}
func (hub *Hub2242) ClearBuffer() error {
	hub.buffer = []byte{}
	return nil
}

func (hub *Hub2242) waitForAck(cmd []byte) error {
	bufChan, errChan := hub.GetBuffer()
	for {
		select {
		case err := <-errChan:
			return err
		case buf := <-bufChan:
			idx := bytes.Index(buf, cmd)
			fmt.Printf("idx: %d, len: %d,\ncmd: ", idx, len(buf))
			for i := 0; i < len(cmd); i++ {
				fmt.Printf("%02x ", cmd[i])
			}
			fmt.Printf("buf: ")
			for i := 0; i < len(buf); i++ {
				fmt.Printf("%02x ", buf[i])
			}
			fmt.Println()
			if idx >= 0 && idx+len(cmd)+1 <= len(buf) {
				if buf[idx+len(cmd)] == serialACK {
					return nil
				} else if buf[idx+len(cmd)] == serialNAK {
					return fmt.Errorf("device not ready for commands")
				}
			}
		}
	}
	return fmt.Errorf("timeout waiting for device")
}

func (hub *Hub2242) read() {
	buf := make([]byte, 255)
	var err error
	var cnt int
	for err == nil {
		cnt, err = hub.conn.Read(buf)
		if err != nil {
			hub.readErrors <- err
		}
		if cnt > 0 {
			fmt.Printf("Got %d bytes: ", cnt)
			for i := 0; i < cnt; i++ {
				fmt.Printf("%02x ", buf[i])
			}
			fmt.Println()
			hub.buffer = append(hub.buffer, buf[0:cnt]...)
			hub.readChan <- hub.buffer
		}
	}
}
