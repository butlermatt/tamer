package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"bytes"
	"time"
)

type BoardCmd struct {
	Cmd   int
	Icao  uint
	Since time.Time
}

const (
	GetCurrent = iota
	GetAll
	GetPlane
	GetLocations
)

var zeroTime = time.Time{}

type Server struct {
	json chan string
	cmd  chan<- *BoardCmd
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	parts := strings.Split(r.URL.Path, "/")[1:]
	reqCmd := strings.ToLower(parts[0])
	var icao uint64
	var err error
	if len(parts) == 2 {
		icao, err = strconv.ParseUint(parts[1], 16, 0)
		if err != nil {
			s.badRequest(w, http.StatusBadRequest, fmt.Sprintf("invalid ICAO number: %q", parts[1]), r.URL.Path)
			return
		}
	}

	bc := &BoardCmd{Icao: uint(icao)}

	ss := r.URL.Query()["s"]
	if len(ss) >= 1 {
		si, err := strconv.ParseInt(ss[0], 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to convert since time: %q. %#v\n", ss[0], err)
		}
		bc.Since = time.Unix(si, 0)
	}

	switch reqCmd {
	case "active":
		bc.Cmd = GetCurrent
	case "planes":
		if icao > 0 {
			bc.Cmd = GetPlane
		} else {
			bc.Cmd = GetAll
		}
	case "locations":
		if icao == 0 {
			s.badRequest(w, http.StatusBadRequest, "missing required plane icao number", r.URL.Path)
			return
		}
		bc.Cmd = GetLocations
	default:
		http.ServeFile(w, r, "www" + r.URL.Path)
		return
	}
	s.cmd <- bc
	s.writeResponse(<-s.json, w)
}

func (s *Server) badRequest(w http.ResponseWriter, status int, errMsg string, path string) {
	w.WriteHeader(status)
	fmt.Fprint(w, errMsg)
	fmt.Fprintf(os.Stderr, "Bad request URL: %q", path)
}

func (s *Server) writeResponse(resp string, w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "application/json")
	buf := bufio.NewWriter(w)
	_, err := buf.WriteString(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing response: %v", err)
	}

	err = buf.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing response: %v", err)
	}
}

var server *Server

func StartServer(c chan<- *BoardCmd) chan<- string {
	if server == nil || server.json == nil {
		server = &Server{cmd: c}
		server.json = make(chan string)
	}

	addr := fmt.Sprintf(":%d", port)
	fmt.Println("Starting server")
	go func() {
		err := http.ListenAndServe(addr, server)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting HTTP server: %v", err)
			close(server.json)
			close(server.cmd)
			return
		}
	}()

	return server.json
}


func currentPlanes(t time.Time) string {
	buf := bytes.Buffer{}

	buf.WriteString("[")

	sl := []string{}
	for _, pl := range planeCache {
		if t == zeroTime || pl.LastSeen.After(t) {
			sl = append(sl, pl.ToJson())
		}
	}

	buf.WriteString(strings.Join(sl, ",\n"))

	buf.WriteString("]")
	return buf.String()
}

func getAllPlanes(t time.Time) string {
	buf := bytes.Buffer{}

	buf.WriteString("{\"current\": ")
	buf.WriteString(currentPlanes(t))

	buf.WriteString(",\n\"past\": [")

	planes, err := LoadAll(t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading planes: %v\n", err)
		return "[]"
	}
	sl := []string{}
	for _, pl := range planes {
		if t == zeroTime || pl.LastSeen.After(t) {
			sl = append(sl, pl.ToJson())
		}
	}

	buf.WriteString(strings.Join(sl, ",\n"))
	buf.WriteString("] }")
	return buf.String()
}

func detailedPlane(icao uint) string {
	pl, err := getPlaneByIcao(icao)
	if err != nil {
		return ""
	}

	return pl.ToJson()
}

func getPlaneLocations(icao uint, t time.Time) string {
	locs, err := LoadLocations(icao, t)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading locations: %#v", err)
	}

	pl := planeCache[icao]
	if pl != nil {
		locs = append(locs, pl.Locations...)
	}

	if len(locs) <= 0 {
		return "[]"
	}

	ll := make([]string, len(locs))
	buf := bytes.Buffer{}
	buf.WriteString("[")

	for i, l := range locs {
		ll[i] = fmt.Sprintf("{\"id\": %d, \"latitude\": %f, \"longitude\": %f, \"time\": %q}", l.id, l.Latitude, l.Longitude, l.Time.String())
	}

	buf.WriteString(strings.Join(ll, ",\n"))
	buf.WriteString("]")
	return buf.String()
}