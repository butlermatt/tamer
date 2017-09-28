package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// Locations
// +-------------------------------------------------------+
// | RowID | ICAO (i) | Location (Lat,Lon (s)) | TimeStamp |
// +-------------------------------------------------------+
const (
	locationTable = `Locations`
	locationTableCreate = `
CREATE TABLE IF NOT EXISTS Locations (icao INTEGER NOT NULL, loc TEXT, time INTEGER)
`
)

// Messages
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
// | RowID | ICAO (i) | TimeStamp | CallSign (s) | Altitude (i) | Track (f) | Speed (f) | vertical (i) | Lat (s) | lon (s) | Squawk (s) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
const (
	msgsTable = `Messages`
	msgsTableCreate = `
CREATE TABLE IF NOT EXISTS Messages (icao INTEGER NOT NULL, time INTEGER, altitude INTEGER, track REAL, speed REAL, vertical INTEGER, lat TEXT, lon TEXT, sqch INTEGER, emerg INTEGER, ident INTEGER, grnd INTEGER)
`
)

// Squawks
// +-------------------------------+
// | RowID | ICAO (i) | Squawk (s) |
// +-------------------------------+
const (
	squawksTable = `Squawks`
	squawksTableCreate = `
CREATE TABLE IF NOT EXISTS Squawks (icao INTEGER NOT NULL, squawk TEXT)
`
)

// Callsigns
// +---------------------------------+
// | RowID | ICAO (i) | CallSign (s) |
// +---------------------------------+
const (
	callSignsTable = `Planes`
	callsignsTableCreate = `
CREATE TABLE IF NOT EXISTS Callsigns (icao INTEGER NOT NULL, callsign TEXT)
`
)

// Planes
// +------------------------------------------------------------------------------------------------------------------------------------------+
// | ICAO (i) Primary Key | Altitude (i) | Track (f) | Speed (f) | Vertical (i) | LastSeen (int) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +------------------------------------------------------------------------------------------------------------------------------------------+
const (
	planeTable = `Planes`
	planeTableCreate = `
CREATE TABLE IF NOT EXISTS Planes (icao INTEGER PRIMARY KEY, altitude INTEGER, track REAL, speed REAL, vertical INTEGER, lastSeen INTEGER, sqch INTEGER, emerg INTEGER, ident INTEGER, grnd INTEGER)
`
)