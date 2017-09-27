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
	addr string
	port uint
)

var (
	planeCache = make(map[uint]*Plane)
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30003", "Address and port to connect to for input.")
	flag.UintVar(&port, "p", 8888, "Port to bind output webserver.")
}

func main() {
	flag.Parse()

	msgs := make(chan *planeMsg)
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

func updatePlane(p *planeMsg) {
	buf := bytes.Buffer{}

	pl, ok := planeCache[p.icoa]
	if !ok {
		pl = new(Plane)
		pl.Icao = p.icoa
		planeCache[pl.Icao] = pl
	}

	if p.msg.dGen.After(pl.LastSeen) {
		pl.LastSeen = p.msg.dGen
	}

	buf.WriteString("Received message: Plane: \"")
	buf.WriteString(fmt.Sprintf("%06X", p.icoa))
	buf.WriteString("\" At: ")
	buf.WriteString(p.msg.dRec.String())

	var dataStr string
	var written bool
	switch p.msg.tType {
	case 1:
		written = pl.SetCallSign(p.msg.callSign)
		dataStr = fmt.Sprintf(" callsign: %q", p.msg.callSign)
	case 2:
		written = pl.SetAltitude(p.msg.altitude) || written
		written = pl.SetSpeed(p.msg.groundSpeed) || written
		written = pl.SetTrack(p.msg.track) || written
		written = pl.SetLocation(p.msg.latitude, p.msg.longitude, p.msg.dGen) || written
		written = pl.SetOnGround(p.msg.onGround) || written

		dataStr = fmt.Sprintf(" Altitude: %d, Speed: %.2f, Track: %.2f, Lat: %s, Lon: %s", p.msg.altitude, p.msg.groundSpeed, p.msg.track, p.msg.latitude, p.msg.longitude)
	case 3:
		written = pl.SetAltitude(p.msg.altitude) || written
		written = pl.SetLocation(p.msg.latitude, p.msg.longitude, p.msg.dGen) || written
		written = pl.SetSquawkCh(p.msg.squawkCh) || written
		written = pl.SetEmergency(p.msg.emergency) || written
		written = pl.SetIdent(p.msg.ident) || written
		written = pl.SetOnGround(p.msg.onGround) || written

		dataStr = fmt.Sprintf(" Altitude: %d, Lat: %s, Lon: %s", p.msg.altitude, p.msg.latitude, p.msg.longitude)
	case 4:
		written = pl.SetSpeed(p.msg.groundSpeed) || written
		written = pl.SetTrack(p.msg.track) || written
		written = pl.SetVertical(p.msg.vertical) || written

		dataStr = fmt.Sprintf(" Speed: %.2f, Track: %.2f, Vertical Rate: %d", p.msg.groundSpeed, p.msg.track, p.msg.vertical)
	case 5:
		written = pl.SetAltitude(p.msg.altitude) || written
		written = pl.SetSquawkCh(p.msg.squawkCh) || written
		written = pl.SetIdent(p.msg.ident) || written
		written = pl.SetOnGround(p.msg.onGround) || written
		dataStr = fmt.Sprintf(" Altitude: %d", p.msg.altitude)
	case 6:
		written = pl.SetAltitude(p.msg.altitude) || written
		written = pl.SetSquawk(p.msg.squawk) || written
		written = pl.SetSquawkCh(p.msg.squawkCh) || written
		written = pl.SetEmergency(p.msg.emergency) || written
		written = pl.SetIdent(p.msg.ident) || written
		written = pl.SetOnGround(p.msg.onGround) || written
		dataStr = fmt.Sprintf(" Altitude: %d, SquawkCode: %q", p.msg.altitude, p.msg.squawk)
	case 7:
		written = pl.SetAltitude(p.msg.altitude) || written
		written = pl.SetOnGround(p.msg.onGround) || written
		dataStr = fmt.Sprintf(" Altitude: %d", p.msg.altitude)
	case 8:
		written = pl.SetOnGround(p.msg.onGround) || written
		dataStr = fmt.Sprintf(" OnGround: %v", p.msg.onGround)
	}
	if written {
		pl.SetHistory(p.msg)
	}

	buf.WriteString(dataStr)

	fmt.Println(buf.String())
	buf.Reset()
}

func connect(out chan<- *planeMsg) {
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
			fmt.Printf("%s", b)
			go parseMessage(b, out)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading connection. %v. Retrying\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Connection closed, reconnecting.")
		}
	}
}
