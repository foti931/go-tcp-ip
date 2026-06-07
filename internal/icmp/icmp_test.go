package icmp

import "testing"

func TestEchoReply(t *testing.T) {
	req := Message{Type: TypeEchoRequest, Code: CodeEcho, Identifier: 1, Sequence: 2, Payload: []byte("abc")}
	rep, ok := EchoReply(req)
	if !ok {
		t.Fatal("expected reply")
	}
	if rep.Type != TypeEchoReply || rep.Identifier != req.Identifier || rep.Sequence != req.Sequence || string(rep.Payload) != "abc" {
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
	if got.Type != TypeEchoReply || string(got.Payload) != "abc" {
		t.Fatalf("unexpected parsed message: %+v", got)
	}
}
