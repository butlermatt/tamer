package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	savePeriod time.Duration = time.Minute * 1
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

	msgs := make(chan *message, 50)
	cmds := make(chan BoardCmd)
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	if verbose {
		fmt.Println("Initalizing databases")
	}
	err := init_db()
	if err != nil {
		fmt.Fprintf(os.Stderr, "errors setting up database: %v", err)
		os.Exit(1)
	}

	json := StartServer(cmds)
	tick := time.NewTicker(savePeriod)

	go connect(msgs)

	for {
		select {
		case m := <-msgs:
			updatePlane(m)
		case cmd := <-cmds:
			if cmd == ListAll {
				json <- dumpJson()
			}
		case t := <-tick.C:
			saveData(t)
		case <-sigint:
			saveData(time.Time{})
			err = close_db()
			tick.Stop()
			close(cmds)
			cmds = nil
			close(json)
			close(msgs)
			msgs = nil
			os.Exit(0)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing database: %#v", err)
	}
	os.Exit(0)
}

func saveData(t time.Time) {
	if len(planeCache) == 0 {
		return
	}

	var toSave []*Plane

	var zero time.Time
	if t == zero { // Save all planes if time 0. Probably an interrupt received.
		i := 0
		toSave = make([]*Plane, len(planeCache))
		for icao, pl := range planeCache {
			toSave[i] = pl
			delete(planeCache, icao)
			i++
		}
	} else {
		period := t.Add(-savePeriod)

		for icao, pl := range planeCache {
			if period.After(pl.LastSeen) {
				toSave = append(toSave, pl)
				delete(planeCache, icao)
			}
		}
	}

	if len(toSave) == 0 {
		return
	}

	err := SavePlanes(toSave)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error saving planes to database: %v\n", err)
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
	if m == nil {
		return
	}

	pl, ok := planeCache[m.icao]
	var err error
	if !ok {
		pl, err = LoadPlane(m.icao)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading from database: %v\n", err)
			pl = &Plane{Icao: m.icao}
		}
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
