package icmp

import (
	"encoding/binary"
	"errors"

	"tcpip-go/internal/ipv4"
)

const (
	HeaderLen       = 8
	TypeEchoReply   = 0
	TypeEchoRequest = 8
	CodeEcho        = 0
)

var ErrShortPacket = errors.New("icmp: packet too short")

type Message struct {
	Type       uint8
	Code       uint8
	Identifier uint16
	Sequence   uint16
	Payload    []byte
}

func Parse(b []byte) (Message, error) {
	if len(b) < HeaderLen {
		return Message{}, ErrShortPacket
	}
	if ipv4.Checksum(b) != 0 {
		return Message{}, errors.New("icmp: bad checksum")
	}
	return Message{
		Type:       b[0],
		Code:       b[1],
		Identifier: binary.BigEndian.Uint16(b[4:6]),
		Sequence:   binary.BigEndian.Uint16(b[6:8]),
		Payload:    b[8:],
	}, nil
}

func Marshal(m Message) ([]byte, error) {
	out := make([]byte, HeaderLen+len(m.Payload))
	out[0] = m.Type
	out[1] = m.Code
	binary.BigEndian.PutUint16(out[4:6], m.Identifier)
	binary.BigEndian.PutUint16(out[6:8], m.Sequence)
	copy(out[8:], m.Payload)
	binary.BigEndian.PutUint16(out[2:4], ipv4.Checksum(out))
	return out, nil
}

func EchoReply(req Message) (Message, bool) {
	if req.Type != TypeEchoRequest || req.Code != CodeEcho {
		return Message{}, false
	}
	return Message{
		Type:       TypeEchoReply,
		Code:       CodeEcho,
		Identifier: req.Identifier,
		Sequence:   req.Sequence,
		Payload:    req.Payload,
	}, true
}
