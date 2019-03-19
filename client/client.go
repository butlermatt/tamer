package client

import "github.com/butlermatt/tamer/messages"

// Listen will spawn a new go routine that listens for messages
// from a remote client, parse the message and send the composed
// message to the specified out channel.
func New(address string, out ->chan *Message)  {
	
}