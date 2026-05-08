package response

import (
	"fmt"
	"http-from-tcp/internal/headers"
	"io"
	"strconv"
)

type StatusCode string

const (
	StatusCodeOK                  StatusCode = "200"
	StatusCodeBadRequest          StatusCode = "400"
	StatusCodeInternalServerError StatusCode = "500"
)

func (W *Writer) GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}

type Writer struct {
	writer io.Writer
}

func NewWriter(conn io.Writer) *Writer {
	return &Writer{
		writer: conn,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	var statusLine []byte
	switch statusCode {
	case StatusCodeOK:
		statusLine = []byte("HTTP/1.1 200 OK\r\n")
	case StatusCodeBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request\r\n")
	case StatusCodeInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error\r\n")
	default:
		return fmt.Errorf("invalid error code")
	}

	_, err := w.writer.Write(statusLine)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	byteHeaders := make([]byte, 0)
	for headerKey := range headers {
		byteHeaders = fmt.Appendf(byteHeaders, "%s: %s\r\n", headerKey, headers[headerKey])
	}
	byteHeaders = fmt.Append(byteHeaders, "\r\n")

	_, err := w.writer.Write(byteHeaders)
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}
