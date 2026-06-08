//go:build chapter09

package stack

import (
	"testing"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/tcp"
)

func TestChapter09TCPCloseReturnsToListen(t *testing.T) {
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

	_ = sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 1,
		Ack:     synAck.Seq + 1,
		Flags:   tcp.FlagACK,
	})

	finRaw := sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 1,
		Ack:     synAck.Seq + 1,
		Flags:   tcp.FlagACK | tcp.FlagFIN,
	})
	fin := parseTCPReplyForChapterTest(t, finRaw, cfg.IP, hostIP)
	if fin.Flags != tcp.FlagACK|tcp.FlagFIN {
		t.Fatalf("Flags=%02x", fin.Flags)
	}
	if fin.Ack != clientSeq+2 {
		t.Fatalf("FIN の ACK は clientSeq+2: got %d", fin.Ack)
	}

	_ = sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort,
		DstPort: TCPEchoPort,
		Seq:     clientSeq + 2,
		Ack:     fin.Seq + 1,
		Flags:   tcp.FlagACK,
	})

	nextSynRaw := sendTCPForChapterTest(t, st, cfg, hostMAC, hostIP, tcp.Segment{
		SrcPort: clientPort + 1,
		DstPort: TCPEchoPort,
		Seq:     9000,
		Flags:   tcp.FlagSYN,
	})
	nextSynAck := parseTCPReplyForChapterTest(t, nextSynRaw, cfg.IP, hostIP)
	if nextSynAck.Flags != tcp.FlagSYN|tcp.FlagACK {
		t.Fatalf("close 後に LISTEN に戻っていない: %+v", nextSynAck)
	}
}
