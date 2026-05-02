package main

import (
	"fmt"
	"http-from-tcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("started listening on port 42069")
	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			fmt.Println("A connection has been accepted!")
			defer fmt.Println("The connection has been closed!")
			defer c.Close()
			for {
				req, err := request.RequestFromReader(c)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Request Line: \n-Method: %s\n-Target: %s\n-Version: %s\n", req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
				fmt.Println("Headers:")
				for key := range req.Headers {
					fmt.Printf("-%s: %s\n", key, req.Headers.Get(key))
				}
			}
		}(connection)

	}

}
