//go:build !linux

package tap

import (
	"errors"
	"os"
)

func Open(name string) (*os.File, error) {
	return nil, errors.New("tap: Linux TAP devices are supported only on linux")
}
