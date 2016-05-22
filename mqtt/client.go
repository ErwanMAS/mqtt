/*
MQTT package implements MQTT protocol. It can be use as a client library to write MQTT clients in Go.
*/
package mqtt

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/processone/gomqtt/mqtt/packet"
)

// Client is the main structure use to connect as a client on an MQTT
// server.
type Client struct {
	mu sync.RWMutex
	// Store user defined options
	options ClientOptions
	// TCP level connection / can be replaced by a TLS session after starttls
	conn         net.Conn
	backoff      backoff
	keepaliveCtl chan int
	message      chan *Message
	sender       chan []byte
}

// TODO split channel between status signals (informing about the state of the client) and message received (informing
// about the publish we have received)
// We also should abstract the Message to hide the details of the protocol from the developer client: MQTT protocol could
// change on the wire, but we can likely keep the same internal format for publish messages received.

const (
	stateConnecting   = iota
	stateConnected    = iota
	stateReconnecting = iota
	stateDisconnected = iota
)

// Message encapsulates Publish MQTT payload from the MQTT client perspective.
type Message struct {
	Topic   string
	Payload []byte
}

// NewClient generates a new MQTT client, based on Options passed as parameters.
// Default the port to 1883.
func NewClient(options ClientOptions) (c *Client, err error) {
	if options.Address, err = checkAddress(options.Address); err != nil {
		return
	}

	c = new(Client)
	c.options = options

	return
}

func checkAddress(addr string) (string, error) {
	var err error
	hostport := strings.Split(addr, ":")
	if len(hostport) > 2 {
		err = errors.New("too many colons in server address")
		return addr, err
	}

	// Address is composed of two parts, we are good
	if len(hostport) == 2 && hostport[1] != "" {
		return addr, err
	}

	// Port was not passed, we append default MQTT port:
	return strings.Join([]string{hostport[0], "1883"}, ":"), err
}

// Connect initiates synchronous connection to MQTT server
func (c *Client) Connect() error {
	return c.connect(false)
}

// TODO Serialize packet send into its own channel / go routine
//
// FIXME(mr) packet.Topic does not seem a good name
func (c *Client) Subscribe(topic packet.Topic) {
	subscribe := packet.NewSubscribe()
	subscribe.AddTopic(topic)
	buf := subscribe.Marshall()
	c.send(&buf)
}

func (c *Client) Unsubscribe(topic string) {
	unsubscribe := packet.NewUnsubscribe()
	unsubscribe.AddTopic(topic)
	buf := unsubscribe.Marshall()
	c.send(&buf)
}

func (c *Client) ReadNext() *Message {
	return <-c.message
}

func (c *Client) Publish(topic string, payload []byte) {
	publish := packet.NewPublish()
	publish.SetTopic(topic)
	publish.SetPayload(payload)
	buf := publish.Marshall()
	c.send(&buf)
}

// Disconnect sends DISCONNECT MQTT packet to other party
func (c *Client) Disconnect() {
	buf := packet.NewDisconnect().Marshall()
	c.send(&buf)
}

func (c *Client) send(buf *bytes.Buffer) {
	buf.WriteTo(c.getConn())
	c.resetTimer()
}

func (c *Client) resetTimer() {
	c.keepaliveCtl <- keepaliveReset
}

func (c *Client) connect(retry bool) error {
	fmt.Println("Trying to connect")
	conn, err := net.DialTimeout("tcp", c.options.Address, 5*time.Second)
	if err != nil {
		if !retry {
			return err
		}
		// Sleep with exponential backoff (and jitter) before triggering reconnect:
		time.AfterFunc(c.backoff.Duration(), func() { c.connect(retry) })
		return nil
	}

	// 1. Open session - Login
	// Send connect packet
	connectPacket := packet.NewConnect()
	connectPacket.SetKeepalive(c.options.Keepalive)
	buf := connectPacket.Marshall()
	buf.WriteTo(conn)

	// TODO Check connack value to validate session open success
	packet.Read(conn)

	// 2. Connected. We set environment up
	c.backoff.Reset()
	// For now we do not need to change the message channel
	if c.message == nil {
		c.message = make(chan *Message)
	}
	c.setConn(conn) // TODO: Avoid double lock / unlock on first connect

	// Start go routine that manage keepalive timer:
	c.keepaliveCtl = startKeepalive(c, func() {
		pingReq := packet.NewPingReq()
		buf := pingReq.Marshall()
		buf.WriteTo(conn)
	})

	// Routine to receive incoming data
	// func receiver2(conn net.Conn, tearDown chan<- struct{}, message chan<- *Message) {
	tearDown := initReceiver(conn, c.message)
	go c.disconnected(tearDown)
	return nil
}

// get receiver tearDown signal, clean client state and trigger reconnect
func (c *Client) disconnected(stopSignal <-chan struct{}) {
	<-stopSignal
	c.conn.Close()
	c.keepaliveCtl <- keepaliveStop
	go c.connect(true)
}

func (c *Client) getConn() net.Conn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

func (c *Client) setConn(conn net.Conn) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn = conn
}

/*
type sender struct {
}

func (c *Client) sender(out chan *bytes.Buffer) {
	select {
	case buf := <-out:
		buf.WriteTo(c.getConn())
		c.resetTimer()
	}
}
*/
