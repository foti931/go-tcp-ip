//go:build linux

package tap

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

const devicePath = "/dev/net/tun"

type ifreq struct {
	Name  [unix.IFNAMSIZ]byte
	Flags uint16
	_pad  [40 - unix.IFNAMSIZ - 2]byte
}

// Open connects this process to an existing TAP device such as tap0.
// This file is complete from chapter 1; the protocol work starts above it.
func Open(name string) (*os.File, error) {
	fd, err := unix.Open(devicePath, unix.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	req := ifreq{Flags: unix.IFF_TAP | unix.IFF_NO_PI}
	copy(req.Name[:], name)
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(&req))); errno != 0 {
		_ = unix.Close(fd)
		return nil, errno
	}
	return os.NewFile(uintptr(fd), devicePath), nil
}
