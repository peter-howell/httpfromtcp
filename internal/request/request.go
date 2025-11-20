// Package request implements parsing and processing of HTTP requests
package request

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/peter-howell/httpfromtcp/internal/headers"
)

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

func (r *RequestLine) String() string {
	return fmt.Sprintf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s", r.Method, r.RequestTarget, r.HttpVersion)
}

type parserState string
const (
	StateInit parserState = "init"
	StateParseHeaders parserState = "parsingHeaders"
	StateParseBody parserState = "parsingBody"
	StateDone parserState = "done"
)

type Request struct {
	RequestLine RequestLine
	Headers headers.Headers
	Body []byte

	state parserState
}

func (r* Request) String() string {
	return fmt.Sprintf("%s\n%s\nBody:\n%s", &r.RequestLine, r.Headers, r.Body)
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
		Headers: headers.NewHeaders(),
		Body: make([]byte, 0),
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
	// fmt.Printf("%s\n", string(data))
outer:
	for {
		if len(data) == 0 {
			break outer
		}
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
			r.state = StateParseHeaders
		case StateParseHeaders:
			n, done, err := r.Headers.Parse(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if done {
				r.state = StateParseBody
			}
		case StateParseBody:
			expecLenS, ok := r.Headers.Get("content-length")
			if !ok {
				r.state = StateDone
				break outer
			}
			expecLen, err := strconv.Atoi(expecLenS)
			if err != nil {
				return 0, fmt.Errorf("unknown content-length value: %v", expecLenS)
			}
			if expecLen == 0 {
				r.state = StateDone
				break outer
			}

			r.Body = append(r.Body, data[read:]...)

			read += len(data[read:])
			if len(r.Body) > expecLen {
				return 0, fmt.Errorf("too many bytes in body")
			}

			if len(r.Body) == expecLen {
				r.state = StateDone
			}
			return read, nil
		case StateDone:
			break outer
		default:
			panic("uh oh")
		}
	}
	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()
	buf := make([]byte, 1024)
	bufLen := 0
	for !req.done() {
		if bufLen >= len(buf) {
			newBuf := make([]byte, 2*len(buf))
			copy(newBuf, buf)
			buf = newBuf

		}
		nRead, err := reader.Read(buf[bufLen:])
		//todo what to do with errors?
		if err != nil {
			return nil, err
		}
		bufLen += nRead
		fmt.Printf("buff: %s\n", buf[:bufLen])
		nParsed, err := req.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}
		copy(buf, buf[nParsed:bufLen])
		bufLen -= nParsed
	}
	return req, nil
}

