//go:build chapter03

package stack

import (
	"testing"

	"tcpip-go/internal/arp"
	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
)

func TestChapter03StackRepliesToARPRequest(t *testing.T) {
	cfg := Config{
		MAC: ethernet.MAC{0x02, 0, 0, 0, 0, 2},
		IP:  ipv4.Addr{192, 168, 100, 2},
	}
	hostMAC := ethernet.MAC{0x02, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}

	arpb, err := arp.Marshal(arp.Packet{
		HardwareType: arp.HardwareEthernet,
		ProtocolType: arp.ProtocolIPv4,
		HardwareLen:  6,
		ProtocolLen:  4,
		Operation:    arp.OpRequest,
		SenderMAC:    hostMAC,
		SenderIP:     hostIP,
		TargetIP:     cfg.IP,
	})
	if err != nil {
		t.Fatal(err)
	}
	frame, err := ethernet.Marshal(ethernet.Frame{
		Dst:       ethernet.MAC{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		Src:       hostMAC,
		EtherType: ethernet.TypeARP,
		Payload:   arpb,
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
	if gotFrame.Dst != hostMAC || gotFrame.Src != cfg.MAC || gotFrame.EtherType != ethernet.TypeARP {
		t.Fatalf("Ethernet reply mismatch: %+v", gotFrame)
	}
	gotARP, err := arp.Parse(gotFrame.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if gotARP.Operation != arp.OpReply || gotARP.SenderIP != cfg.IP || gotARP.TargetIP != hostIP {
		t.Fatalf("ARP reply mismatch: %+v", gotARP)
	}
}
