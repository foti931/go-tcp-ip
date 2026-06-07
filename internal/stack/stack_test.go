package stack

import (
	"testing"

	"tcpip-go/internal/arp"
	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/icmp"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/tcp"
)

func TestHandleARP(t *testing.T) {
	cfg, err := ConfigFromStrings("02:00:00:00:00:02", "192.168.100.2")
	if err != nil {
		t.Fatal(err)
	}
	hostMAC := ethernet.MAC{2, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}
	req, _ := arp.Marshal(arp.Packet{Operation: arp.OpRequest, SenderMAC: hostMAC, SenderIP: hostIP, TargetIP: cfg.IP})
	frame, _ := ethernet.Marshal(ethernet.Frame{Dst: ethernet.Broadcast(), Src: hostMAC, EtherType: ethernet.TypeARP, Payload: req})
	raw, err := New(cfg).HandleFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	gotFrame, err := ethernet.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	gotARP, err := arp.Parse(gotFrame.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if gotARP.Operation != arp.OpReply || gotARP.SenderIP != cfg.IP || gotARP.TargetIP != hostIP {
		t.Fatalf("unexpected arp reply: %+v", gotARP)
	}
}

func TestHandleICMP(t *testing.T) {
	cfg, err := ConfigFromStrings("02:00:00:00:00:02", "192.168.100.2")
	if err != nil {
		t.Fatal(err)
	}
	hostMAC := ethernet.MAC{2, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}
	msg, _ := icmp.Marshal(icmp.Message{Type: icmp.TypeEchoRequest, Identifier: 1, Sequence: 1, Payload: []byte("ping")})
	ipb, _ := ipv4.Marshal(ipv4.Packet{Protocol: ipv4.ProtocolICMP, Src: hostIP, Dst: cfg.IP, Payload: msg})
	frame, _ := ethernet.Marshal(ethernet.Frame{Dst: cfg.MAC, Src: hostMAC, EtherType: ethernet.TypeIPv4, Payload: ipb})
	raw, err := New(cfg).HandleFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	repFrame, _ := ethernet.Parse(raw)
	repIP, err := ipv4.Parse(repFrame.Payload)
	if err != nil {
		t.Fatal(err)
	}
	repICMP, err := icmp.Parse(repIP.Payload)
	if err != nil {
		t.Fatal(err)
	}
	if repICMP.Type != icmp.TypeEchoReply || string(repICMP.Payload) != "ping" {
		t.Fatalf("unexpected icmp reply: %+v", repICMP)
	}
}

func TestTCPHandshakeEchoAndClose(t *testing.T) {
	cfg, err := ConfigFromStrings("02:00:00:00:00:02", "192.168.100.2")
	if err != nil {
		t.Fatal(err)
	}
	st := New(cfg)
	hostMAC := ethernet.MAC{2, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}
	clientPort := uint16(50000)
	clientSeq := uint32(7000)

	synAckRaw := sendTCP(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq,
		Flags:   tcp.FlagSYN,
	})
	synAck := parseTCPReply(t, synAckRaw, cfg.IP, hostIP)
	if synAck.Flags != tcp.FlagSYN|tcp.FlagACK || synAck.Ack != clientSeq+1 {
		t.Fatalf("unexpected syn-ack: %+v", synAck)
	}

	ackRaw := sendTCP(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 1,
		Ack:     synAck.Seq + 1,
		Flags:   tcp.FlagACK,
	})
	if len(ackRaw) != 0 {
		t.Fatalf("expected no reply to final handshake ack, got %d bytes", len(ackRaw))
	}

	echoRaw := sendTCP(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 1,
		Ack:     synAck.Seq + 1,
		Flags:   tcp.FlagACK | tcp.FlagPSH,
		Payload: []byte("hi"),
	})
	echo := parseTCPReply(t, echoRaw, cfg.IP, hostIP)
	if echo.Flags != tcp.FlagACK|tcp.FlagPSH || echo.Ack != clientSeq+3 || string(echo.Payload) != "hi" {
		t.Fatalf("unexpected echo: %+v", echo)
	}

	finRaw := sendTCP(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 3,
		Ack:     echo.Seq + uint32(len(echo.Payload)),
		Flags:   tcp.FlagACK | tcp.FlagFIN,
	})
	fin := parseTCPReply(t, finRaw, cfg.IP, hostIP)
	if fin.Flags != tcp.FlagACK|tcp.FlagFIN || fin.Ack != clientSeq+4 {
		t.Fatalf("unexpected fin reply: %+v", fin)
	}
}

func sendTCP(t *testing.T, st *Stack, cfg Config, hostMAC ethernet.MAC, hostIP ipv4.Addr, seg tcp.Segment) []byte {
	t.Helper()
	tcpb, err := tcp.Marshal(seg, hostIP, cfg.IP)
	if err != nil {
		t.Fatal(err)
	}
	ipb, err := ipv4.Marshal(ipv4.Packet{Protocol: ipv4.ProtocolTCP, Src: hostIP, Dst: cfg.IP, Payload: tcpb})
	if err != nil {
		t.Fatal(err)
	}
	frame, err := ethernet.Marshal(ethernet.Frame{Dst: cfg.MAC, Src: hostMAC, EtherType: ethernet.TypeIPv4, Payload: ipb})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := st.HandleFrame(frame)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func parseTCPReply(t *testing.T, raw []byte, src, dst ipv4.Addr) tcp.Segment {
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
