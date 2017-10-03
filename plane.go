package main

import (
	"time"
	"bytes"
	"fmt"
)

const FreshPeriod = time.Minute * 10

type Location struct {
	id        int
	Time      time.Time
	Latitude  float32
	Longitude float32
}

type ValuePair struct {
	loaded bool
	value  string
}

// TODO: Don't change a value (or add squawk etc) if it already exists with that value.
// Don't add that to the history. However add new changes. Always update LastSeen if after
type Plane struct {
	Icao      uint
	CallSign  string
	CallSigns []ValuePair
	Squawks   []ValuePair
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

func (p *Plane) ToJson() string {
	buf := bytes.Buffer{}
	buf.WriteString("{")
	buf.WriteString(fmt.Sprintf("\"icao\": \"%06X\", ", p.Icao))
	buf.WriteString(fmt.Sprintf("\"callsign\": %q, ", p.CallSign))
	buf.WriteString("\"callsigns\": [")
	for i, cs := range p.CallSigns {
		buf.WriteString(fmt.Sprintf("%q", cs.value))
		if i != len(p.CallSigns) - 1 {
			buf.WriteString(", ")
		}
	}
	if len(p.Locations) > 0 {
		lastLoc := p.Locations[len(p.Locations) - 1]
		buf.WriteString(fmt.Sprintf("], \"location\": \"%f,%f\", ", lastLoc.Latitude, lastLoc.Longitude))
	} else {
		buf.WriteString("], ")
	}
	buf.WriteString("\"squawks\": [")
	for i, sq := range p.Squawks {
		buf.WriteString(fmt.Sprintf("%q", sq.value))
		if i != len(p.Squawks) - 1 {
			buf.WriteString(", ")
		}
	}
	buf.WriteString(fmt.Sprintf("], \"altitude\": %d, ", p.Altitude))
	buf.WriteString(fmt.Sprintf("\"track\": %.2f, ", p.Track))
	buf.WriteString(fmt.Sprintf("\"speed\": %.2f, ", p.Speed))
	buf.WriteString(fmt.Sprintf("\"vertical\": %d, ", p.Vertical))
	buf.WriteString(fmt.Sprintf("\"lastSeen\": %q", p.LastSeen.String()))
	buf.WriteString("}")

	return buf.String()
}

// SetCallSign tries to add the call sign to the slice of CallSigns for the Plane.
// Returns true if it was added, false if the callsign was already in the slice.
func (p *Plane) SetCallSign(cs string) bool {
	if cs == "" || p.CallSign == cs {
		return false
	}
	p.CallSign = cs

	// More likely to find a match at the end, so search backwards.
	for i := len(p.CallSigns) - 1; i >= 0; i-- {
		if cs == p.CallSigns[i].value {
			// if it exists, don't add it again
			return true // True because at the least the current callsign has been set.
		}
	}

	p.CallSigns = append(p.CallSigns, ValuePair{value: cs})
	return true
}

// SetSquawk tries to add the Squawk code to the slice of Squawks for the Plane.
// Returns true if it was added, false if the Squawk was already in the slice.
func (p *Plane) SetSquawk(s string) bool {
	// Search backwards as more likely to find result at end.
	if s == "" {
		return false
	}

	for i := len(p.Squawks) - 1; i >= 0; i-- {
		if s == p.Squawks[i].value {
			return false
		}
	}

	p.Squawks = append(p.Squawks, ValuePair{value: s})
	return true
}

// SetLocation creates a location from the specified Lat/lon and time and appends it
// to the locations slice. Returns true if successful, and false if there are no values to add
func (p *Plane) SetLocation(lat, lon float32, t time.Time) bool {
	if lat == 0.0 || lon == 0.0 {
		return false
	}
	l := Location{Time: t, Latitude: lat, Longitude: lon}
	p.Locations = append(p.Locations, l)
	return true
}

// SetAltitude will update the altitude if different from existing altitude.
// Returns true if successful, false if there is no change.
func (p *Plane) SetAltitude(a int) bool {
	if a != 0 && p.Altitude != a {
		p.Altitude = a
		return true
	}
	return false
}

// SetTrack sets the Plane's current track if different from existing value.
// Returns true if successful, and false if there is no change.
func (p *Plane) SetTrack(t float32) bool {
	if p.Track != t {
		p.Track = t
		return true
	}
	return false
}

// SetSpeed sets the Plane's speed if different from existing value.
// Returns true if successful, and false if there is no change.
func (p *Plane) SetSpeed(s float32) bool {
	if p.Speed != s && s != 0.0 {
		p.Speed = s
		return true
	}

	return false
}

// SetVertical sets the Plane's vertical speed, if different from existing value.
// Returns true if successful, and false if there was no change.
func (p *Plane) SetVertical(v int) bool {
	if p.Vertical != v {
		p.Vertical = v
		return true
	}
	return false
}

// SetHistory appends the history message to the history slice.
func (p *Plane) SetHistory(m *message) {
	p.History = append(p.History, m)
}

// SetSquawkCh sets the Plane's SquawkChange flag if different from existing value.
// Returns true on success, and false if there is no change.
func (p *Plane) SetSquawkCh(s bool) bool {
	if p.SquawkCh != s {
		p.SquawkCh = s
		return true
	}
	return false
}

// SetEmergency sets the Plane's emergency flag if different from existing value.
// Returns true on success, and false if there is no change.
func (p *Plane) SetEmergency(e bool) bool {
	if p.Emergency != e {
		p.Emergency = e
		return true
	}
	return false
}

// SetIdent sets the Plane's ident flag if different from existing value.
// Returns true on success, and false if there is no change.
func (p *Plane) SetIdent(i bool) bool {
	if p.Ident != i {
		p.Ident = i
		return true
	}
	return false
}

// SetOnGround sets the Plane's onGround flag if different from existing value.
// Returns true on success, and false if there is no change.
func (p *Plane) SetOnGround(g bool) bool {
	if p.OnGround != g {
		p.OnGround = g
		return true
	}
	return false
}

func updatePlane(m *message, pl *Plane) {
	buf := bytes.Buffer{}
	if m == nil {
		return
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
			dataStr = fmt.Sprintf(" Altitude: %d, Lat: %f, Lon: %f", m.altitude, m.latitude, m.longitude)
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

	// Log message if it updated a value, or the last message was more than 10 minutes ago
	if written || m.dGen.Sub(pl.LastSeen) > FreshPeriod {
		pl.SetHistory(m)
	}

	if verbose {
		buf.WriteString(dataStr)

		fmt.Println(buf.String())
		buf.Reset()
	}
}
