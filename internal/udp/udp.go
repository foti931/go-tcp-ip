package udp

import (
	"errors"

	"tcpip-go/internal/ipv4"
)

const HeaderLen = 8

var ErrShortPacket = errors.New("udp: packet too short")

type Datagram struct {
	SrcPort uint16
	DstPort uint16
	Payload []byte
}

func Parse(b []byte, src, dst ipv4.Addr, verifyChecksum bool) (Datagram, error) {
	// TODO(chapter-06):
	// SrcPort/DstPort/Length/Checksum/Payload を読む。
	return Datagram{}, ErrShortPacket
}

func Marshal(d Datagram, src, dst ipv4.Addr) ([]byte, error) {
	// TODO(chapter-06):
	// UDP header と checksum を作る。
	return nil, errors.New("udp: TODO chapter-06 Marshal")
}
