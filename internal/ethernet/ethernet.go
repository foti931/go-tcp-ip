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
	//
	// Ethernet frame の先頭 14 byte は Ethernet header です。
	//
	// offset   slice      bytes  意味
	// 0        b[0:6]     6      宛先 MAC address
	// 6        b[6:12]    6      送信元 MAC address
	// 12       b[12:14]   2      EtherType
	// 14       b[14:]     可変   次の protocol の中身
	//
	// 例:
	//   ff ff ff ff ff ff  02 00 00 00 00 01  08 06  ...
	//   |宛先 broadcast |  |送信元 MAC      |  |ARP |  |ARP packet
	//
	// EtherType は 2 byte の整数なので binary.BigEndian で読みます。
	// MAC address は [6]byte なので copy で配列へ移します。
	return Frame{}, ErrShortFrame
}

func Marshal(f Frame) ([]byte, error) {
	// TODO(chapter-02):
	//
	// Parse の逆です。HeaderLen + len(f.Payload) の byte slice を作り、
	// 次の順番で詰めます。
	//
	// out[0:6]   = 宛先 MAC
	// out[6:12]  = 送信元 MAC
	// out[12:14] = EtherType
	// out[14:]   = Payload
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
