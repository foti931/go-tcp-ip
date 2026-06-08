//go:build chapter04

package ipv4

import "testing"

func TestChapter04ChecksumValidatesHeader(t *testing.T) {
	header := []byte{
		0x45, 0x00, 0x00, 0x54,
		0x00, 0x00, 0x40, 0x00,
		0x40, 0x01, 0x00, 0x00,
		0xc0, 0xa8, 0x64, 0x01,
		0xc0, 0xa8, 0x64, 0x02,
	}
	sum := Checksum(header)
	header[10] = byte(sum >> 8)
	header[11] = byte(sum)
	if Checksum(header) != 0 {
		t.Fatalf("checksum did not validate: %04x", Checksum(header))
	}
}

func TestChapter04MarshalParseIPv4(t *testing.T) {
	p := Packet{
		ID:       10,
		TTL:      DefaultTTL,
		Protocol: ProtocolICMP,
		Src:      Addr{192, 168, 100, 2},
		Dst:      Addr{192, 168, 100, 1},
		Payload:  []byte{1, 2, 3, 4},
	}
	raw, err := Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != p.ID || got.Protocol != p.Protocol || got.Src != p.Src || got.Dst != p.Dst || string(got.Payload) != string(p.Payload) {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}
