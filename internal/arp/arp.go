package arp

import (
	"errors"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
)

const (
	PacketLen        = 28
	HardwareEthernet = 1
	ProtocolIPv4     = 0x0800
	OpRequest        = 1
	OpReply          = 2
)

var ErrShortPacket = errors.New("arp: packet too short")

type Packet struct {
	HardwareType uint16
	ProtocolType uint16
	HardwareLen  uint8
	ProtocolLen  uint8
	Operation    uint16
	SenderMAC    ethernet.MAC
	SenderIP     ipv4.Addr
	TargetMAC    ethernet.MAC
	TargetIP     ipv4.Addr
}

func Parse(b []byte) (Packet, error) {
	// TODO(chapter-03):
	// ARP は Ethernet/IPv4 なら 28 bytes。
	// binary.BigEndian で各 field を読む。
	return Packet{}, ErrShortPacket
}

func Marshal(p Packet) ([]byte, error) {
	// TODO(chapter-03):
	// HardwareType/ProtocolType/HLen/PLen の default を補い、28 bytes に詰める。
	return nil, errors.New("arp: TODO chapter-03 Marshal")
}

func Reply(req Packet, localMAC ethernet.MAC, localIP ipv4.Addr) (Packet, bool) {
	// TODO(chapter-03):
	// Operation が Request で TargetIP が localIP のときだけ Reply を返す。
	return Packet{}, false
}
