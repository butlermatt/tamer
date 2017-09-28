package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var (
	// Command line flags
	addr        string
	port        uint
	verbose     bool
	veryVerbose bool
)

var (
	planeCache = make(map[uint]*Plane)
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30003", "Address and port to connect to for input.")
	flag.UintVar(&port, "p", 8888, "Port to bind output webserver.")
	flag.BoolVar(&verbose, "v", false, "Enable verbose message logging. This will list contents of received messages.")
	flag.BoolVar(&veryVerbose, "vv", false, "Enable very verbose message logging. This will list raw received messages. Requires verbose flag")
}

func main() {
	flag.Parse()

	msgs := make(chan *message)
	cmds := make(chan BoardCmd)

	json := StartServer(cmds)

	go connect(msgs)

	for {
		select {
		case p := <-msgs:
			updatePlane(p)
		case cmd := <-cmds:
			if cmd == ListAll {
				json <- dumpJson()
			}
		}
	}
}

func dumpJson() string {
	buf := bytes.Buffer{}

	buf.WriteString("[")

	l := len(planeCache)
	sl := make([]string, l)

	l -= 1
	for _, pl := range planeCache {
		sl[l] = pl.ToJson()
		l -= 1
	}

	buf.WriteString(strings.Join(sl, ",\n"))

	buf.WriteString("]")
	return buf.String()
}

func updatePlane(m *message) {
	buf := bytes.Buffer{}

	pl, ok := planeCache[m.icao]
	if !ok {
		pl = new(Plane)
		pl.Icao = m.icao
		planeCache[pl.Icao] = pl
	}

	if m.dGen.After(pl.LastSeen) {
		pl.LastSeen = m.dGen
	}

	if verbose {
		buf.WriteString(fmt.Sprintf("%s - %06X -", m.dGen.String(), m.icao))
	}

	var dataStr string
	var written bool
	switch m.tType {
	case 1:
		written = pl.SetCallSign(m.callSign)
		if verbose {
			dataStr = fmt.Sprintf(" Callsign: %q", m.callSign)
		}
	case 2:
		written = pl.SetAltitude(m.altitude) || written
		written = pl.SetSpeed(m.groundSpeed) || written
		written = pl.SetTrack(m.track) || written
		written = pl.SetLocation(m.latitude, m.longitude, m.dGen) || written
		written = pl.SetOnGround(m.onGround) || written
		if verbose {
			dataStr = fmt.Sprintf(" Altitude: %d, Speed: %.2f, Track: %.2f, Lat: %s, Lon: %s", m.altitude, m.groundSpeed, m.track, m.latitude, m.longitude)
		}
	case 3:
		written = pl.SetAltitude(m.altitude) || written
		written = pl.SetLocation(m.latitude, m.longitude, m.dGen) || written
		written = pl.SetSquawkCh(m.squawkCh) || written
		written = pl.SetEmergency(m.emergency) || written
		written = pl.SetIdent(m.ident) || written
		written = pl.SetOnGround(m.onGround) || written
		if verbose {
			dataStr = fmt.Sprintf(" Altitude: %d, Lat: %s, Lon: %s", m.altitude, m.latitude, m.longitude)
		}
	case 4:
		written = pl.SetSpeed(m.groundSpeed) || written
		written = pl.SetTrack(m.track) || written
		written = pl.SetVertical(m.vertical) || written
		if verbose {
			dataStr = fmt.Sprintf(" Speed: %.2f, Track: %.2f, Vertical Rate: %d", m.groundSpeed, m.track, m.vertical)
		}
	case 5:
		written = pl.SetAltitude(m.altitude) || written
		written = pl.SetSquawkCh(m.squawkCh) || written
		written = pl.SetIdent(m.ident) || written
		written = pl.SetOnGround(m.onGround) || written
		if verbose {
			dataStr = fmt.Sprintf(" Altitude: %d", m.altitude)
		}
	case 6:
		written = pl.SetAltitude(m.altitude) || written
		written = pl.SetSquawk(m.squawk) || written
		written = pl.SetSquawkCh(m.squawkCh) || written
		written = pl.SetEmergency(m.emergency) || written
		written = pl.SetIdent(m.ident) || written
		written = pl.SetOnGround(m.onGround) || written
		if verbose {
			dataStr = fmt.Sprintf(" Altitude: %d, SquawkCode: %q", m.altitude, m.squawk)
		}
	case 7:
		written = pl.SetAltitude(m.altitude) || written
		written = pl.SetOnGround(m.onGround) || written
		if verbose {
			dataStr = fmt.Sprintf(" Altitude: %d", m.altitude)
		}
	case 8:
		written = pl.SetOnGround(m.onGround) || written
		if verbose {
			dataStr = fmt.Sprintf(" OnGround: %v", m.onGround)
		}
	}
	if written {
		pl.SetHistory(m)
	}

	if verbose {
		buf.WriteString(dataStr)

		fmt.Println(buf.String())
		buf.Reset()
	}
}

func connect(out chan<- *message) {
	i := 1
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			dur := time.Millisecond * time.Duration(i) * time.Duration(100)
			fmt.Fprintf(os.Stderr, "Failed to connect. %v. Retrying in %v\n", err, dur)
			time.Sleep(dur)
			i += i
			continue
		}
		i = 1
		fmt.Println("Connected")
		reader := bufio.NewReader(conn)
		for b, err := reader.ReadBytes('\n'); err == nil; b, err = reader.ReadBytes('\n') {
			if verbose && veryVerbose {
				fmt.Printf("%s", b)
			}
			go parseMessage(b, out)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading connection. %v. Retrying\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Connection closed, reconnecting.")
		}
	}
}
