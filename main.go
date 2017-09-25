package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
)

var (
	addr string
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30003", "Address and port to connect to for input.")
}

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect. %v", err)
		os.Exit(1)
	}
	fmt.Println("Connected")

	reader := bufio.NewReader(conn)

	for b, err := reader.ReadBytes('\n'); err == nil; b, err = reader.ReadBytes('\n') {
		fmt.Printf("%s", b)
		go parseMessage(b)
	}

	fmt.Println("Readloop ended. Maybe closed?")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stream: %v", err)
	}
}

