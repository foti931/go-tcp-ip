//go:build chapter05

package stack

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/icmp"
	"tcpip-go/internal/ipv4"
)

func TestChapter05StackRepliesToICMPEcho(t *testing.T) {
	cfg := Config{
		MAC: ethernet.MAC{0x02, 0, 0, 0, 0, 2},
		IP:  ipv4.Addr{192, 168, 100, 2},
	}
	hostMAC := ethernet.MAC{0x02, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}

	msg, err := icmp.Marshal(icmp.Message{
		Type:       icmp.TypeEchoRequest,
		Code:       icmp.CodeEcho,
		Identifier: 1,
		Sequence:   2,
		Payload:    []byte("ping"),
	})
	if err != nil {
		t.Fatal(err)
	}
	ipb, err := ipv4.Marshal(ipv4.Packet{
		TTL:      ipv4.DefaultTTL,
		Protocol: ipv4.ProtocolICMP,
		Src:      hostIP,
		Dst:      cfg.IP,
		Payload:  msg,
	})
	if err != nil {
		t.Fatal(err)
	}
	frame, err := ethernet.Marshal(ethernet.Frame{
		Dst:       cfg.MAC,
		Src:       hostMAC,
		EtherType: ethernet.TypeIPv4,
		Payload:   ipb,
	})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := New(cfg).HandleFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	gotFrame, err := ethernet.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	gotIP, err := ipv4.Parse(gotFrame.Payload)
	if err != nil {
		t.Fatal(err)
	}
	gotICMP, err := icmp.Parse(gotIP.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if gotICMP.Type != icmp.TypeEchoReply || gotICMP.Identifier != 1 || gotICMP.Sequence != 2 || string(gotICMP.Payload) != "ping" {
		t.Fatalf("ICMP reply mismatch: %+v", gotICMP)
	}
}
