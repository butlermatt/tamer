package main

import "time"

type Location struct {
	Time      time.Time
	Latitude  string
	Longitude string
}

// TODO: Don't change a value (or add squawk etc) if it already exists with that value.
// Don't add that to the history. However add new changes. Always update LastSeen if after
type Plane struct {
	Icao      uint
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

// SetCallSign tries to add the call sign to the slice of CallSigns for the Plane.
// Returns true if it was added, false if the callsign was already in the slice.
func (p *Plane) SetCallSign(cs string) bool {
	if cs == "" {
		return false
	}

	// More likely to find a match at the end, so search backwards.
	for i := len(p.CallSigns) - 1; i >= 0; i-- {
		if cs == p.CallSigns[i] {
			return false
		}
	}

	p.CallSigns = append(p.CallSigns, cs)
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
		if s == p.Squawks[i] {
			return false
		}
	}

	p.Squawks = append(p.Squawks, s)
	return true
}

// SetLocation creates a location from the specified Lat/lon and time and appends it
// to the locations slice. Returns true if successful, and false if there are no values to add
func (p *Plane) SetLocation(lat, lon string, t time.Time) bool {
	if lat == "" || lon == "" {
		return false
	}
	l := Location{t, lat, lon}
	p.Locations = append(p.Locations, l)
	return true
}

// SetAltitude will update the altitude if different from existing altitude.
// Returns true if successful, false if there is no change.
func (p *Plane) SetAltitude(a int) bool {
	if p.Altitude != a {
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