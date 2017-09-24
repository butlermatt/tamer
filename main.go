package main

import (
	"flag"
	"net"
	"os"
	"fmt"
	"bufio"
)

var (
	addr string
)

const (
	// AC Mode
	modeAc byte = '1'
	// Mode S Short Frame
	modeSs byte = '2'
	// Mode S Long Frame
	modeSl byte = '3'
	// Status Data
	modeSd byte = '4'
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30005", "Address and port to connect to for input.")
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

	for data, err := reader.ReadBytes(0x1A); err == nil; data, err = reader.ReadBytes(0x1A) {
		var msgType string
		switch data[0] {
		case modeAc: msgType = "AC"
		case modeSs: msgType = "S-Short"
		case modeSl: msgType = "S-Long"
		case modeSd: msgType = "Data dump?"
		default: msgType = "Continuation?"
		}
		fmt.Printf("%X - Type: %s\n", data, msgType)
	}
}