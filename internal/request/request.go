// Package request implements parsing and processing of HTTP requests
package request

import (
	"bytes"
	"io"
	"fmt"
	"regexp"
)

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

func (r *RequestLine) String() string {
	return fmt.Sprintf("%s %s %s", r.Method, r.RequestTarget, r.HttpVersion)
}

type parserState string
const (
	StateInit parserState = "init"
	StateDone parserState = "done"
)

type Request struct {
	RequestLine RequestLine
	state parserState
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func (r *Request) done() bool {
	return r.state == StateDone
}

var REQ_LINE_SEP = []byte("\r\n")

func parseRequestLine(rawReq []byte) (*RequestLine, int, error) {
	allLetter := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
	idx := bytes.Index(rawReq, REQ_LINE_SEP)
	if idx == -1 {
		return nil, 0, nil
	}

	startLine := rawReq[:idx]
	read := idx + len(REQ_LINE_SEP)

	fields := bytes.Split(startLine, []byte(" "))
	if size := len(fields); size != 3 {
		return nil, 0, fmt.Errorf("bad number of fields. expected 3, got %d", size)
	}

	httpParts := bytes.Split(fields[2], []byte("/"))
	if size := len(httpParts); size != 2 {
		return nil, 0, fmt.Errorf("bad HTTP version")
	}
	httpV := string(httpParts[1])

	if httpV != "1.1" {
		return nil, 0, fmt.Errorf("HTTP version must be '1.1', got '%s'", httpV)
	}
	if string(httpParts[0]) != "HTTP" {
		return nil, 0, fmt.Errorf("HTTP version must be 'HTTP', got '%s'", httpParts[0])
	}

	reqTarget := string(fields[1])

	method := string(fields[0])
	if !allLetter(method) {
		return nil, 0, fmt.Errorf("method must be all letters. got %s", method)
	}

	return &RequestLine{httpV, reqTarget, method}, read, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
	fmt.Printf("%s\n", string(data))
outer:
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.state = StateDone
		case StateDone:
			break outer
		}
	}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	buf := make([]byte, 1024)
	bufLen := 0
	for !req.done() {
		nRead, err := reader.Read(buf[bufLen:])
		//todo what to do with errors?
		if err != nil {
			return nil, err
		}
		bufLen += nRead
		nParsed, err := req.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[nParsed:bufLen])
		bufLen -= nParsed
	}
	return req, nil
}

