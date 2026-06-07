package ethernet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	HeaderLen     = 14
	TypeIPv4      = 0x0800
	TypeARP       = 0x0806
	BroadcastAddr = "ff:ff:ff:ff:ff:ff"
)

var ErrShortFrame = errors.New("ethernet: frame too short")

type MAC [6]byte

type Frame struct {
	Dst       MAC
	Src       MAC
	EtherType uint16
	Payload   []byte
}

func Parse(b []byte) (Frame, error) {
	if len(b) < HeaderLen {
		return Frame{}, ErrShortFrame
	}
	var f Frame
	copy(f.Dst[:], b[0:6])
	copy(f.Src[:], b[6:12])
	f.EtherType = binary.BigEndian.Uint16(b[12:14])
	f.Payload = b[14:]
	return f, nil
}

func Marshal(f Frame) ([]byte, error) {
	if f.EtherType == 0 {
		return nil, errors.New("ethernet: ethertype is zero")
	}
	out := make([]byte, HeaderLen+len(f.Payload))
	copy(out[0:6], f.Dst[:])
	copy(out[6:12], f.Src[:])
	binary.BigEndian.PutUint16(out[12:14], f.EtherType)
	copy(out[14:], f.Payload)
	return out, nil
}

func ParseMAC(s string) (MAC, error) {
	var m MAC
	var x [6]uint
	n, err := fmt.Sscanf(s, "%02x:%02x:%02x:%02x:%02x:%02x", &x[0], &x[1], &x[2], &x[3], &x[4], &x[5])
	if err != nil || n != 6 {
		return MAC{}, fmt.Errorf("ethernet: invalid mac %q", s)
	}
	for i := range m {
		if x[i] > 0xff {
			return MAC{}, fmt.Errorf("ethernet: invalid mac byte in %q", s)
		}
		m[i] = byte(x[i])
	}
	return m, nil
}

func (m MAC) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
}

func Broadcast() MAC {
	return MAC{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
}
