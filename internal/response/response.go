// Package response provides code to make HTTP responses
package response

import (
	"fmt"
	"io"

	"github.com/peter-howell/httpfromtcp/internal/headers"
)

type StatusCode int
const (
	StatusOK StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var StatusLines = map[StatusCode]string{
	StatusOK: "HTTP/1.1 200 OK",
	StatusBadRequest:"HTTP/1.1 400 Bad Request",
	StatusInternalServerError:"HTTP/1.1 500 Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	line, ok := StatusLines[statusCode]
	if !ok {
		return fmt.Errorf("unknown status code %v", statusCode)
	}
	_, err := fmt.Fprintf(w, "%s\r\n", line)
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
	_, err =  w.Write([]byte("\r\n"))
	return err
}

