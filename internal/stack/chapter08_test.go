//go:build chapter08

package stack

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/tcp"
)

func TestChapter08TCPEcho(t *testing.T) {
	cfg := Config{
		MAC: ethernet.MAC{0x02, 0, 0, 0, 0, 2},
		IP:  ipv4.Addr{192, 168, 100, 2},
	}
	st := New(cfg)
	hostMAC := ethernet.MAC{0x02, 0, 0, 0, 0, 1}
	hostIP := ipv4.Addr{192, 168, 100, 1}
	clientPort := uint16(50000)
	clientSeq := uint32(7000)

	synAckRaw := sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq,
		Flags:   tcp.FlagSYN,
	})
	synAck := parseTCPReplyForChapterTest(t, synAckRaw, cfg.IP, hostIP)

	ackRaw := sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 1,
		Ack:     synAck.Seq + 1,
		Flags:   tcp.FlagACK,
	})
	if len(ackRaw) != 0 {
		t.Fatalf("handshake 最後の ACK には返信しない: %d bytes", len(ackRaw))
	}

	echoRaw := sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 1,
		Ack:     synAck.Seq + 1,
		Flags:   tcp.FlagACK | tcp.FlagPSH,
		Payload: []byte("hello"),
	})
	echo := parseTCPReplyForChapterTest(t, echoRaw, cfg.IP, hostIP)
	if echo.Flags != tcp.FlagACK|tcp.FlagPSH {
		t.Fatalf("Flags=%02x", echo.Flags)
	}
	if echo.Ack != clientSeq+1+uint32(len("hello")) {
		t.Fatalf("Ack=%d", echo.Ack)
	}
	if string(echo.Payload) != "hello" {
		t.Fatalf("Payload=%q", string(echo.Payload))
	}
}
