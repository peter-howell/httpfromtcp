// Package response provides code to make HTTP responses
package response

import (
	"fmt"
	"io"

	"github.com/peter-howell/httpfromtcp/internal/headers"
)


type WriterState string 
const (
	StateNeedStatus WriterState = "need status line"
	StateNeedHeaders WriterState = "need headers"
	StateNeedBody WriterState = "need body"
	StateDone WriterState = "done"
)

type Writer struct {
	State WriterState
	Conn io.Writer
}

type StatusCode int
const (
	StatusOK StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternalServerError StatusCode = 500
)


func GetStatusLine(code StatusCode) ([]byte, error) {
	var line = []byte{}
	switch code {
	case StatusOK:
		line = []byte("HTTP/1.1 200 OK")
	case StatusBadRequest:
		line = []byte("HTTP/1.1 400 Bad Request")
	case StatusInternalServerError:
		line = []byte("HTTP/1.1 500 Internal Server Error")
	default:
		return nil, fmt.Errorf("unknown status code %v", code)
	}
	return line, nil
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	line, err := GetStatusLine(statusCode)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\r\n", line)
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	h := headers.NewHeaders()

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	err := headers.Write(w)
	if err != nil {
		return err
	}
	return err
}


func (ws WriterState) Next() WriterState{
	switch ws {
	case StateNeedStatus:
		return StateNeedHeaders
	case StateNeedHeaders:
		return StateNeedBody
	default:
		return ws
	}
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		State: StateNeedStatus,
		Conn: conn,
	}
}

func (w *Writer) Advance() {
	w.State = w.State.Next()
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.State != StateNeedStatus  {
		return fmt.Errorf("status line is not needed based on current state")
	}
	w.Advance()
	err := WriteStatusLine(w.Conn, statusCode)

	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.State != StateNeedHeaders {
		return fmt.Errorf("headers aren't needed based on current state")
	}
	w.Advance()
	return WriteHeaders(w.Conn, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.State != StateNeedBody {
		return 0, fmt.Errorf("body isn't needed based on current state")
	}
	w.Advance()
	return w.Conn.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {

	return 0, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {

	return 0, nil
}

