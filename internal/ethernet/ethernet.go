package ethernet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	HeaderLen = 14
	TypeIPv4  = 0x0800
	TypeARP   = 0x0806
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
	// TODO(chapter-02):
	// 1. len(b) が HeaderLen 未満なら ErrShortFrame
	// 2. b[0:6] を Dst、b[6:12] を Src に copy
	// 3. binary.BigEndian.Uint16(b[12:14]) を EtherType にする
	// 4. b[14:] を Payload にする
	return Frame{}, ErrShortFrame
}

func Marshal(f Frame) ([]byte, error) {
	// TODO(chapter-02):
	// HeaderLen + len(f.Payload) の buffer を作り、
	// Dst/Src/EtherType/Payload を Ethernet header 順に詰める。
	return nil, errors.New("ethernet: TODO chapter-02 Marshal")
}

func MustMAC(s string) MAC {
	var m MAC
	var x [6]uint
	n, err := fmt.Sscanf(s, "%02x:%02x:%02x:%02x:%02x:%02x", &x[0], &x[1], &x[2], &x[3], &x[4], &x[5])
	if err != nil || n != 6 {
		panic(err)
	}
	for i := range m {
		m[i] = byte(x[i])
	}
	return m
}

func (m MAC) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", m[0], m[1], m[2], m[3], m[4], m[5])
}

func PutEtherType(b []byte, typ uint16) {
	binary.BigEndian.PutUint16(b, typ)
}
