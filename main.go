package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	addr string
)

const (
	// Escape code
	esc byte = 0x1A
)

const (
	// AC Mode
	modeAc byte = '1'
	// Mode S Short Frame
	modeSs byte = '2'
	// Mode S Long Frame
	modeSl byte = '3'
	// Status Data
	modeSd byte = '4'
)

const (
	dfShortTacs         uint = 0
	dfSurveillanceAlt   uint = 4
	dfSurveillanceIdent uint = 5
	dfAllCall           uint = 11
	dfLongTacs          uint = 16
	dfExtended          uint = 17
	dfTisB              uint = 18
	dfExtendedMil       uint = 19
	dfCommBAlt          uint = 20
	dfCommBModeA        uint = 21
	dfMilitary          uint = 22
	dfLongReply         uint = 24
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30005", "Address and port to connect to for input.")
}

func main() {
	flag.Parse()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect. %v", err)
		os.Exit(1)
	}
	fmt.Println("Connected")

	reader := bufio.NewReader(conn)

	buf := bytes.Buffer{}

	for b, err := reader.ReadByte(); err == nil; b, err = reader.ReadByte() {
		if b == esc {
			t, err := reader.Peek(1)
			if err != nil {
				break
			}
			if t[0] == esc {
				err = buf.WriteByte(b)
				_, _ = reader.ReadByte()
				//fmt.Println("Saved invalid break")
				continue
			}
			if buf.Len() > 0 {
				processMessage(buf.Bytes()) // Make this a channel instead? Or a go routine
				buf = bytes.Buffer{}
			}
		}

		buf.WriteByte(b)
	}
}

func processMessage(m []byte) {
	if len(m) < 2 {
		fmt.Println("Discarding empty message")
		return
	}

	if m[0] != esc {
		fmt.Printf("Invalid message received: %X\n", m)
	}

	var msgType string
	switch m[1] {
	case modeAc:
		msgType = "AC"
	case modeSs:
		msgType = "S-Short"
	case modeSl:
		msgType = "S-Long"
	case modeSd:
		msgType = "Data dump?"
	default:
		msgType = "Continuation?"
	}

	tm := parseTime(m[2:8])
	modeS := parseModeS(m[9:])
	fmt.Printf("%X - Type: %s, Time: %v Signal: %d. Msg Type: %v\n", m, msgType, tm, m[8], modeS)
}

func parseTime(t []byte) time.Time {
	// Takes a 6 byte array, which represents a 48bit GPS timestamp
	// http://wiki.modesbeast.com/Radarcape:Firmware_Versions#The_GPS_timestamp
	// and parses it into a Time.time

	upper := []byte{
		t[0]<<2 + t[1]>>6,
		t[1]<<2 + t[2]>>6,
		0, 0, 0, 0}
	lower := []byte{
		t[2] & 0x3F, t[3], t[4], t[5]}

	// the 48bit timestamp is 18bit day seconds | 30bit nanoseconds
	secs := binary.BigEndian.Uint16(upper)
	nano := int(binary.BigEndian.Uint32(lower))

	hr := int(secs / 3600)
	min := int(secs / 60 % 60)
	sec := int(secs % 60)

	utcDate := time.Now().UTC()

	return time.Date(
		utcDate.Year(), utcDate.Month(), utcDate.Day(),
		hr, min, sec, nano, time.UTC)
}

func parseModeS(m []byte) string {
	bad := true
	for i := 0; i < len(m); i++ {
		if m[i] != 0 {
			bad = false
			break
		}
	}

	if bad {
		return "All zeros"
	}

	typ := uint((m[0] & 0xF8) >> 3)

	switch typ {
	case dfShortTacs:
		return "Short Tacs"
	case dfSurveillanceAlt:
		return "Surveillance Altitude"
	case dfSurveillanceIdent:
		return "Surveillance Mode A"
	case dfAllCall:
		return "All Call"
	case dfLongTacs:
		return "Long Tacs"
	case dfExtended:
		return "Extended Squitter"
	case dfTisB:
		return "TisB"
	case dfExtendedMil:
		return "Military Extended Squitter"
	case dfCommBAlt:
		return "Comm B Altitude"
	case dfCommBModeA:
		return "Comm B with Mode A"
	case dfMilitary:
		return "Military"
	case dfLongReply:
		return "Long Rely"
	}

	return ""
}
