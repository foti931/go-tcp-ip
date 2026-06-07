package arp

import (
	"encoding/binary"
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
	if len(b) < PacketLen {
		return Packet{}, ErrShortPacket
	}
	p := Packet{
		HardwareType: binary.BigEndian.Uint16(b[0:2]),
		ProtocolType: binary.BigEndian.Uint16(b[2:4]),
		HardwareLen:  b[4],
		ProtocolLen:  b[5],
		Operation:    binary.BigEndian.Uint16(b[6:8]),
	}
	copy(p.SenderMAC[:], b[8:14])
	copy(p.SenderIP[:], b[14:18])
	copy(p.TargetMAC[:], b[18:24])
	copy(p.TargetIP[:], b[24:28])
	return p, nil
}

func Marshal(p Packet) ([]byte, error) {
	if p.HardwareLen == 0 {
		p.HardwareLen = 6
	}
	if p.ProtocolLen == 0 {
		p.ProtocolLen = 4
	}
	if p.HardwareType == 0 {
		p.HardwareType = HardwareEthernet
	}
	if p.ProtocolType == 0 {
		p.ProtocolType = ProtocolIPv4
	}
	out := make([]byte, PacketLen)
	binary.BigEndian.PutUint16(out[0:2], p.HardwareType)
	binary.BigEndian.PutUint16(out[2:4], p.ProtocolType)
	out[4] = p.HardwareLen
	out[5] = p.ProtocolLen
	binary.BigEndian.PutUint16(out[6:8], p.Operation)
	copy(out[8:14], p.SenderMAC[:])
	copy(out[14:18], p.SenderIP[:])
	copy(out[18:24], p.TargetMAC[:])
	copy(out[24:28], p.TargetIP[:])
	return out, nil
}

func Reply(req Packet, localMAC ethernet.MAC, localIP ipv4.Addr) (Packet, bool) {
	if req.Operation != OpRequest || req.TargetIP != localIP {
		return Packet{}, false
	}
	return Packet{
		HardwareType: HardwareEthernet,
		ProtocolType: ProtocolIPv4,
		HardwareLen:  6,
		ProtocolLen:  4,
		Operation:    OpReply,
		SenderMAC:    localMAC,
		SenderIP:     localIP,
		TargetMAC:    req.SenderMAC,
		TargetIP:     req.SenderIP,
	}, true
}
