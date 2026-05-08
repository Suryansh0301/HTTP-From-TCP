package server

import (
	"http-from-tcp/enums"
	"http-from-tcp/internal/request"
	"http-from-tcp/internal/response"
	"io"
	"log"
	"net"
	"strconv"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	state    *atomic.Bool
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port int, handleFunc Handler) (server *Server, err error) {
	list, err := net.Listen(enums.NetworkTCP.String(), strconv.Itoa(port))
	if err != nil {
		return
	}

	server = newServer(list)
	go server.listen(handleFunc)
	return
}

func (s *Server) Close() error {
	s.state.Store(false)
	err := s.listener.Close()
	if err != nil {
		log.Println(err)
	}
	return err
}

func newServer(listener net.Listener) *Server {
	var running atomic.Bool
	running.Store(true)

	return &Server{
		listener: listener,
		state:    &running,
	}
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
		responseWriter.WriteStatusLine(enums.StatusCodeBadRequest)
		responseWriter.WriteHeaders(responseWriter.GetDefaultHeaders(0))
	}

	handleFunc(responseWriter, req)
}
