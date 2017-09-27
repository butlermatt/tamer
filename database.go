package main

// Locations
// +-------------------------------------------------------+
// | RowID | ICAO (i) | Location (Lat,Lon (s)) | TimeStamp |
// +-------------------------------------------------------+

// Messages
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+
// | RowID | ICAO (i) | TimeStamp | CallSign (s) | Altitude (i) | Speed (f) | Track (f) | Lat (s) | lon (s) | vertical (i) | Squawk (s) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------+

// Squawks
// +-------------------------------+
// | RowID | ICAO (i) | Squawk (s) |
// +-------------------------------+

// Callsigns
// +---------------------------------+
// | RowID | ICAO (i) | CallSign (s) |
// +---------------------------------+

// Planes
// +------------------------------------------------------------------------------------------------------------------------------------------+
// | ICAO (i) Primary Key | Altitude (i) | Track (f) | Speed (f) | Vertical (i) | LastSeen (ts) | SqCh (b) | Emerg (b) | Ident (b) | Grnd (b) |
// +------------------------------------------------------------------------------------------------------------------------------------------+

