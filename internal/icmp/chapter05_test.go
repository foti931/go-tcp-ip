//go:build chapter05

package icmp

import "testing"

func TestChapter05EchoReply(t *testing.T) {
	req := Message{
		Type:       TypeEchoRequest,
		Code:       CodeEcho,
		Identifier: 123,
		Sequence:   7,
		Payload:    []byte("hello"),
	}
	rep, ok := EchoReply(req)
	if !ok {
		t.Fatal("expected echo reply")
	}
	if rep.Type != TypeEchoReply || rep.Code != CodeEcho || rep.Identifier != req.Identifier || rep.Sequence != req.Sequence || string(rep.Payload) != "hello" {
		t.Fatalf("unexpected reply: %+v", rep)
	}
	raw, err := Marshal(rep)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Type != TypeEchoReply || got.Identifier != req.Identifier || got.Sequence != req.Sequence || string(got.Payload) != "hello" {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}
