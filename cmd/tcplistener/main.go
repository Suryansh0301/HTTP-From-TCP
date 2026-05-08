package main

import (
	"http-from-tcp/enums"
	"http-from-tcp/internal/request"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen(enums.NetworkTCP.String(), "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("started listening on port 42069")
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go func(c net.Conn) {
			log.Println("A connection has been accepted!")

			defer log.Println("The connection has been closed!")
			defer c.Close()

			for {
				req, err := request.RequestFromReader(c)
				if err != nil {
					log.Fatal(err)
				}

				req.Log()
			}
		}(connection)
	}
}
