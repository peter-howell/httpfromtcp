// Package headers parses headers (aka field lines) in HTTP requests
package headers

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Headers map[string]string

func (h Headers) String() string {
	if len(h) == 0 {
		return ""
	}
	s := "Headers:"
	for key, val := range h {
		s += fmt.Sprintf("\n- %s: %s", key, val)
	}
	return s
}

func (h Headers) Write(w io.Writer) error {
	if len(h) == 0 {
		return nil
	}
	for key, val := range h {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, val)
		if err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\r\n")

	return err
}



var REQ_LINE_SEP = []byte("\r\n")
var FIELD_SEP = []byte(":")

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(key string) (string, bool) {
	val, ok := h[strings.ToLower(key)]
	return val, ok
}



func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	//allLetter := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
	idx := bytes.Index(data, REQ_LINE_SEP)
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	line := data[:idx] // get data on this line, before the \r\n
	lineTrim := bytes.TrimSpace(line)
	// separate field name and field value
	idx = bytes.Index(lineTrim, FIELD_SEP)
	if idx == -1 {
		return 0, false, fmt.Errorf("no colon found separating field name and value")
	}
	if idx == 0 {
		return 0, false, fmt.Errorf("no field name given before colon")
	}
	if idx == len(lineTrim) {
		return 0, false, fmt.Errorf("there is no value for the field, the colon was at the end of the line")
	}
	fieldName := bytes.ToLower(lineTrim[:idx])
	if len(fieldName) != len(bytes.TrimSpace(fieldName)) {
		return 0, false, fmt.Errorf("there should be no whitespace between the field name and ':', or there must be a field name")
	}
	if !validTokens(fieldName) {
		return 0, false, fmt.Errorf("invalid field name (header token) %s", fieldName)
	}
	
	fieldValue := bytes.TrimSpace(lineTrim[idx+len(FIELD_SEP):])

	h.Set(string(fieldName), string(fieldValue))

	return len(line) + len(REQ_LINE_SEP), false, nil
}

func (h Headers) Set(key, val string) {
	key = strings.ToLower(key)
	oldVal, ok := h[key]
	if ok {
		h[key] = fmt.Sprintf("%s, %s", oldVal, val)
	} else {
		h[key] = val
	}
}

func (h Headers) Replace(key, val string) {
	h[strings.ToLower(key)] = val
}

func validTokens(data []byte) bool {
	for _, c := range data {
		if ((c < 'A' || c > 'Z') &&
			(c < 'a' || c > 'z') &&
			(c < '0' || c > '9') &&
			c != '-') {
					return false
				}
		}
		return true
}




