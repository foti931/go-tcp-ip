//go:build chapter06

package stack

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/udp"
)

func TestChapter06StackRepliesToUDPEcho(t *testing.T) {
	cfg := Config{
		MAC: ethernet.MAC{0x02, 0, 0, 0, 0, 2},
		IP:  ipv4.Addr{192, 168, 100, 2},
	}
	hostMAC := ethernet.MAC{0x02, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}
	clientPort := uint16(54321)

	udpb, err := udp.Marshal(udp.Datagram{
		SrcPort: clientPort,
		DstPort: UDPEchoPort,
		Payload: []byte("hello"),
	}, hostIP, cfg.IP)
	if err != nil {
		t.Fatal(err)
	}
	ipb, err := ipv4.Marshal(ipv4.Packet{
		TTL:      ipv4.DefaultTTL,
		Protocol: ipv4.ProtocolUDP,
		Src:      hostIP,
		Dst:      cfg.IP,
		Payload:  udpb,
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
	gotUDP, err := udp.Parse(gotIP.Payload, cfg.IP, hostIP, true)
	if err != nil {
		t.Fatal(err)
	}
	if gotUDP.SrcPort != UDPEchoPort || gotUDP.DstPort != clientPort || string(gotUDP.Payload) != "hello" {
		t.Fatalf("UDP echo mismatch: %+v", gotUDP)
	}
}
