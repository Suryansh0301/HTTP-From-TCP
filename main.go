package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
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
	file, err := os.Open("message.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	lines := getLinesChannel(file)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
	// scanner := bufio.NewScanner(file)
	// for scanner.Scan() {
	// 	fmt.Printf("read: %s\n", scanner.Bytes())
	// }
}
