package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"fmt"
)

// Locations
// +--------------------------------------------------+
// | RowID | ICAO (i) | Lat (s) | Lon (s) | TimeStamp |
// +--------------------------------------------------+
const (
	locationTable = `Locations`
	createLocationTable = `
CREATE TABLE IF NOT EXISTS Locations (icao INTEGER NOT NULL, lat TEXT, lon TEXT, time INTEGER)
`
)

// Messages
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
// | RowID | ICAO (i) | TimeStamp | CallSign (s) | Altitude (i) | Track (f) | Speed (f) | vertical (i) | Lat (s) | lon (s) | Squawk (s) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
const (
	msgsTable = `Messages`
	createMsgsTable = `
CREATE TABLE IF NOT EXISTS Messages (icao INTEGER NOT NULL, time INTEGER, altitude INTEGER, track REAL, speed REAL, vertical INTEGER, lat TEXT, lon TEXT, sqch INTEGER, emerg INTEGER, ident INTEGER, grnd INTEGER)
`
)

// Squawks
// +-------------------------------+
// | RowID | ICAO (i) | Squawk (s) |
// +-------------------------------+
const (
	squawksTable = `Squawks`
	createSquawksTable = `
CREATE TABLE IF NOT EXISTS Squawks (icao INTEGER NOT NULL, squawk TEXT)
`
)

// Callsigns
// +---------------------------------+
// | RowID | ICAO (i) | CallSign (s) |
// +---------------------------------+
const (
	callSignsTable = `Callsigns`
	createCallsignsTable = `
CREATE TABLE IF NOT EXISTS Callsigns (icao INTEGER NOT NULL, callsign TEXT)
`
)

// Planes
// +------------------------------------------------------------------------------------------------------------------------------------------+
// | ICAO (i) Primary Key | Altitude (i) | Track (f) | Speed (f) | Vertical (i) | LastSeen (int) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +------------------------------------------------------------------------------------------------------------------------------------------+
const (
	planeTable = `Planes`
	createPlaneTable = `
CREATE TABLE IF NOT EXISTS Planes (icao INTEGER PRIMARY KEY, altitude INTEGER, track REAL, speed REAL, vertical INTEGER, lastSeen INTEGER, sqch INTEGER, emerg INTEGER, ident INTEGER, grnd INTEGER)
`
	loadPlaneQuery = `SELECT altitude, track, speed, vertical, lastSeen, sqch, emerg, ident, grnd FROM Planes WHERE icao = ?`
)

var db *sql.DB

func init_db() error {
	var err error
	db, err = sql.Open("sqlite3", "./planes.db")
	if err != nil {
		return err
	}

	_, err = db.Exec(createPlaneTable)
	if err != nil {
		return errors.Wrap(err, "unable to create Plane table.")
	}
	_, err = db.Exec(createCallsignsTable)
	if err != nil {
		return errors.Wrap(err, "unable to create Callsign table.")
	}
	_, err = db.Exec(createSquawksTable)
	if err != nil {
		return errors.Wrap(err, "unable to create Squawks table.")
	}
	_, err = db.Exec(createMsgsTable)
	if err != nil {
		return errors.Wrap(err, "unable to create Messages table.")
	}
	_, err = db.Exec(createLocationTable)
	if err != nil {
		return errors.Wrap(err, "unable to create Locations table.")
	}

	return nil
}

func LoadPlane(icao uint) (*Plane, error) {
	p := &Plane{Icao: icao}
	err := db.QueryRow(loadPlaneQuery, int(icao)).Scan(&p.Altitude, &p.Track, &p.Speed, &p.Vertical, &p.LastSeen, &p.SquawkCh, &p.Emergency, &p.Ident, &p.OnGround)

	if err == sql.ErrNoRows {
		fmt.Printf("Unable to find plane: %06X in the db.\n", icao)
		return p, nil
	} else if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to load plane %06X", icao))
	}

	fmt.Println("Found plane in DB. Loading other values")
	err = LoadSquawks(p)
	if err != nil {
		return nil, err
	}
	err = LoadCallsigns(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func LoadSquawks(p *Plane) error {
	rows, err := db.Query("SELECT ROWID, squawk FROM Squawks WHERE icao = ?", int(p.Icao))
	if err != nil {
		return errors.Wrap(err, "error retrieving Squawks")
	}
	defer rows.Close()

	for rows.Next() {
		var sw ValuePair
		err = rows.Scan(&sw.id, &sw.value)
		if err != nil {
			return errors.Wrap(err, "error reading values from row in Squawk table")
		}
		p.Squawks = append(p.Squawks, sw)
	}

	return nil
}

func LoadCallsigns(p *Plane) error {
	rows, err := db.Query("SELECT ROWID, callsign FROM Callsigns WHERE icao = ?", int(p.Icao))
	if err != nil {
		return errors.Wrap(err, "error retrieving Callsigns")
	}
	defer rows.Close()

	for rows.Next() {
		var cs ValuePair
		err = rows.Scan(&cs.id, &cs.value)
		if err != nil {
			return errors.Wrap(err, "error reading values from row in Callsigns table")
		}
		p.CallSigns = append(p.CallSigns, cs)
	}

	return nil
}

func SavePlanes(planes []*Plane) error {
	// TODO: This
	return fmt.Errorf("Not yet implemented! %#v", planes)
}