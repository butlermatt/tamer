package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	port := flag.Uint("p", 8888, "Port to bind output webserver.")
	flag.Parse()

	_ = port

	clientList := flag.Args()
	if len(clientList) < 1 {
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	_, _ = fmt.Fprintln(os.Stderr, "Tamer is a tool for monitoring ADS-B receivers.")
	_, _ = fmt.Fprintln(os.Stderr, "usage: tamer [options] client list")
	flag.PrintDefaults()
	_, _ = fmt.Fprintln(os.Stderr, "\n  Client list is a space separated list of client IPs and Ports in the form of: ip:port ip2:port")
}

// Levels of data:
// Plane (ICAO)
// Flight (All messages received with no more than 30 minutes separating them)
// Message (Specifics received for the flight)
// Station (?? The receiver that picked up the message)

type tamer struct {
	// TODO
}