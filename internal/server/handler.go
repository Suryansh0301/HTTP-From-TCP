package server

import (
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
)

type Handler func(w *response.Writer, req *request.Request)
