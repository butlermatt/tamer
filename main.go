package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
	"bytes"
)

var (
	addr string
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30003", "Address and port to connect to for input.")
}

// TODO: Don't change a value (or add squawk etc) if it already exists with that value.
// Don't add that to the history. However add new changes. Always update LastSeen if after
type Plane struct {
	Icao		  uint
	CallSigns []string
	Squawks   []string
	Locations []Location
	Altitude  int
	Track     float32
	Speed     float32
	Vertical  int
	LastSeen  time.Time
	History   []*message // won't contain duplicate messages such as "on ground" unless they change
	// Various flags
	SquawkCh  bool
	Emergency bool
	Ident     bool
	OnGround  bool
}

func (p *Plane) WriteCallSign(cs string) bool {
	var found bool
	for _, c := range p.CallSigns {
		if cs == c {
			found = true
			break
		}
	}

	if !found {
		p.CallSigns = append(p.CallSigns, cs)
	}
	return found
}

func (p *Plane) WriteHistory(m *message) {
	p.History = append(p.History, m)
}

type Location struct {
	Time      time.Time
	Latitude  string
	Longitude string
}

func main() {
	flag.Parse()

	msgs := make(chan *planeMsg)

	go connect(msgs)

	planeCache := make(map[uint]*Plane)

	buf := bytes.Buffer{}
	for p := range msgs {

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
		buf.WriteString(fmt.Sprintf("%X", p.icoa))
		buf.WriteString("\" At: ")
		buf.WriteString(p.msg.dRec.String())

		var dataStr string
		switch p.msg.tType {
		case 1:
			if pl.WriteCallSign(p.msg.callSign) {
				pl.WriteHistory(p.msg)
			}
			dataStr = fmt.Sprintf(" callsign: %q", p.msg.callSign)
		case 2:
			dataStr = fmt.Sprintf(" Altitude: %d, Speed: %.2f, Track: %.2f, Lat: %s, Lon: %s", p.msg.altitude, p.msg.groundSpeed, p.msg.track, p.msg.latitude, p.msg.longitude)
		case 3:
			dataStr = fmt.Sprintf(" Altitude: %d, Lat: %s, Lon: %s", p.msg.altitude, p.msg.latitude, p.msg.longitude)
		case 4:
			dataStr = fmt.Sprintf(" Speed: %.2f, Track: %.2f, Vertical Rate: %d", p.msg.groundSpeed, p.msg.track, p.msg.vertical)
		case 5:
			dataStr = fmt.Sprintf(" Altitude: %d", p.msg.altitude)
		case 6:
			dataStr = fmt.Sprintf(" Altitude: %d, SquawkCode: %q", p.msg.altitude, p.msg.squawk)
		case 7:
			dataStr = fmt.Sprintf(" Altitude: %d", p.msg.altitude)
		case 8:
			dataStr = fmt.Sprintf(" OnGround: %v", p.msg.onGround)
		}

		buf.WriteString(dataStr)

		fmt.Println(buf.String())
		buf.Reset()
	}

}

func connect(out chan<- *planeMsg) {
	i := 1;
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