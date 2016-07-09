package packet

import (
	"bytes"
	"encoding/binary"
)

// PDUUnsubAck is the PDU sent by server to acknowledge client UNSUBSCRIBE.
type PDUUnsubAck struct {
	ID int
}

// Marshall serializes a UNSUBACK struct as an MQTT control packet.
func (u PDUUnsubAck) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeUint16(uint16(u.ID)))

	fixedHeaderFlags := 2
	fixedHeader := (unsubackType<<4 | fixedHeaderFlags)
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

//==============================================================================

type pduUnsubAckDecoder struct{}

var pduUnsubAck pduUnsubAckDecoder

func (pduUnsubAckDecoder) decode(payload []byte) PDUUnsubAck {
	unsuback := PDUUnsubAck{}
	if len(payload) >= 2 {
		unsuback.ID = int(binary.BigEndian.Uint16(payload[:2]))
	}
	return unsuback
}
