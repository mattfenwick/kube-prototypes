package http_tester

import "encoding/json"

type Request struct {
	MessageNumber int
	Message       string
}

type Response struct {
	Request         *Request
	ResponseNumber  int
	ResponseMessage string
}

func (resp *Response) JSONString(indent bool) string {
	var bytes []byte
	var err error
	if indent {
		bytes, err = json.MarshalIndent(resp, "", "  ")
	} else {
		bytes, err = json.Marshal(resp)
	}
	doOrDie(err)
	return string(bytes)
}
