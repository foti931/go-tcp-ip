package udp

import (
	"encoding/binary"
	"errors"

	"tcpip-go/internal/ipv4"
)

const HeaderLen = 8

var ErrShortPacket = errors.New("udp: packet too short")

type Datagram struct {
	SrcPort  uint16
	DstPort  uint16
	Checksum uint16
	Payload  []byte
}

func Parse(b []byte, src, dst ipv4.Addr, verify bool) (Datagram, error) {
	if len(b) < HeaderLen {
		return Datagram{}, ErrShortPacket
	}
	length := int(binary.BigEndian.Uint16(b[4:6]))
	if length < HeaderLen || length > len(b) {
		return Datagram{}, errors.New("udp: invalid length")
	}
	check := binary.BigEndian.Uint16(b[6:8])
	if verify && check != 0 && transportChecksum(src, dst, ipv4.ProtocolUDP, b[:length]) != 0 {
		return Datagram{}, errors.New("udp: bad checksum")
	}
	return Datagram{
		SrcPort:  binary.BigEndian.Uint16(b[0:2]),
		DstPort:  binary.BigEndian.Uint16(b[2:4]),
		Checksum: check,
		Payload:  b[HeaderLen:length],
	}, nil
}

func Marshal(d Datagram, src, dst ipv4.Addr) ([]byte, error) {
	length := HeaderLen + len(d.Payload)
	if length > 0xffff {
		return nil, errors.New("udp: datagram too large")
	}
	out := make([]byte, length)
	binary.BigEndian.PutUint16(out[0:2], d.SrcPort)
	binary.BigEndian.PutUint16(out[2:4], d.DstPort)
	binary.BigEndian.PutUint16(out[4:6], uint16(length))
	copy(out[HeaderLen:], d.Payload)
	sum := transportChecksum(src, dst, ipv4.ProtocolUDP, out)
	if sum == 0 {
		sum = 0xffff
	}
	binary.BigEndian.PutUint16(out[6:8], sum)
	return out, nil
}

func transportChecksum(src, dst ipv4.Addr, proto uint8, segment []byte) uint16 {
	pseudo := ipv4.PseudoHeader(src, dst, proto, uint16(len(segment)))
	buf := make([]byte, 0, len(pseudo)+len(segment))
	buf = append(buf, pseudo...)
	buf = append(buf, segment...)
	return ipv4.Checksum(buf)
}
