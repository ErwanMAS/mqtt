package packet

import (
	"bytes"
	"fmt"
)

// Packet interface shared by all MQTT control packets
type Marshaller interface {
	Marshall() bytes.Buffer
	PacketType() int
}

// TODO interface should be packet.Marshaller ?

func NewConnect() *Connect {
	return new(Connect)
}

func Decode(packetType int, payload []byte) Marshaller {
	switch packetType {
	case 2:
		return DecodeConnAck(payload)
	default:
		fmt.Println("Unsupported MQTT packet type")
		return nil
	}
}
