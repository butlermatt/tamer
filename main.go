package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"bufio"
	"bytes"
	"strconv"
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
}


const (
	// Various indexes of data
	msgType = iota // Message type
	tType 		  // Transmission type. MSG type only
	_			  // Session Id. Don't care
	_			  // Aircraft ID. Don't care (usually 11111)
	icao			  // ModeS or ICAO Hex number
	_			  // Flight ID. Don't care (usually 11111)
	dGen			  // Date message was Generated
	tGen			  // Time message was Generated
	dLog			  // Date message was logged.
	tLog			  // Time message was logged.

	// May not be in every message
	callSign		  // Call Sign (Flight ID, Flight Number or Registration)
	alt			  // Altitude
	groundSpeed    // Ground Speed (not indicative of air speed)
	track		  // Track of aircraft, not heading. Derived from Velocity E/w and Velocity N/S
	latitude		  // As it says
	longitude	  // As it says
	verticalRate   // 64ft resolution
	squawk		  // Assigned Mode A squawk code
	squawkAlert    // Flag to indicate squawk change.
	emergency	  // Flag to indicate Emergency
	spi			  // Flag to indicate transponder Ident has been activated
	onGround		  // Flag to indicate ground squawk switch is active
)

func parseMessage(m []byte) {
	parts := bytes.Split(m, []byte{','})
	if len(parts) != 22 {
		fmt.Fprintf(os.Stderr, "Discarding bad message: %q", m)
		return
	}

	modesHex := string(parts[icao])
	if modesHex == "000000" {
		fmt.Println("Discarding message with empty ICAO")
		return
	}

	mtype := string(parts[msgType])
	if mtype != "MSG" {
		// TODO I'm not ready to handle this yet.
		fmt.Println("Unable to handle message of type %q\n", mtype)
		return
	}

	ttype, err := strconv.Atoi(string(parts[tType]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting transmission type (only on msg fields?). %q, error: %v", parts[tType], err)
	}
	fmt.Printf("Received message of type: %v - %d for plane: %v\n", mtype, ttype, modesHex)

}