package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
	"bytes"
)

var (
	addr string
)

func init() {
	flag.StringVar(&addr, "a", "localhost:30003", "Address and port to connect to for input.")
}

func main() {
	flag.Parse()

	msgs := make(chan *planeMsg)

	go connect(msgs)

	buf := bytes.Buffer{}
	for p := range msgs {

		buf.WriteString("Received message: Plane: \"")
		buf.WriteString(p.icoa)
		buf.WriteString("\" At: ")
		buf.WriteString(p.msg.dRec.String())

		var dataStr string
		switch p.msg.tType {
		case 1:
			dataStr = fmt.Sprintf(" callsign: %q", p.msg.callSign)
		case 2:
			dataStr = fmt.Sprintf(" Altitude: %d, Speed: %.2f, Track: %.2f, Lat: %s, Lon: %s", p.msg.altitude, p.msg.groundSpeed, p.msg.track, p.msg.latitude, p.msg.longitude)
		case 3:
			dataStr = fmt.Sprintf(" Altitude: %d, Lat: %s, Lon: %s", p.msg.altitude, p.msg.latitude, p.msg.longitude)
		case 4:
			dataStr = fmt.Sprintf(" Speed: %.2f, Track: %.2f, Vertical Rate: %d", p.msg.groundSpeed, p.msg.track, p.msg.vertical)
		case 5:
			dataStr = fmt.Sprintf(" Altitude: %d", p.msg.altitude)
		case 6:
			dataStr = fmt.Sprintf(" Altitude: %d, SquawkCode: %q", p.msg.altitude, p.msg.squawk)
		case 7:
			dataStr = fmt.Sprintf(" Altitude: %d", p.msg.altitude)
		case 8:
			dataStr = fmt.Sprintf(" OnGround: %v", p.msg.onGround)
		}

		buf.WriteString(dataStr)

		fmt.Println(buf.String())
		buf.Reset()
	}

}

func connect(out chan<- *planeMsg) {
	i := 1;
	for {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			dur := time.Millisecond * time.Duration(i) * time.Duration(100)
			fmt.Fprintf(os.Stderr, "Failed to connect. %v. Retrying in %v\n", err, dur)
			time.Sleep(dur)
			i += i
			continue
		}
		i = 1
		fmt.Println("Connected")
		reader := bufio.NewReader(conn)
		for b, err := reader.ReadBytes('\n'); err == nil; b, err = reader.ReadBytes('\n') {
			fmt.Printf("%s", b)
			go parseMessage(b, out)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading connection. %v. Retrying\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Connection closed, reconnecting.")
		}
	}
}