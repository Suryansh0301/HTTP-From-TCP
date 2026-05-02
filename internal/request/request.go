package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *RequestLine) Valid() bool {
	return strings.Compare(r.Method, strings.ToUpper(r.Method)) == 0 && r.HttpVersion == "1.1"
}

var SEPERATOR string = "\r\n"
var MALFORMED_REQUEST_LINE error = fmt.Errorf("malformed request-line")

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, _, err := parseRequestLine(string(data))
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *requestLine,
	}, nil
}

func parseRequestLine(s string) (*RequestLine, string, error) {
	startLine, restOfMessage, found := strings.Cut(s, SEPERATOR)
	if !found {
		return nil, "", MALFORMED_REQUEST_LINE
	}

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, restOfMessage, MALFORMED_REQUEST_LINE
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return nil, restOfMessage, MALFORMED_REQUEST_LINE
	}

	reqLine := &RequestLine{
		HttpVersion:   httpParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	if !reqLine.Valid() {
		return nil, restOfMessage, MALFORMED_REQUEST_LINE
	}

	return reqLine, restOfMessage, nil
}
