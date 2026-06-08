package tcp

import "tcpip-go/internal/ipv4"

func Checksum(src, dst ipv4.Addr, segment []byte) uint16 {
	// TODO(chapter-07):
	// IPv4 pseudo header + TCP segment で checksum を計算する。
	return 0
}
