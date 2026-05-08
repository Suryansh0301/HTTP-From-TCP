package main

import (
	"bufio"
	"http-from-tcp/enums"
	"log"
	"net"
	"os"
)

func main() {
	raddr, err := net.ResolveUDPAddr(enums.NetworkUDP.String(), "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP(enums.NetworkUDP.String(), nil, raddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)
	log.Println(">")

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Fatal(err)
		}
	}
}
