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
func ParseCSV(m []byte) (*Message, error) {
	parts = bytes.Split(m, []byte{','})

	if len(parts) != msgParts {
		return nil, fmt.Errorf("invalid message format, wrong number of parts: %s", m)
	}

	 := string(parts[icao])
	if modesHex
}

func validIcao(icao []byte) bool {
	const byte zero = '0'
	if len(icao) != icaoLength {
		return false
	}
	for _, b := range icao {
		if b != zero {
			return true
		}
	}

	return false;
}