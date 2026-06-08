package ipv4

import (
	"errors"
	"fmt"
)

const (
	MinHeaderLen = 20
	Version4     = 4
	ProtocolICMP = 1
	ProtocolTCP  = 6
	ProtocolUDP  = 17
	DefaultTTL   = 64
)

var (
	ErrShortPacket    = errors.New("ipv4: packet too short")
	ErrInvalidVersion = errors.New("ipv4: invalid version")
	ErrInvalidIHL     = errors.New("ipv4: invalid ihl")
	ErrInvalidLength  = errors.New("ipv4: invalid total length")
	ErrBadChecksum    = errors.New("ipv4: bad header checksum")
)

type Addr [4]byte

type Packet struct {
	ID            uint16
	FlagsFragment uint16
	TTL           uint8
	Protocol      uint8
	Src           Addr
	Dst           Addr
	Payload       []byte
}

func Parse(b []byte) (Packet, error) {
	// TODO(chapter-04):
	// Version/IHL/TotalLength/HeaderChecksum/SrcIP/DstIP/Payload を読む。
	return Packet{}, ErrShortPacket
}

func Marshal(p Packet) ([]byte, error) {
	// TODO(chapter-04):
	// IPv4 header を組み立て、Header Checksum を入れる。
	return nil, errors.New("ipv4: TODO chapter-04 Marshal")
}

func PseudoHeader(src, dst Addr, protocol uint8, length uint16) []byte {
	// TODO(chapter-06/07):
	// UDP/TCP checksum 用の 12 byte pseudo header を返す。
	return nil
}

func (a Addr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3])
}
