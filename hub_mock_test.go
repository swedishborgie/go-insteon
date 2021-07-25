package insteon_test

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/swedishborgie/go-insteon"

	"github.com/google/go-cmp/cmp"
)

type InsteonHubMock struct {
	inBuffer   bytes.Buffer
	outPipeIn  *io.PipeReader
	outPipeOut *io.PipeWriter
	ctx        context.Context
	cancel     context.CancelFunc
	expect     []*insteonHubMockExpect
}

type insteonHubMockExpect struct {
	req []byte
	rsp []byte
}

func newMock() (insteon.Hub, *InsteonHubMock) {
	ctx, cancel := context.WithCancel(context.Background())
	mock := &InsteonHubMock{
		ctx:    ctx,
		cancel: cancel,
	}
	mock.outPipeIn, mock.outPipeOut = io.Pipe()
	hub, _ := insteon.NewHubStreaming(mock)

	return hub, mock
}

func (mock *InsteonHubMock) Expect(conv ...[]byte) {
	if len(conv)%2 > 0 {
		panic("conversation must be a multiple of 2")
	}

	for idx := 0; idx < len(conv); idx += 2 {
		mock.expect = append(mock.expect, &insteonHubMockExpect{req: conv[idx], rsp: conv[idx+1]})
	}
}

func (mock *InsteonHubMock) Read(p []byte) (int, error) {
	return mock.outPipeIn.Read(p)
}

func (mock *InsteonHubMock) Write(p []byte) (int, error) {
	cnt, err := mock.inBuffer.Write(p)
	if err != nil {
		return cnt, err
	}

	if len(mock.expect) > 0 {
		expect := mock.expect[0]
		if mock.inBuffer.Len() >= len(expect.req) {
			// We have a request.
			reqBytes := make([]byte, len(expect.req))
			if _, err := mock.inBuffer.Read(reqBytes); err != nil {
				return cnt, err
			}

			if !cmp.Equal(reqBytes, expect.req) {
				mock.cancel()

				return cnt, fmt.Errorf("request didn't match expectation: %s", cmp.Diff(reqBytes, expect.req))
			}

			// Send the canned response.
			if _, err := mock.outPipeOut.Write(expect.rsp); err != nil {
				return cnt, err
			}

			mock.expect = mock.expect[1:]
		}
	}

	return cnt, nil
}

func (mock *InsteonHubMock) Close() error {
	return nil
}
