//go:build chapter03

package arp

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
)

func TestChapter03ARPReply(t *testing.T) {
	hostMAC := ethernet.MAC{0x02, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}
	localMAC := ethernet.MAC{0x02, 0, 0, 0, 0, 2}
	localIP := ipv4.Addr{192, 168, 100, 2}

	req := Packet{
		HardwareType: HardwareEthernet,
		ProtocolType: ProtocolIPv4,
		HardwareLen:  6,
		ProtocolLen:  4,
		Operation:    OpRequest,
		SenderMAC:    hostMAC,
		SenderIP:     hostIP,
		TargetIP:     localIP,
	}
	rep, ok := Reply(req, localMAC, localIP)
	if !ok {
		t.Fatal("expected ARP reply")
	}
	if rep.Operation != OpReply {
		t.Fatalf("Operation=%d", rep.Operation)
	}
	if rep.SenderMAC != localMAC || rep.SenderIP != localIP {
		t.Fatalf("sender mismatch: %+v", rep)
	}
	if rep.TargetMAC != hostMAC || rep.TargetIP != hostIP {
		t.Fatalf("target mismatch: %+v", rep)
	}
}

func TestChapter03MarshalParseARP(t *testing.T) {
	p := Packet{
		HardwareType: HardwareEthernet,
		ProtocolType: ProtocolIPv4,
		HardwareLen:  6,
		ProtocolLen:  4,
		Operation:    OpRequest,
		SenderMAC:    ethernet.MAC{0x02, 0, 0, 0, 0, 1},
		SenderIP:     ipv4.Addr{192, 168, 100, 1},
		TargetIP:     ipv4.Addr{192, 168, 100, 2},
	}
	raw, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Operation != p.Operation || got.SenderMAC != p.SenderMAC || got.SenderIP != p.SenderIP || got.TargetIP != p.TargetIP {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}
