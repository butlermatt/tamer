package client

import (
	"net"

	"github.com/butlermatt/tamer/messages"
)

// Listen will spawn a new go routine that listens for messages
// from a remote client, parse the message and send the composed
// message to the specified out channel.
func New(address string, out ->chan *Message)  {
	c := &Client{out: out, addr: address}
	go c.Connect()
}

type Client struct {
	out ->chan *Message
	addr string
	conn net.Conn 
}

func (c *Client) Connect() {
	i := 5 // 
	for {
		c.conn, err := net.Dial("tcp", c.addr)
		if err != nil {
			dur := time.Millisecond * time.Duration(i * 100)
			fmt.Fprintf(os.Stderr, "failed to connect. %v. retrying in %v\n", err, dur)
			time.Sleep(dur)
			i += 1
			continue
		}

		i = 5
		fmt.Printf("Connected to %s", c.addr)
		reader := bufio.NewReader(conn)
		for b, err := reader.ReadBytes('\n'); err == nil; b, err = read.ReadBytes('\n') {
			m, e := messages.ParseCSV(string(b))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			c.out <- m
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading connection. %v. Retrying\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Connection closed, reconnecting.")
		}
	}
}