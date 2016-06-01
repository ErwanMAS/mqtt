package mqtt

import (
	"fmt"
	"io"
	"net"

	"github.com/processone/gomqtt/mqtt/packet"
)

// Receiver actually need:
// - Net.conn
// - Sender (to send ack packet when packets requiring acks are received)
// - Error send channel to trigger teardown
// - MessageSendChannel to dispatch messages to client

func initReceiver(conn net.Conn, messageChannel chan *Message, s sender) <-chan struct{} {
	tearDown := make(chan struct{})
	go receiver(conn, tearDown, messageChannel, s)
	return tearDown
}

// Receive, decode and dispatch messages to the message channel
func receiver(conn net.Conn, tearDown chan<- struct{}, message chan<- *Message, s sender) {
	var p packet.Marshaller
	var err error

Loop:
	for {
		if p, err = packet.Read(conn); err != nil {
			if err == io.EOF {
				fmt.Printf("Connection closed\n")
			}
			fmt.Printf("packet read error: %q\n", err)
			break Loop
		}
		fmt.Printf("Received: %+v\n", p)

		sendAckIfNeeded(p, s)

		// Only broadcast message back to client when we receive publish packets
		switch packetType := p.(type) {
		case *packet.Publish:
			m := new(Message)
			m.Topic = packetType.Topic
			m.Payload = packetType.Payload
			message <- m // TODO Back pressure. We may block on processing message if client does not read fast enough. Make sure we can quit.
		default:
		}
	}

	// Loop ended, send tearDown signal and close tearDown channel
	tearDown <- struct{}{}
	close(tearDown)
}

// Send acks if needed, depending on packet QOS
func sendAckIfNeeded(pkt packet.Marshaller, s sender) {
	switch p := pkt.(type) {
	case *packet.Publish:
		if p.Qos == 1 {
			puback := packet.NewPubAck(p.ID)
			buf := puback.Marshall()
			s.send(&buf)
		}
	}
}
