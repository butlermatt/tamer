package messages

import (
	"bytes"
	"fmt"
	"time"
)

const msgParts = 22
const icaoLength = 6

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

// Message contains the various fieilds that a message may contain. Not all values
// will be populated and it will depend on the Transmission type.
type Message struct {
	Icao          uint
	Transmission  int
	Generated     time.Time
	Recorded      time.Time
	Callsign      string
	Altitude      int
	Groundspeed   float32
	Track         float32
	Latitude      float32
	Longitude     float32
	VeriticalRate int
	Squawk        string
	SquawkChange  bool
	Emergency     bool
	IdentActive   bool
	Onground      bool
}

// ParseCSV will Parse the bytes of a CVS message format and return a pointer to a Message.
func ParseCSV(csv string) (*Message, error) {
	// Based on information from http://woodair.net/sbs/Article/Barebones42_Socket_Data.htm
	parts = strings.Split(csv, ",")

	if len(parts) != msgParts {
		return nil, fmt.Errorf("invalid message format, wrong number of parts: %s", m)
	}

	if parts[icao] == "000000" {
		return nil, fmt.Errorf("message contains invalid icao")
	}

	if parts[msgType] != "MSG" {
		return nil, fmt.Errof("unable to handle message of type %q", parts[msgType])
	}

	m := new(Message)
	m.Transmission, err := strconv.Atoi(parts[tType])
	if err != nil {
		return nil, fmt.Errorf("error parsing transmission type: %v", err)
	}

	m.Generated, err = parseTime(parts[dGen], parts[tGen])
	if err != nil {
		return nil, fmt.Errorf("error parsing generated DateTime: %v", err)
	}

	m.Recorded, err = parseTime(parts[dLog], parts[tLog])
	if err != nil {
		return nil, fmt.Errorf("error parsing recorded DateTime: %v", err)
	}

	switch m.Transmission {
	case 1:
		m.Callsign = strings.TrimSpace(parts[callSign])
	case 2:
		m.Altitude = parseInt(parts[alt])
		m.Groundspeed = parseFloat32(parts[groundSpeed])
		m.Track = parseFloat32(parts[track])
		m.Latitude = parseFloat32(parts[latitude])
		m.Longitude = parseFloat32(parts[longitude])
		m.Onground = parseBool(parts[onGround])
	case 3:
		m.Altitude = parseInt(parts[alt])
		m.Latitude = parseFloat32(parts[latitude])
		m.Longitude = parseFloat32(parts[longitude])
		m.SquawkChange = parseBool(parts[squawkAlert])
		m.Emergency = parseBool(parts[emergency])
		m.IdentActive = parseBool(parts[identActive])
		m.Onground = parseBool(parts[onGround])
	case 4:
		m.Groundspeed = parseFloat32(parts[groundSpeed])
		m.Track = parseFloat32(parts[track])
		m.VeriticalRate = parseInt(parts[verticalRate])
	case 5:
		m.Altitude = parseInt(parts[alt])
		m.SquawkChange = parseBool(parts[squawkAlert])
		m.IdentActive = parseBool(parts[identActive])
		m.Onground = parseBool(parts[onGround])
	case 6:
		m.Altitude = parseInt(parts[alt])
		m.Squawk = strings.TrimSpace(parts[squawk])
		m.SquawkChange = parseBool(parts[squawkAlert])
		m.Emergency = parseBool(parts[emergency])
		m.IdentActive = parseBool(parts[identActive])
		m.Onground = parseBool(parts[onGround])
	case 7:
		m.Altitude = parseInt(parts[alt])
		m.Onground = parseBool(parts[onGround])
	case 8:
		m.Onground = parseBool(parts[onGround])
	}

	return m, nil
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

func parseInt(s string) int {
	var i int
	if len(s) <= 0 {
		return i
	}

	i, _ = strconv.Atoi(s)
	return i
}

func parseFloat32(s string) float32 {
	if len(s) <= 0 {
		return float32(0)
	}
	f, _ := strconv.ParseFloat(s, 32)
	return float32(f)
}

func parseBool(s string) bool {
	var b bool
	if len(s) <= 0 {
		return b
	}

	b, _ = strconv.ParseBool(s)
	return b
}