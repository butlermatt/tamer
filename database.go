package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"fmt"
	"os"
	"time"
)

// Locations
// +--------------------------------------------------+
// | RowID | ICAO (i) | Lat (s) | Lon (s) | TimeStamp |
// +--------------------------------------------------+
const (
	createLocationTable = `
CREATE TABLE IF NOT EXISTS Locations (icao INTEGER NOT NULL, lat TEXT, lon TEXT, time INTEGER)
`
)

// Messages
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
// | RowID | ICAO (i) | TimeStamp | CallSign (s) | Altitude (i) | Track (f) | Speed (f) | vertical (i) | Lat (s) | lon (s) | Squawk (s) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
const (
	createMsgsTable = `
CREATE TABLE IF NOT EXISTS Messages (icao INTEGER NOT NULL, time INTEGER, callsign TEXT, altitude INTEGER, track REAL, speed REAL, vertical INTEGER, lat TEXT, lon TEXT, squawk TEXT, sqch INTEGER, emerg INTEGER, ident INTEGER, grnd INTEGER)
`
)

// Squawks
// +-------------------------------+
// | RowID | ICAO (i) | Squawk (s) |
// +-------------------------------+
const (
	createSquawksTable = `
CREATE TABLE IF NOT EXISTS Squawks (icao INTEGER NOT NULL, squawk TEXT)
`
)

// Callsigns
// +---------------------------------+
// | RowID | ICAO (i) | CallSign (s) |
// +---------------------------------+
const (
	createCallsignsTable = `
CREATE TABLE IF NOT EXISTS Callsigns (icao INTEGER NOT NULL, callsign TEXT)
`
)

// Planes
// +------------------------------------------------------------------------------------------------------------------------------------------+
// | ICAO (i) Primary Key | Altitude (i) | Track (f) | Speed (f) | Vertical (i) | LastSeen (int) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +------------------------------------------------------------------------------------------------------------------------------------------+
const (
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

func close_db() error {
	err := db.Close()

	return err
}

func LoadPlane(icao uint) (*Plane, error) {
	var tt int64
	p := &Plane{Icao: icao}

	err := db.QueryRow(loadPlaneQuery, int(icao)).Scan(&p.Altitude, &p.Track, &p.Speed, &p.Vertical, &tt, &p.SquawkCh, &p.Emergency, &p.Ident, &p.OnGround)
	if err == sql.ErrNoRows {
		fmt.Printf("Unable to find plane: %06X in the db.\n", icao)
		return p, nil
	} else if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("unable to load plane %06X", icao))
	}

	p.LastSeen = time.Unix(0, tt)

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
	rows, err := db.Query("SELECT squawk FROM Squawks WHERE icao = ?", int(p.Icao))
	if err != nil {
		return errors.Wrap(err, "error retrieving Squawks")
	}
	defer rows.Close()

	for rows.Next() {
		sw := ValuePair{loaded: true}
		err = rows.Scan(&sw.value)
		if err != nil {
			return errors.Wrap(err, "error reading values from row in Squawk table")
		}
		p.Squawks = append(p.Squawks, sw)
	}

	return nil
}

func LoadCallsigns(p *Plane) error {
	rows, err := db.Query("SELECT callsign FROM Callsigns WHERE icao = ?", int(p.Icao))
	if err != nil {
		return errors.Wrap(err, "error retrieving Callsigns")
	}
	defer rows.Close()

	for rows.Next() {
		cs := ValuePair{loaded: true}
		err = rows.Scan(&cs.value)
		if err != nil {
			return errors.Wrap(err, "error reading values from row in Callsigns table")
		}
		p.CallSigns = append(p.CallSigns, cs)
	}
	if err = rows.Err(); err != nil {
		return errors.Wrap(err, "error iterating over results.")
	}

	return nil
}

func SavePlanes(planes []*Plane) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	//icao, altitude, track, speed, vertical, lastSeen, sqch, emerg, ident, grnd
	plSt, err := tx.Prepare(`INSERT OR REPLACE INTO Planes(icao, altitude, track, speed, vertical, lastSeen, sqch, emerg, ident, grnd)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	csSt, err := tx.Prepare(`INSERT INTO Callsigns(icao, callsign) VALUES(?, ?)`)
	if err != nil {
		return err
	}
	sqSt, err := tx.Prepare(`INSERT INTO Squawks(icao, squawk) VALUES(?, ?)`)
	if err != nil {
		return err
	}
	lcSt, err := tx.Prepare(`INSERT INTO Locations(icao, lat, lon, time) VALUES(?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	msgSt, err := tx.Prepare(`INSERT INTO Messages(icao, time, callsign, altitude, track, speed, vertical, lat, lon, squawk, sqch, emerg, ident, grnd)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	for _, pl := range planes {
		if pl == nil {
			continue
		}
		if verbose {
			fmt.Printf("Saving Plane: %06X\n", pl.Icao)
		}
		_, err = plSt.Exec(int(pl.Icao), pl.Altitude, pl.Track, pl.Speed, pl.Vertical, pl.LastSeen.UnixNano(), pl.SquawkCh, pl.Emergency, pl.Ident, pl.OnGround)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error writing plane: %#v", err)
		}

		for _, cs := range pl.CallSigns {
			if !cs.loaded {
				_, err = csSt.Exec(int(pl.Icao), cs.value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error writing callsign: %#v", err)
				}
			}
		}

		for _, sq := range pl.Squawks {
			if !sq.loaded {
				_, err = sqSt.Exec(int(pl.Icao), sq.value)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error writing squawk: %#v", err)
				}
			}
		}

		for _, lc := range pl.Locations {
			_, err = lcSt.Exec(int(pl.Icao), lc.Latitude, lc.Longitude, lc.Time.UnixNano())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing location: %#v", err)
			}
		}

		for _, msg := range pl.History {
			_, err = msgSt.Exec(int(msg.icao), msg.dGen.UnixNano(), msg.callSign, msg.altitude, msg.track, msg.groundSpeed, msg.vertical, msg.latitude, msg.longitude, msg.squawk, msg.squawkCh, msg.emergency, msg.ident, msg.onGround)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error writing message: %#v", err)
			}
		}
	}

	err = sqSt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing squawk statement: %#v\n", err)
	}
	err = csSt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing callsign statement: %#v\n", err)
	}
	err = lcSt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing location statement: %#v\n", err)
	}
	err = plSt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing plane statement: %#v\n", err)
	}
	err = lcSt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing locations statement: %#v\n", err)
	}
	err = msgSt.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error closing messages statement: %#v", err)
	}

	err = tx.Commit()
	return err
}