package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strconv"

	"http-from-tcp/constants"
	"http-from-tcp/enums"
	"http-from-tcp/internal/headers"
)

var (
	MALFORMED_REQUEST_LINE error = errors.New("malformed request-line")
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	State       enums.ParseState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (r *Request) Done() bool {
	return r.State == enums.ParseStateDone
}

func newRequest() *Request {
	return &Request{
		State:   enums.ParseStateInit,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
}

// Valid checks whether the parsed HTTP request line
// contains a supported HTTP method and a valid HTTP version.
//
// Currently supported methods:
//   - GET
//   - POST
//   - PUT
//   - DELETE
//
// The server only accepts HTTP/1.1 requests.
func (r *RequestLine) Valid() bool {
	validMethods := map[string]bool{
		"GET":    true,
		"POST":   true,
		"PUT":    true,
		"DELETE": true,
	}

	return validMethods[r.Method] &&
		r.HttpVersion == "1.1"
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
		currentData := data[read:]
		switch r.State {
		case enums.ParseStateInit:
			rl, n, err := parseRequestLine(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			r.RequestLine = *rl
			read += n
			r.State = enums.ParseStateHeaders

		case enums.ParseStateHeaders:
			n, done, err := r.Headers.Parse(currentData)
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}
			read += n
			if done {
				r.State = enums.ParseStateBody
			}

		case enums.ParseStateBody:
			contentLength := r.Headers.Get(constants.ContentLength)
			if len(contentLength) == 0 {
				r.State = enums.ParseStateDone
				break
			}

			contentLen, err := strconv.Atoi(contentLength)
			if err != nil {
				return 0, fmt.Errorf("invalid content-length: %w", err)
			}
			if contentLen < 0 {
				return 0, fmt.Errorf("invalid negative content-length: %d", contentLen)
			}

			if len(currentData) == 0 {
				break outer
			}

			remaining := min(contentLen-len(r.Body), len(currentData))
			r.Body = append(r.Body, currentData[:remaining]...)
			read += remaining

			slog.Debug("body chunk read", "remaining", remaining, "totalBody", len(r.Body), "contentLength", contentLen)

			if len(r.Body) > contentLen {
				return 0, fmt.Errorf("body exceeds content-length: got %d, expected %d", len(r.Body), contentLen)
			}
			if len(r.Body) == contentLen {
				r.State = enums.ParseStateDone
			}

		case enums.ParseStateDone:
			break outer
		}
	}

	return read, nil
}

func (r *Request) Log() {
	log.Printf("Request Line: \n-Method: %s\n-Target: %s\n-Version: %s\n", r.RequestLine.Method, r.RequestLine.RequestTarget, r.RequestLine.HttpVersion)
	log.Println("Headers:")
	for key := range r.Headers {
		log.Printf("-%s: %s\n", key, r.Headers.Get(key))
	}
	log.Println("Body:")
	log.Println(string(r.Body))
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buff := make([]byte, 1024)
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

func parseRequestLine(s []byte) (*RequestLine, int, error) {
	index := bytes.Index(s, []byte(constants.CRLF))
	if index == -1 {
		return nil, 0, nil
	}

	startLine := s[:index]

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	reqLine := &RequestLine{
		HttpVersion:   string(httpParts[1]),
		RequestTarget: string(parts[1]),
		Method:        string(parts[0]),
	}

	if !reqLine.Valid() {
		return nil, 0, MALFORMED_REQUEST_LINE
	}

	return reqLine, index + len(constants.CRLF), nil
}
