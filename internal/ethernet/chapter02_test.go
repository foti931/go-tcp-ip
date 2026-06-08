//go:build chapter02

package ethernet

import "testing"

func TestChapter02ParseEthernetFrame(t *testing.T) {
	raw := []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0x02, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06,
		0xde, 0xad, 0xbe, 0xef,
	}
	f, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f.Dst != (MAC{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}) {
		t.Fatalf("Dst=%s", f.Dst)
	}
	if f.Src != (MAC{0x02, 0, 0, 0, 0, 1}) {
		t.Fatalf("Src=%s", f.Src)
	}
	if f.EtherType != TypeARP {
		t.Fatalf("EtherType=%04x", f.EtherType)
	}
	if string(f.Payload) != string([]byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Fatalf("Payload=%x", f.Payload)
	}
}

func TestChapter02MarshalEthernetFrame(t *testing.T) {
	f := Frame{
		Dst:       MAC{1, 2, 3, 4, 5, 6},
		Src:       MAC{6, 5, 4, 3, 2, 1},
		EtherType: TypeIPv4,
		Payload:   []byte{0xaa, 0xbb},
	}
	raw, err := Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != HeaderLen+2 {
		t.Fatalf("len=%d", len(raw))
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Dst != f.Dst || got.Src != f.Src || got.EtherType != f.EtherType || string(got.Payload) != string(f.Payload) {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}
