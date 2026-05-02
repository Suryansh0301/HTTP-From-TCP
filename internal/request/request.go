package request

import (
	"fmt"
	"io"
	"strings"
)

var (
	SEPERATOR              string = "\r\n"
	MALFORMED_REQUEST_LINE error  = fmt.Errorf("malformed request-line")
)

type parsedState string

const (
	DoneParsedState parsedState = "done"
	InitParsedState parsedState = "initialized"
)

type Request struct {
	RequestLine RequestLine
	State       parsedState
}

func (r *Request) Done() bool {
	return r.State == DoneParsedState
}

func newRequest() *Request {
	return &Request{
		State: InitParsedState,
	}
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *RequestLine) Valid() bool {
	return strings.Compare(r.Method, strings.ToUpper(r.Method)) == 0 && r.HttpVersion == "1.1"
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

outer:
	for {
		switch r.State {
		case InitParsedState:
			rl, n, err := parseRequestLine(string(data[read:]))
			if err != nil {
				return 0, err
			}

			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.State = DoneParsedState
		case DoneParsedState:
			break outer
		}
	}

	return read, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buff := make([]byte, 8)
	buffLen := 0

	for !request.Done() {
		n, err := reader.Read(buff[buffLen:])
		if err != nil {
			return nil, err
		}

		buffLen += n
		readN, err := request.parse(buff[:buffLen])
		if err != nil {
			return nil, err
		}

		copy(buff, buff[readN:buffLen])
		buffLen -= readN
	}

	return request, nil
}

func parseRequestLine(s string) (*RequestLine, int, error) {
	index := strings.Index(s, SEPERATOR)
	if index == -1 {
		return nil, 0, nil
	}

	startLine := s[:index]

	parts := strings.Split(startLine, " ")
	if len(parts) != 3 {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	reqLine := &RequestLine{
		HttpVersion:   httpParts[1],
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	if !reqLine.Valid() {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	return reqLine, index + len(SEPERATOR), nil
}
