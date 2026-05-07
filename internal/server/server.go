package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
)

// close represented by false and open represented by trye
type Server struct {
	listener net.Listener
	state    *atomic.Bool
}

func Serve(port int) (*Server, error) {
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

	go server.listen()
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

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()

		if !s.state.Load() {
			return
		}
		if err != nil {
			//currently sirf error printing daali h
			log.Println("Accept error:", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn io.ReadWriteCloser) {
	defer conn.Close()
	output := []byte(
		"HTTP/1.1 200 OK\r\n" +
			"Content-Type: text/plain\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"Hello World!\n",
	)
	_, err := conn.Write(output)
	if err != nil {
		log.Print(err)
	}

}
