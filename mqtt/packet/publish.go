package packet

import (
	"bytes"
	"encoding/binary"
)

type Publish struct {
	ID      int
	Dup     bool
	Qos     int
	Retain  bool
	Topic   string
	Payload []byte
}

func (p *Publish) SetTopic(topic string) {
	p.Topic = topic
}

func (p *Publish) SetPayload(payload []byte) {
	p.Payload = payload
}

func (p *Publish) Marshall() bytes.Buffer {
	var variablePart bytes.Buffer
	var packet bytes.Buffer

	variablePart.Write(encodeString(p.Topic))
	if p.Qos == 1 || p.Qos == 2 {
		variablePart.Write(encodeUint16(uint16(p.ID)))
	}
	variablePart.Write([]byte(p.Payload))

	fixedHeader := (publishType<<4 | bool2int(p.Dup)<<3 | p.Qos<<1 | bool2int(p.Retain))
	packet.WriteByte(byte(fixedHeader))
	packet.WriteByte(byte(variablePart.Len()))
	packet.Write(variablePart.Bytes())

	return packet
}

func decodePublish(fixedHeaderFlags int, payload []byte) *Publish {
	publish := NewPublish()
	publish.Dup = int2bool(fixedHeaderFlags >> 3)
	publish.Qos = int((fixedHeaderFlags & 6) >> 1)
	publish.Retain = int2bool((fixedHeaderFlags & 1))
	var rest []byte
	publish.Topic, rest = extractNextString(payload)
	var index int
	if len(rest) > 0 {
		if publish.Qos == 1 || publish.Qos == 2 {
			offset := 2
			publish.ID = int(binary.BigEndian.Uint16(rest[:offset]))
			index = offset
		}
		if len(rest) > index {
			publish.Payload = rest[index:]
		}
	}
	return publish
}
