package ethernet

import "testing"

func TestMarshalParse(t *testing.T) {
	dst := MAC{1, 2, 3, 4, 5, 6}
	src := MAC{6, 5, 4, 3, 2, 1}
	raw, err := Marshal(Frame{Dst: dst, Src: src, EtherType: TypeIPv4, Payload: []byte{0xaa, 0xbb}})
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dst != dst || got.Src != src || got.EtherType != TypeIPv4 || string(got.Payload) != string([]byte{0xaa, 0xbb}) {
		t.Fatalf("unexpected frame: %+v", got)
	}
}

func TestParseShort(t *testing.T) {
	if _, err := Parse([]byte{1, 2, 3}); err == nil {
		t.Fatal("expected error")
	}
}
