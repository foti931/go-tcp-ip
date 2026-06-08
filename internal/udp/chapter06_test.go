//go:build chapter06

package udp

import (
	"testing"

	"tcpip-go/internal/ipv4"
)

func TestChapter06MarshalParseUDP(t *testing.T) {
	src := ipv4.Addr{192, 168, 100, 2}
	dst := ipv4.Addr{192, 168, 100, 1}
	d := Datagram{
		SrcPort: 9000,
		DstPort: 54321,
		Payload: []byte("hello"),
	}
	raw, err := Marshal(d, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw, src, dst, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.SrcPort != d.SrcPort || got.DstPort != d.DstPort || string(got.Payload) != string(d.Payload) {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}
