package ethernet

import "testing"

func TestParseShortFrame(t *testing.T) {
	if _, err := Parse([]byte{1, 2, 3}); err != ErrShortFrame {
		t.Fatalf("err=%v", err)
	}
}

func TestChapter02MarshalParseRoundTrip(t *testing.T) {
	t.Skip("chapter-02: TODO を実装したら Skip を外す")
}
