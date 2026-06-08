//go:build chapter08 || chapter09

package stack

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/tcp"
)

func sendTCPForChapterTest(t *testing.T, st *Stack, cfg Config, hostMAC ethernet.MAC, hostIP ipv4.Addr, seg tcp.Segment) []byte {
	t.Helper()
	tcpb, err := tcp.Marshal(seg, hostIP, cfg.IP)
	if err != nil {
		t.Fatal(err)
	}
	ipb, err := ipv4.Marshal(ipv4.Packet{
		TTL:      ipv4.DefaultTTL,
		Protocol: ipv4.ProtocolTCP,
		Src:      hostIP,
		Dst:      cfg.IP,
		Payload:  tcpb,
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
	raw, err := st.HandleFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func parseTCPReplyForChapterTest(t *testing.T, raw []byte, src, dst ipv4.Addr) tcp.Segment {
	t.Helper()
	frame, err := ethernet.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	ipb, err := ipv4.Parse(frame.Payload)
	if err != nil {
		t.Fatal(err)
	}
	seg, err := tcp.Parse(ipb.Payload, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	return seg
}
