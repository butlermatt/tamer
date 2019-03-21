package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

var magicTimestampMLAT = []byte{0xFF, 0x00, 0x4D, 0x4C, 0x41, 0x54}

func main() {
	conn, err := net.Dial("tcp", "192.168.2.19:30005")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error connecting to server\n")
		os.Exit(1);
	}

	reader := bufio.NewReader(conn)
	var buffer []byte
	for {
		// TODO: (mbutler) Instead of reading a "line" like this,
		// do character reading and check for double escapes etc.
		// Should start from 0x1A not end.
		msg, _ := reader.ReadBytes(0x1A)

		if buffer == nil {
			buffer = msg
		} else {
			buffer = append(buffer, msg...)
		}

		if msg[0] < 0x31 || msg[0] > 0x34 {
			fmt.Printf("Buffering\n")
			continue // Don't have a valid message code yet.
		}

		msg = buffer
		buffer = nil
		var msgLen int
		switch msg[0] {
		case 0x31:
			msgLen = 11
		case 0x32:
			msgLen = 16
		case 0x33:
			msgLen = 23
		case 0x34:
			continue
		default:
			continue // I dunno?
		}

		if len(msg) != msgLen {
			fmt.Printf("Message length was wrong: %d\n", len(msg))
			continue
		}

		mlat := isMlat(msg[1:7])
		t := parseTime(msg[1:7])
		fmt.Printf("%v: Mlat: %t Message: %v\n", t.Local(), mlat, hexMsg(msg))
	}
}

func hexMsg(b []byte) string {
	var buf strings.Builder

	buf.WriteByte('[')
	for i, bb := range b {
		buf.WriteString(fmt.Sprintf("%#x", bb))
		if i != len(b) - 1 {
			buf.WriteByte(' ')
		}
	}

	buf.WriteByte(']')
	return buf.String()
}

func isMlat(ts []byte) bool {
	if len(ts) != len(magicTimestampMLAT) {
		return false
	}

	for i := range ts {
		if ts[i] != magicTimestampMLAT[i] {
			return false
		}
	}

	return true
}

func parseTime(timebytes []byte) time.Time {
	// Takes a 6 byte array, which represents a 48bit GPS timestamp
	// http://wiki.modesbeast.com/Radarcape:Firmware_Versions#The_GPS_timestamp
	// and parses it into a Time.time

	upper := []byte{
		timebytes[0]<<2 + timebytes[1]>>6,
		timebytes[1]<<2 + timebytes[2]>>6,
		0, 0}
	lower := []byte{
		timebytes[2] & 0x3F, timebytes[3], timebytes[4], timebytes[5]}

	// TODO: (mbutler) Try looking at this:
	// https://github.com/flightaware/dump1090/blob/master/net_io.c#L40

	// the 48bit timestamp is 18bit day seconds | 30bit nanoseconds
	daySeconds := binary.BigEndian.Uint16(upper)
	nanoSeconds := int(binary.BigEndian.Uint32(lower))

	hr := int(daySeconds / 3600)
	min := int(daySeconds / 60 % 60)
	sec := int(daySeconds % 60)

	fmt.Printf("Hour: %d, Min: %d, Sec: %d\n", hr, min, sec)

	utcDate := time.Now().UTC()

	return time.Date(
		utcDate.Year(), utcDate.Month(), utcDate.Day(),
		hr, min, sec, nanoSeconds, time.UTC)
}