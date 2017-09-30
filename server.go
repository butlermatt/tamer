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
)


type Server struct {
	json chan string
	cmd  chan<- *BoardCmd
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request: ", r.URL)

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
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "invlaid ICAO number: %q", parts[1])
			fmt.Fprintf(os.Stderr, "Bad request URL: %q", r.URL.Path)
			return
		}
	}

	bc := &BoardCmd{Icao: uint(icao)}

	switch reqCmd {
	case "":
	case "active":
		bc.Cmd = GetCurrent
	case "planes":
		if icao > 0 {
			bc.Cmd = GetPlane
		} else {
			bc.Cmd = GetAll
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unkown command: %q", reqCmd)
		fmt.Fprintf(os.Stderr, "Bad request URL: %q", r.URL.Path)
		return
	}
	s.cmd <- bc
	s.writeResponse(<-s.json, w)
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


func currentPlanes() string {
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

func getAllPlanes() string {
	buf := bytes.Buffer{}

	buf.WriteString("{current: ")
	buf.WriteString(currentPlanes())

	buf.WriteString(",\npast: [")

	planes, err := LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading planes: %v\n", err)
		return "[]"
	}
	sl := make([]string, len(planes))
	for i, pl := range planes {
		sl[i] = pl.ToJson()
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

	LoadLocations(pl)
	var buf bytes.Buffer
	buf.WriteString("{plane: ")
	buf.WriteString(pl.ToJson())
	buf.WriteString(",\nlocations: [")

	locs := make([]string, len(pl.Locations))
	i := 0
	for _, l := range pl.Locations {
		locs[i] = fmt.Sprintf("{location: \"%s,%s\", time: %q}", l.Latitude, l.Longitude, l.Time.String())
		i++
	}
	buf.WriteString(strings.Join(locs, ","))
	buf.WriteString("]}")
	return buf.String()
}