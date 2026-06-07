package tcp

import (
	"encoding/binary"
	"errors"

	"tcpip-go/internal/ipv4"
)

const (
	MinHeaderLen = 20
	FlagFIN      = 0x01
	FlagSYN      = 0x02
	FlagRST      = 0x04
	FlagPSH      = 0x08
	FlagACK      = 0x10
	DefaultWin   = 65535
)

var ErrShortSegment = errors.New("tcp: segment too short")

type Segment struct {
	SrcPort uint16
	DstPort uint16
	Seq     uint32
	Ack     uint32
	Flags   uint8
	Window  uint16
	Urgent  uint16
	Options []byte
	Payload []byte
}

func Parse(b []byte, src, dst ipv4.Addr) (Segment, error) {
	if len(b) < MinHeaderLen {
		return Segment{}, ErrShortSegment
	}
	dataOffset := int(b[12]>>4) * 4
	if dataOffset < MinHeaderLen || dataOffset > len(b) {
		return Segment{}, errors.New("tcp: invalid data offset")
	}
	if Checksum(src, dst, b) != 0 {
		return Segment{}, errors.New("tcp: bad checksum")
	}
	return Segment{
		SrcPort: binary.BigEndian.Uint16(b[0:2]),
		DstPort: binary.BigEndian.Uint16(b[2:4]),
		Seq:     binary.BigEndian.Uint32(b[4:8]),
		Ack:     binary.BigEndian.Uint32(b[8:12]),
		Flags:   b[13],
		Window:  binary.BigEndian.Uint16(b[14:16]),
		Urgent:  binary.BigEndian.Uint16(b[18:20]),
		Options: append([]byte(nil), b[20:dataOffset]...),
		Payload: b[dataOffset:],
	}, nil
}

func Marshal(s Segment, src, dst ipv4.Addr) ([]byte, error) {
	if len(s.Options)%4 != 0 {
		return nil, errors.New("tcp: options length must be multiple of 4")
	}
	headerLen := MinHeaderLen + len(s.Options)
	if headerLen > 60 {
		return nil, errors.New("tcp: header too long")
	}
	out := make([]byte, headerLen+len(s.Payload))
	binary.BigEndian.PutUint16(out[0:2], s.SrcPort)
	binary.BigEndian.PutUint16(out[2:4], s.DstPort)
	binary.BigEndian.PutUint32(out[4:8], s.Seq)
	binary.BigEndian.PutUint32(out[8:12], s.Ack)
	out[12] = byte(headerLen/4) << 4
	out[13] = s.Flags
	if s.Window == 0 {
		s.Window = DefaultWin
	}
	binary.BigEndian.PutUint16(out[14:16], s.Window)
	binary.BigEndian.PutUint16(out[18:20], s.Urgent)
	copy(out[20:headerLen], s.Options)
	copy(out[headerLen:], s.Payload)
	binary.BigEndian.PutUint16(out[16:18], Checksum(src, dst, out))
	return out, nil
}

func SeqLen(s Segment) uint32 {
	n := uint32(len(s.Payload))
	if s.Flags&FlagSYN != 0 {
		n++
	}
	if s.Flags&FlagFIN != 0 {
		n++
	}
	return n
}
