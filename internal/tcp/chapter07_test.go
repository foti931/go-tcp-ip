//go:build chapter07

package tcp

import (
	"testing"

	"tcpip-go/internal/ipv4"
)

func TestChapter07MarshalParseTCP(t *testing.T) {
	src := ipv4.Addr{192, 168, 100, 2}
	dst := ipv4.Addr{192, 168, 100, 1}
	s := Segment{
		SrcPort: 8080,
		DstPort: 50000,
		Seq:     1000,
		Ack:     7001,
		Flags:   FlagSYN | FlagACK,
		Window:  DefaultWin,
	}
	raw, err := Marshal(s, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	if got.SrcPort != s.SrcPort || got.DstPort != s.DstPort || got.Seq != s.Seq || got.Ack != s.Ack || got.Flags != s.Flags {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestChapter07SeqLen(t *testing.T) {
	if got := SeqLen(Segment{Flags: FlagSYN, Payload: nil}); got != 1 {
		t.Fatalf("SYN SeqLen=%d", got)
	}
	if got := SeqLen(Segment{Flags: FlagFIN, Payload: []byte("abc")}); got != 4 {
		t.Fatalf("FIN+payload SeqLen=%d", got)
	}
}
