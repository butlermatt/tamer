package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

var magicTimestampMLAT = []byte{0xFF, 0x00, 0x4D, 0x4C, 0x41, 0x54}

func main() {
	conn, err := net.Dial("tcp", "192.168.2.19:30005")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error connecting to server\n")
		os.Exit(1);
	}

	reader := bufio.NewReader(conn)
	var buffer []byte
	for {
		msg, _ := reader.ReadBytes(0x1A)

		if buffer == nil {
			buffer = msg
		} else {
			buffer = append(buffer, msg...)
		}

		if msg[0] < 0x31 || msg[0] > 0x34 {
			fmt.Printf("Buffering\n")
			continue // Don't have a valid message code yet.
		}

		msg = buffer
		buffer = nil
		var msgLen int
		switch msg[0] {
		case 0x31:
			msgLen = 11
		case 0x32:
			msgLen = 16
		case 0x33:
			msgLen = 23
		case 0x34:
			continue
		default:
			continue // I dunno?
		}

		if len(msg) != msgLen {
			fmt.Printf("Message length was wrong: %d\n", len(msg))
			continue
		}

		mlat := isMlat(msg[1:7])

		fmt.Printf("Mlat: %t Message: %v\n", mlat, msg)
	}
}

func isMlat(ts []byte) bool {
	if len(ts) != len(magicTimestampMLAT) {
		return false
	}

	for i := range ts {
		if ts[i] != magicTimestampMLAT[i] {
			return false
		}
	}

	return true
}