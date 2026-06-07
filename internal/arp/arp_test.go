package arp

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
)

func TestReply(t *testing.T) {
	localMAC := ethernet.MAC{2, 0, 0, 0, 0, 2}
	localIP := ipv4.Addr{192, 168, 100, 2}
	req := Packet{
		Operation: OpRequest,
		SenderMAC: ethernet.MAC{2, 0, 0, 0, 0, 1},
		SenderIP:  ipv4.Addr{192, 168, 100, 1},
		TargetIP:  localIP,
	}
	rep, ok := Reply(req, localMAC, localIP)
	if !ok {
		t.Fatal("expected reply")
	}
	if rep.Operation != OpReply || rep.SenderMAC != localMAC || rep.SenderIP != localIP || rep.TargetMAC != req.SenderMAC {
		t.Fatalf("unexpected reply: %+v", rep)
	}
}

func TestMarshalParse(t *testing.T) {
	p := Packet{
		Operation: OpRequest,
		SenderMAC: ethernet.MAC{1, 2, 3, 4, 5, 6},
		SenderIP:  ipv4.Addr{10, 0, 0, 1},
		TargetIP:  ipv4.Addr{10, 0, 0, 2},
	}
	raw, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Operation != OpRequest || got.SenderMAC != p.SenderMAC || got.SenderIP != p.SenderIP || got.TargetIP != p.TargetIP {
		t.Fatalf("unexpected packet: %+v", got)
	}
}
