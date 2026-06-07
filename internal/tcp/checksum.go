package tcp

import "tcpip-go/internal/ipv4"

func Checksum(src, dst ipv4.Addr, segment []byte) uint16 {
	pseudo := ipv4.PseudoHeader(src, dst, ipv4.ProtocolTCP, uint16(len(segment)))
	buf := make([]byte, 0, len(pseudo)+len(segment))
	buf = append(buf, pseudo...)
	buf = append(buf, segment...)
	return ipv4.Checksum(buf)
}
