package response

import (
	"bytes"
	"fmt"
	"http-from-tcp/constants"
	"http-from-tcp/enums"
	"http-from-tcp/internal/headers"
	"io"
	"strconv"
)

func (W *Writer) GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()

	headers.Set(constants.ContentLength, strconv.Itoa(contentLen))
	headers.Set(constants.Connection, "close")
	headers.Set(constants.ContentType, enums.ContentTypePlain.String())
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

func (w *Writer) WriteStatusLine(statusCode enums.StatusCode) error {
	var statusLine []byte
	switch statusCode {
	case enums.StatusCodeOK:
		statusLine = []byte("HTTP/1.1 200 OK" + constants.CRLF)
	case enums.StatusCodeBadRequest:
		statusLine = []byte("HTTP/1.1 400 Bad Request" + constants.CRLF)
	case enums.StatusCodeInternalServerError:
		statusLine = []byte("HTTP/1.1 500 Internal Server Error" + constants.CRLF)
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
	byteHeaders = fmt.Append(byteHeaders, constants.CRLF)

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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	var size int
	n, err := w.WriteBody([]byte(fmt.Sprintf("%x\r\n", len(bytes.TrimRight(p, "\x00")))))
	if err != nil {
		return size, err
	}
	size += n
	n, err = w.WriteBody(p[:len(bytes.TrimRight(p, "\x00"))])
	if err != nil {
		return size, err
	}
	size += n
	n, err = w.WriteBody([]byte(constants.CRLF))
	if err != nil {
		return size, err
	}
	size += n
	return size, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.WriteBody([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	return w.WriteHeaders(h)
}
