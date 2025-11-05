// Package headers parses headers (aka field lines) in HTTP requests
package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string



var REQ_LINE_SEP = []byte("\r\n")
var FIELD_SEP = []byte(":")

func NewHeaders() Headers {
	return make(map[string]string)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	//allLetter := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString
	idx := bytes.Index(data, REQ_LINE_SEP)
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 0, true, nil
	}
	line := data[:idx] // get data on this line, before the \r\n
	lineTrim := bytes.TrimSpace(line)
	idx = bytes.Index(lineTrim, FIELD_SEP)
	if idx == -1 {
		return 0, false, fmt.Errorf("no colon found separating field name and value")
	}
	if idx == 0 {
		return 0, true, fmt.Errorf("no field name given before colon")
	}
	fieldName := lineTrim[:idx]
	if len(fieldName) != len(bytes.TrimSpace(fieldName)) {
		return 0, false, fmt.Errorf("there should be no whitespace between the field name and ':', or there must be a field name")
	}
	if idx == len(lineTrim) {
		return 0, false, fmt.Errorf("there is no value for the field, the colon was at the end of the line")
	}
	
	fieldValue := bytes.TrimSpace(lineTrim[idx+len(FIELD_SEP):])

	h[string(fieldName)] = string(fieldValue)

	return len(line) + len(REQ_LINE_SEP), false, nil
}



