package tcp

import (
	"testing"

	"tcpip-go/internal/ipv4"
)

func TestMarshalParse(t *testing.T) {
	src := ipv4.Addr{192, 168, 100, 2}
	dst := ipv4.Addr{192, 168, 100, 1}
	raw, err := Marshal(Segment{SrcPort: 8080, DstPort: 55555, Seq: 10, Ack: 20, Flags: FlagACK | FlagPSH, Payload: []byte("ok")}, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw, src, dst)
	if err != nil {
		t.Fatal(err)
	}
	if got.SrcPort != 8080 || got.DstPort != 55555 || got.Seq != 10 || got.Ack != 20 || got.Flags != FlagACK|FlagPSH || string(got.Payload) != "ok" {
		t.Fatalf("unexpected segment: %+v", got)
	}
}

func TestSeqLen(t *testing.T) {
	if got := SeqLen(Segment{Flags: FlagSYN | FlagFIN, Payload: []byte("abc")}); got != 5 {
		t.Fatalf("SeqLen=%d", got)
	}
}
