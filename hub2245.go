package insteon

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Hub2245 is a reference to an Insteon Hub
type Hub2245 struct {
	address    string
	userName   string
	password   string
	readChan   chan []byte
	readErrors chan error
}

// NewHub2245 creates a new reference to an Insteon hub.
func NewHub2245(address string, userName string, password string) Hub {
	return &Hub2245{
		address:    address,
		userName:   userName,
		password:   password,
		readChan:   make(chan []byte, 10),
		readErrors: make(chan error, 10),
	}
}

// GetBuffer returns the current PLM buffer from the Insteon hub.
func (hub *Hub2245) GetBuffer() (chan []byte, chan error) {
	resp, err := hub.doRequest("/buffstatus.xml")
	if err != nil {
		hub.readErrors <- err
		return hub.readChan, hub.readErrors
	}

	buffer, err := parseBufferResponse(resp.Body)
	if err != nil {
		hub.readErrors <- err
		return hub.readChan, hub.readErrors
	}
	hub.readChan <- buffer
	return hub.readChan, hub.readErrors
}

// ClearBuffer clears the PLM buffer.
func (hub *Hub2245) ClearBuffer() error {
	_, err := hub.doRequest("/1?XB=M=1")
	return err
}

// SendCommand sends a standard command to a specific Insteon device.
func (hub *Hub2245) SendCommand(hostCmd byte, addr []byte, imCmd1 byte, imCmd2 byte) ([]byte, error) {
	plmCmd := buildPlmCommand(hostCmd, addr, imCmd1, imCmd2)
	uri := fmt.Sprintf("/%X?%X=I=%X", cmdTypeFull, plmCmd, cmdTypeFull)
	_, err := hub.doRequest(uri)
	return plmCmd, err
}

// SendGroupCommand sends a command to a group.
func (hub *Hub2245) SendGroupCommand(hostCmd byte, group byte) error {
	uri := fmt.Sprintf("/%X?%02X%02X=I=%X", cmdTypeShort, hostCmd, group, cmdTypeShort)
	_, err := hub.doRequest(uri)
	return err
}

func (hub *Hub2245) waitForAck(cmd []byte) error {
	for i := 0; i < 5; i++ {
		bufChan, errChan := hub.GetBuffer()
		select {
		case err := <-errChan:
			return err
		case buf := <-bufChan:
			idx := bytes.Index(buf, cmd)
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

// buildPlmCommand builds a power line modem command that gets sent by the
// Insteon hub.
func buildPlmCommand(hostCmd byte, addr []byte, imCmd1, imCmd2 byte) []byte {
	return []byte{
		serialStart,
		hostCmd,
		addr[0], addr[1], addr[2],
		0x0f,
		imCmd1, imCmd2,
	}
}

// doRequest submits a request to the Insteon hub.
func (hub *Hub2245) doRequest(uri string) (*http.Response, error) {
	req, err := http.NewRequest("POST", hub.address+uri, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(hub.userName, hub.password)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// bufResponse wraps the XML response from buffstatus.xml
type bufResponse struct {
	Buffer string `xml:"BS"`
}

// parseBufferResponse reads the buffstatus.xml from the hub.
func parseBufferResponse(body io.Reader) ([]byte, error) {
	dec := xml.NewDecoder(body)
	buf := &bufResponse{}
	err := dec.Decode(buf)
	if err != nil {
		return nil, err
	}

	bufBytes, err := hex.DecodeString(buf.Buffer)
	if err != nil {
		return nil, err
	}
	return bufBytes, nil
}
