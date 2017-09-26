package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strconv"
	"time"
)

const (
	// Various indexes of data
	msgType = iota // Message type
	tType          // Transmission type. MSG type only
	_              // Session Id. Don't care
	_              // Aircraft ID. Don't care (usually 11111)
	icao           // ModeS or ICAO Hex number
	_              // Flight ID. Don't care (usually 11111)
	dGen           // Date message was Generated
	tGen           // Time message was Generated
	dLog           // Date message was logged.
	tLog           // Time message was logged.

	// May not be in every message
	callSign     // Call Sign (Flight ID, Flight Number or Registration)
	alt          // Altitude
	groundSpeed  // Ground Speed (not indicative of air speed)
	track        // Track of aircraft, not heading. Derived from Velocity E/w and Velocity N/S
	latitude     // As it says
	longitude    // As it says
	verticalRate // 64ft resolution
	squawk       // Assigned Mode A squawk code
	squawkAlert  // Flag to indicate squawk change.
	emergency    // Flag to indicate Emergency
	identActive  // Flag to indicate transponder Ident has been activated
	onGround     // Flag to indicate ground squawk switch is active
)

func parseMessage(m []byte, out chan<- *planeMsg) {
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

	p, err := parseMsgType(parts, ttype)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error trying to decode message. %v", err)
		return
	}

	out <- p
}

type planeMsg struct {
	icoa uint
	msg  *message
}

type message struct {
	tType       int
	dGen        time.Time
	dRec        time.Time
	callSign    string
	altitude    int
	groundSpeed float32
	track       float32
	latitude    string
	longitude   string
	vertical    int
	squawk      string
	squawkCh    bool
	emergency   bool
	ident       bool
	onGround    bool
}

func parseTime(d string, t string) (time.Time, error) {
	dd, err := time.Parse("2006/01/02", d)
	if err != nil {
		return time.Time{}, err
	}

	tt, err := time.Parse("15:04:05.000", t)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(dd.Year(), dd.Month(), dd.Day(), tt.Hour(), tt.Minute(), tt.Second(), tt.Nanosecond(), time.Local), nil
}

func parseInt(i []byte) int {
	var ii int
	if len(i) <= 0 {
		return ii
	}

	ii, _ = strconv.Atoi(string(i))
	return ii
}

func parseFloat(f []byte) float32 {
	var ff float32
	if len(f) <= 0 {
		return ff
	}

	tmp, _ := strconv.ParseFloat(string(f), 32)
	return float32(tmp)
}

func parseBool(b []byte) bool {
	var bb bool
	if len(b) <= 0 {
		return bb
	}

	bb, _ = strconv.ParseBool(string(b))
	return bb
}

func parseMsgType(msg [][]byte, tt int) (*planeMsg, error) {
	// Based on information from http://woodair.net/sbs/Article/Barebones42_Socket_Data.htm

	sentTime, err := parseTime(string(msg[dGen]), string(msg[tGen]))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse generated time")
	}

	recvTime, err := parseTime(string(msg[dLog]), string(msg[tLog]))
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse received time")
	}

	icaoDec, err := strconv.ParseUint(string(msg[icao]), 16, 0)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse icao hex")
	}

	m := &message{tType: tt, dGen: sentTime, dRec: recvTime}
	p := &planeMsg{icoa: uint(icaoDec), msg: m}

	switch tt {
	case 1:
		m.callSign = string(msg[callSign])
	case 2:
		m.altitude = parseInt(msg[alt])
		m.groundSpeed = parseFloat(msg[groundSpeed])
		m.track = parseFloat(msg[track])
		m.latitude = string(msg[latitude])
		m.longitude = string(msg[longitude])
		m.onGround = parseBool(msg[onGround])
	case 3:
		m.altitude = parseInt(msg[alt])
		m.latitude = string(msg[latitude])
		m.longitude = string(msg[longitude])
		m.squawkCh = parseBool(msg[squawkAlert])
		m.emergency = parseBool(msg[emergency])
		m.ident = parseBool(msg[identActive])
		m.onGround = parseBool(msg[onGround])
	case 4:
		m.groundSpeed = parseFloat(msg[groundSpeed])
		m.track = parseFloat(msg[track])
		m.vertical = parseInt(msg[verticalRate])
	case 5:
		m.altitude = parseInt(msg[alt])
		m.squawkCh = parseBool(msg[squawkAlert])
		m.ident = parseBool(msg[identActive])
		m.onGround = parseBool(msg[onGround])
	case 6:
		m.altitude = parseInt(msg[alt])
		m.squawk = string(msg[squawk])
		m.squawkCh = parseBool(msg[squawkAlert])
		m.emergency = parseBool(msg[emergency])
		m.ident = parseBool(msg[identActive])
		m.onGround = parseBool(msg[onGround])
	case 7:
		m.altitude = parseInt(msg[alt])
		m.onGround = parseBool(msg[onGround])
	case 8:
		m.onGround = parseBool(msg[onGround])
	}

	return p, nil
}
