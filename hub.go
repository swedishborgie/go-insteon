package insteon

import (
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Hub is a reference to an Insteon Hub
type Hub struct {
	address  string
	userName string
	password string
}

// NewHub creates a new reference to an Insteon hub.
func NewHub(address string, userName string, password string) *Hub {
	return &Hub{
		address:  address,
		userName: userName,
		password: password,
	}
}

// GetBuffer returns the current PLM buffer from the Insteon hub.
func (hub *Hub) GetBuffer() ([]byte, error) {
	resp, err := hub.doRequest("/buffstatus.xml")
	if err != nil {
		return nil, err
	}

	buffer, err := parseBufferResponse(resp.Body)
	if err != nil {
		return nil, err
	}
	return buffer, nil
}

// ClearBuffer clears the PLM buffer.
func (hub *Hub) ClearBuffer() error {
	_, err := hub.doRequest("/1?XB=M=1")
	return err
}

// SendCommand sends a standard command to a specific Insteon device.
func (hub *Hub) SendCommand(hostCmd byte, addr []byte, imCmd1 byte, imCmd2 byte) ([]byte, error) {
	plmCmd := hub.buildPlmCommand(hostCmd, addr, imCmd1, imCmd2)
	uri := fmt.Sprintf("/%X?%X=I=%X", cmdTypeFull, plmCmd, cmdTypeFull)
	_, err := hub.doRequest(uri)
	return plmCmd, err
}

// SendGroupCommand sends a command to a group.
func (hub *Hub) SendGroupCommand(hostCmd byte, group byte) error {
	uri := fmt.Sprintf("/%X?%02X%02X=I=%X", cmdTypeShort, hostCmd, group, cmdTypeShort)
	fmt.Printf("uri: %s\n", uri)
	_, err := hub.doRequest(uri)
	return err
}

// buildPlmCommand builds a power line modem command that gets sent by the
// Insteon hub.
func (hub *Hub) buildPlmCommand(hostCmd byte, addr []byte, imCmd1, imCmd2 byte) []byte {
	return []byte{
		serialStart,
		hostCmd,
		addr[0], addr[1], addr[2],
		0x0f,
		imCmd1, imCmd2,
	}
}

// doRequest submits a request to the Insteon hub.
func (hub *Hub) doRequest(uri string) (*http.Response, error) {
	req, err := http.NewRequest("GET", hub.address+uri, nil)
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
