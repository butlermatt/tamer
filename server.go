package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
)

type BoardCmd int

const (
	ListCurrent BoardCmd = iota
	ListAll
)

type Server struct {
	json chan string
	cmd  chan<- BoardCmd
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request: ", r.URL)

	switch r.URL.Path {
	case "/favicon.ico":
		w.WriteHeader(http.StatusNotFound)
		return
	case "/":
		s.listCurrent(w, r)
	case "/All":
		s.listAll(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found."))
	}
}

func (s *Server) listCurrent(w http.ResponseWriter, r *http.Request) {
	s.cmd <- ListCurrent
	resp := <-s.json

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

func (s *Server) listAll(w http.ResponseWriter, r *http.Request) {
	s.cmd <- ListAll
	resp := <-s.json

	h := w.Header()
	h.Set("Content-Type", "application/json")
	buf := bufio.NewWriter(w)
	_, err := buf.WriteString(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing response: %v", err)
	}

	err = buf.Flush()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error flushing response: %v", err)
	}
}

var server *Server

func StartServer(c chan<- BoardCmd) chan<- string {
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
