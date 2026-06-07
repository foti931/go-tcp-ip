package hexdump

import (
	"encoding/hex"
	"fmt"
	"strings"
)

func Format(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	lines := strings.Split(strings.TrimRight(hex.Dump(b), "\n"), "\n")
	for i, line := range lines {
		lines[i] = fmt.Sprintf("%s", line)
	}
	return strings.Join(lines, "\n")
}
