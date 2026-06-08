package hexdump

import "encoding/hex"

func Format(b []byte) string {
	return hex.Dump(b)
}
