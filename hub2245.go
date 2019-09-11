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
	address  string
	userName string
	password string
}

// NewHub2245 creates a new reference to an Insteon hub.
func NewHub2245(address string, userName string, password string) Hub {
	return &Hub2245{
		address:  address,
		userName: userName,
		password: password,
	}
}

// GetBuffer returns the current PLM buffer from the Insteon hub.
func (hub *Hub2245) GetBuffer() ([]byte, error) {
	resp, err := hub.doRequest("/buffstatus.xml")
	if err != nil {
		return nil, err
	}

	return parseBufferResponse(resp.Body)
}

// ClearBuffer clears the PLM buffer.
func (hub *Hub2245) ClearBuffer() error {
	_, err := hub.doRequest("/1?XB=M=1")
	return err
}

// SendCommand sends a standard command to a specific Insteon device.
func (hub *Hub2245) SendCommand(hostCmd byte, addr Address, imCmd1 byte, imCmd2 byte) (*CommandResponse, error) {
	plmCmd := buildPlmCommand(hostCmd, addr, imCmd1, imCmd2)
	uri := fmt.Sprintf("/%X?%X=I=%X", cmdTypeFull, plmCmd, cmdTypeFull)
	if _, err := hub.doRequest(uri); err != nil {
		return nil, err
	}
	if err := hub.waitForAck(plmCmd); err != nil {
		return nil, err
	}
	return hub.waitForResponse()
}

// SendGroupCommand sends a command to a group.
func (hub *Hub2245) SendGroupCommand(hostCmd byte, group byte) error {
	uri := fmt.Sprintf("/%X?%02X%02X=I=%X", cmdTypeShort, hostCmd, group, cmdTypeShort)
	_, err := hub.doRequest(uri)
	return err
}

func (hub *Hub2245) waitForAck(cmd []byte) error {
	for i := 0; i < 5; i++ {
		buf, err := hub.GetBuffer()
		if err != nil {
			return err
		}
		idx := bytes.Index(buf, cmd)
		if idx >= 0 && idx+len(cmd)+1 <= len(buf) {
			if buf[idx+len(cmd)] == serialACK {
				return nil
			} else if buf[idx+len(cmd)] == serialNAK {
				return fmt.Errorf("device not ready for commands")
			}
		}
	}
	return fmt.Errorf("timeout waiting for device")
}

func (hub *Hub2245) waitForResponse() (*CommandResponse, error) {
	for i := 0; i < 5; i++ {
		buf, err := hub.GetBuffer()
		if err != nil {
			return nil, err
		}
		idx := bytes.Index(buf, []byte{0x02, 0x50})
		if idx >= 0 && idx+11 <= len(buf) {
			rsp := &CommandResponse{}
			rsp.fromBytes(buf[idx : idx+11])
			return rsp, nil
		}
	}
	return nil, fmt.Errorf("timeout waiting for device")
}

// buildPlmCommand builds a power line modem command that gets sent by the
// Insteon hub.
func buildPlmCommand(hostCmd byte, addr Address, imCmd1, imCmd2 byte) []byte {
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
