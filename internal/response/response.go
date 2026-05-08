package response

import (
	"bytes"
	"errors"
	"fmt"
	"http-from-tcp/constants"
	"http-from-tcp/enums"
	"http-from-tcp/internal/headers"
	"io"
	"strconv"
)

// GetDefaultHeaders returns a standard set of HTTP response headers.
// It sets Content-Length to the provided contentLen, Connection to "close",
// and Content-Type to plain text.
func (w *Writer) GetDefaultHeaders(contentLen int) headers.Headers {
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

// WriteStatusLine writes the HTTP/1.1 status line for the given status code.
// For example: "HTTP/1.1 200 OK\r\n"
// Returns an error if the status code is not supported.
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
		return errors.New("invalid error code")
	}

	_, err := w.writer.Write(statusLine)
	if err != nil {
		return err
	}

	return nil
}

// WriteHeaders writes all HTTP headers followed by the terminal CRLF that
// separates headers from the response body.
// Each header is written as "Key: Value\r\n", ending with an extra "\r\n".
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

// WriteBody writes raw bytes directly to the underlying writer.
// Returns the number of bytes written and any error encountered.
func (w *Writer) WriteBody(p []byte) (int, error) {
	n, err := w.writer.Write(p)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// WriteChunkedBody writes a single chunk in HTTP chunked transfer encoding format.
// Each chunk is written as:
//
//	<hex size>\r\n
//	<data>\r\n
//
// Null bytes are trimmed from p before writing to handle fixed-size buffers
// that may not be fully filled.
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

// WriteChunkedBodyDone writes the terminal chunk that signals the end of a
// chunked transfer encoded response body.
// The terminal chunk is always "0\r\n" per the HTTP/1.1 spec.
func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.WriteBody([]byte("0\r\n"))
	if err != nil {
		return 0, err
	}
	return n, nil
}

// WriteTrailers writes trailing headers after the final chunk in a chunked
// transfer encoded response. Trailers follow the same "Key: Value\r\n" format
// as regular headers and must be declared in the Trailer header beforehand.
func (w *Writer) WriteTrailers(h headers.Headers) error {
	return w.WriteHeaders(h)
}
