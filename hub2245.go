package insteon

import (
	"context"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// Hub2245 is a reference to an Insteon Hub.
type Hub2245 struct {
	address  string
	userName string
	password string
	*HubStreaming

	bufferTimer   *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	closeWg       sync.WaitGroup
	outQueueWrite *io.PipeWriter
	outQueueRead  *io.PipeReader
}

// NewHub2245 creates a new reference to an Insteon Hub2. This hub is a little different from the Hub1 and the Serial
// PLM in that it has an HTTP interface. The interface is a little unfortunate though since you lose bi-directional real
// time communication, so you end up having to poll for events which makes interfacing to this modem a bit slower.
func NewHub2245(address string, userName string, password string) (Hub, error) {
	h := &Hub2245{
		address:  address,
		userName: userName,
		password: password,
	}

	h.outQueueRead, h.outQueueWrite = io.Pipe()

	h.ctx, h.cancel = context.WithCancel(context.Background())

	if err := h.clearBuffer(h.ctx); err != nil {
		return nil, err
	}

	streamHub, err := NewHubStreaming(h)
	if err != nil {
		return nil, err
	}
	h.HubStreaming = streamHub

	return h, nil
}

func (hub *Hub2245) startBufferTicker() {
	hub.bufferTimer = time.NewTicker(500 * time.Millisecond)
	hub.closeWg.Add(1)

	go func() {
		defer hub.closeWg.Done()
		defer hub.bufferTimer.Stop()

		for {
			select {
			case <-hub.bufferTimer.C:
				hub.readBuffer()
			case <-hub.ctx.Done():
				return
			}
		}
	}()
}

func (hub *Hub2245) readBuffer() error {
	buf, err := hub.getBuffer(hub.ctx)
	if err != nil {
		return err
	} else if len(buf) == 0 {
		return nil
	}

	if _, err := hub.outQueueWrite.Write(buf); err != nil {
		return err
	}

	// We read what's in the buffer, so we need to clear.
	if err := hub.clearBuffer(hub.ctx); err != nil {
		return err
	}

	return nil
}

func (hub *Hub2245) Read(p []byte) (n int, err error) {
	if hub.bufferTimer == nil {
		hub.startBufferTicker()
	}

	return hub.outQueueRead.Read(p)
}

func (hub *Hub2245) Write(p []byte) (n int, err error) {
	uri := fmt.Sprintf("/%X?%s=I=%X", cmdTypeFull, hex.EncodeToString(p), cmdTypeFull)

	rsp, err := hub.doRequest(hub.ctx, uri)
	if err != nil {
		return 0, err
	}
	defer rsp.Body.Close()

	return len(p), nil
}

func (hub *Hub2245) Close() error {
	hub.cancel()
	hub.closeWg.Wait()

	if err := hub.outQueueWrite.Close(); err != nil {
		return err
	}

	if err := hub.outQueueRead.Close(); err != nil {
		return err
	}

	return nil
}

// getBuffer returns the current PLM buffer from the Insteon hub.
func (hub *Hub2245) getBuffer(ctx context.Context) ([]byte, error) {
	resp, err := hub.doRequest(ctx, "/buffstatus.xml")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	buf, err := parseBufferResponse(resp.Body)
	if err != nil {
		return nil, err
	}

	// The last byte appears to be the length of the buffer in nibbles.
	buf = buf[0:(buf[100] / 2)]

	return buf, nil
}

// clearBuffer clears the PLM buffer.
func (hub *Hub2245) clearBuffer(ctx context.Context) error {
	resp, err := hub.doRequest(ctx, "/1?XB=M=1")
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return err
}

// doRequest submits a request to the Insteon hub.
func (hub *Hub2245) doRequest(ctx context.Context, uri string) (*http.Response, error) {
	log.Printf("post to " + hub.address + uri)
	req, err := http.NewRequestWithContext(ctx, "POST", hub.address+uri, nil)
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

// bufResponse wraps the XML response from buffstatus.xml.
type bufResponse struct {
	Buffer string `xml:"BS"`
}

// parseBufferResponse reads the buffstatus.xml from the hub.
func parseBufferResponse(body io.Reader) ([]byte, error) {
	dec := xml.NewDecoder(body)
	buf := &bufResponse{}

	if err := dec.Decode(buf); err != nil {
		return nil, err
	}

	bufBytes, err := hex.DecodeString(buf.Buffer)
	if err != nil {
		return nil, err
	}

	return bufBytes, nil
}
