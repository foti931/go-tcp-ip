package tcp

import (
	"errors"

	"tcpip-go/internal/ipv4"
)

const (
	MinHeaderLen = 20
	FlagFIN      = 0x01
	FlagSYN      = 0x02
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
	Options []byte
	Payload []byte
}

func Parse(b []byte, src, dst ipv4.Addr) (Segment, error) {
	// TODO(chapter-07):
	// TCP header を parse し、checksum を検証する。
	return Segment{}, ErrShortSegment
}

func Marshal(s Segment, src, dst ipv4.Addr) ([]byte, error) {
	// TODO(chapter-07):
	// TCP segment を組み立て、checksum を入れる。
	return nil, errors.New("tcp: TODO chapter-07 Marshal")
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
