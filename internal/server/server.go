package server

import (
	"fmt"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"io"
	"log"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	state    *atomic.Bool
}

func Serve(port int, handleFunc Handler) (*Server, error) {
	list, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	var running atomic.Bool
	running.Store(true)

	server := &Server{
		listener: list,
		state:    &running,
	}

	go server.listen(handleFunc)
	return server, nil
}

func (s *Server) Close() error {
	s.state.Store(false)
	err := s.listener.Close()
	if err != nil {
		log.Println(err)
	}
	return err
}

func (s *Server) listen(handleFunc Handler) {
	for {
		conn, err := s.listener.Accept()

		if !s.state.Load() {
			return
		}
		if err != nil {
			log.Println("Accept error:", err)
		}

		go s.handle(conn, handleFunc)
	}
}

func (s *Server) handle(conn io.ReadWriteCloser, handleFunc Handler) {
	defer conn.Close()

	responseWriter := response.NewWriter(conn)

	req, err := request.RequestFromReader(conn)
	if err != nil {
		responseWriter.WriteStatusLine(response.StatusCodeBadRequest)
		responseWriter.WriteHeaders(responseWriter.GetDefaultHeaders(0))
	}

	handleFunc(responseWriter, req)
}
