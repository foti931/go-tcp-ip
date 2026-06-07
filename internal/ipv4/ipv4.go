package ipv4

import (
	"encoding/binary"
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
	TOS            uint8
	ID             uint16
	FlagsFragment  uint16
	TTL            uint8
	Protocol       uint8
	Src            Addr
	Dst            Addr
	Options        []byte
	Payload        []byte
	SkipChecksumOK bool
}

func ParseAddr(s string) (Addr, error) {
	var a Addr
	var x [4]uint
	n, err := fmt.Sscanf(s, "%d.%d.%d.%d", &x[0], &x[1], &x[2], &x[3])
	if err != nil || n != 4 {
		return Addr{}, fmt.Errorf("ipv4: invalid address %q", s)
	}
	for i := range a {
		if x[i] > 255 {
			return Addr{}, fmt.Errorf("ipv4: invalid address %q", s)
		}
		a[i] = byte(x[i])
	}
	return a, nil
}

func (a Addr) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", a[0], a[1], a[2], a[3])
}

func Parse(b []byte) (Packet, error) {
	if len(b) < MinHeaderLen {
		return Packet{}, ErrShortPacket
	}
	version := b[0] >> 4
	if version != Version4 {
		return Packet{}, ErrInvalidVersion
	}
	ihl := int(b[0]&0x0f) * 4
	if ihl < MinHeaderLen || ihl > len(b) {
		return Packet{}, ErrInvalidIHL
	}
	totalLen := int(binary.BigEndian.Uint16(b[2:4]))
	if totalLen < ihl || totalLen > len(b) {
		return Packet{}, ErrInvalidLength
	}
	header := b[:ihl]
	if Checksum(header) != 0 {
		return Packet{}, ErrBadChecksum
	}
	p := Packet{
		TOS:           b[1],
		ID:            binary.BigEndian.Uint16(b[4:6]),
		FlagsFragment: binary.BigEndian.Uint16(b[6:8]),
		TTL:           b[8],
		Protocol:      b[9],
		Options:       append([]byte(nil), b[20:ihl]...),
		Payload:       b[ihl:totalLen],
	}
	copy(p.Src[:], b[12:16])
	copy(p.Dst[:], b[16:20])
	return p, nil
}

func Marshal(p Packet) ([]byte, error) {
	if len(p.Options)%4 != 0 {
		return nil, errors.New("ipv4: options length must be multiple of 4")
	}
	ihl := MinHeaderLen + len(p.Options)
	if ihl > 60 {
		return nil, errors.New("ipv4: header too long")
	}
	totalLen := ihl + len(p.Payload)
	if totalLen > 0xffff {
		return nil, errors.New("ipv4: packet too large")
	}
	if p.TTL == 0 {
		p.TTL = DefaultTTL
	}
	out := make([]byte, totalLen)
	out[0] = byte(Version4<<4) | byte(ihl/4)
	out[1] = p.TOS
	binary.BigEndian.PutUint16(out[2:4], uint16(totalLen))
	binary.BigEndian.PutUint16(out[4:6], p.ID)
	binary.BigEndian.PutUint16(out[6:8], p.FlagsFragment)
	out[8] = p.TTL
	out[9] = p.Protocol
	copy(out[12:16], p.Src[:])
	copy(out[16:20], p.Dst[:])
	copy(out[20:ihl], p.Options)
	copy(out[ihl:], p.Payload)
	binary.BigEndian.PutUint16(out[10:12], Checksum(out[:ihl]))
	return out, nil
}

func PseudoHeader(src, dst Addr, protocol uint8, length uint16) []byte {
	out := make([]byte, 12)
	copy(out[0:4], src[:])
	copy(out[4:8], dst[:])
	out[8] = 0
	out[9] = protocol
	binary.BigEndian.PutUint16(out[10:12], length)
	return out
}
