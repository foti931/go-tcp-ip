package udp

import (
	"testing"

	"tcpip-go/internal/ipv4"
)

func TestMarshalParse(t *testing.T) {
	src := ipv4.Addr{192, 168, 100, 2}
	dst := ipv4.Addr{192, 168, 100, 1}
	raw, err := Marshal(Datagram{SrcPort: 9000, DstPort: 54321, Payload: []byte("hello")}, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw, src, dst, true)
	if err != nil {
		t.Fatal(err)
	}
	if got.SrcPort != 9000 || got.DstPort != 54321 || string(got.Payload) != "hello" {
		t.Fatalf("unexpected datagram: %+v", got)
	}
}
