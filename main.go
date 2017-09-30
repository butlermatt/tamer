package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
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
			pl, _ := getPlaneByIcao(m.icao)
			updatePlane(m, pl)
		case cmd := <-cmds:
			json <- handleCommand(cmd)
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

func getPlaneByIcao(icao uint) (*Plane, error) {
	pl, ok := planeCache[icao]
	var err error
	if !ok {
		pl, err = LoadPlane(icao)
		if err != nil && err != planeNotFound {
			fmt.Fprintf(os.Stderr, "error loading from database: %v\n", err)
			pl = &Plane{Icao: icao}
		}
		planeCache[pl.Icao] = pl
		if err != nil {
			return pl, planeNotFound
		}
	}

	return pl, nil
}

func handleCommand(cmd BoardCmd) string {
	switch cmd {
	case ListCurrent:
		return currentPlanes()
	case ListAll:
		return getAllPlanes()
	}

	if cmd < BoardCmd(10) {
		return "" // Empty for unknown command right now.
	}

	icao := uint(cmd)
	return detailedPlane(icao)
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
