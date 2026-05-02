package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) chan string {
	channel := make(chan string, 1)
	scanner := bufio.NewScanner(f)

	go func() {
		for scanner.Scan() {
			channel <- string(scanner.Bytes())
		}
		defer close(channel)
	}()

	return channel

}

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
			defer c.Close()
			lines := getLinesChannel(c)
			for line := range lines {
				fmt.Printf("read: %s\n", line)
			}
			fmt.Println("The connection has been closed!")
		}(connection)

	}

}
