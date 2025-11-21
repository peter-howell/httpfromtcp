// Package response provides code to make HTTP responses
package response

import (
	"fmt"
	"io"

	"github.com/peter-howell/httpfromtcp/internal/headers"
)


type writerState int 
const (
	wStateStatusLine writerState = iota
	wStateHeaders
	wStateBody
)

type Writer struct {
	wState writerState
	writer io.Writer
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



func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		wState: wStateStatusLine,
		writer: conn,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.wState != wStateStatusLine  {
		return fmt.Errorf("status line is not needed based on current state")
	}
	defer func() {w.wState = wStateHeaders}()
	err := WriteStatusLine(w.writer, statusCode)

	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.wState != wStateHeaders {
		return fmt.Errorf("headers aren't needed based on current state")
	}
	defer func() {w.wState = wStateBody}()
	return WriteHeaders(w.writer, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.wState != wStateBody {
		return 0, fmt.Errorf("body isn't needed based on current state")
	}
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {

	return 0, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {

	return 0, nil
}

