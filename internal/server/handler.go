package server

import (
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"io"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError
type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}
