//go:build linux

package tap

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

const devicePath = "/dev/net/tun"

// ifreq is the small C-compatible request object passed to TUNSETIFF.
//
// Linux expects roughly this C layout:
//
//	struct ifreq {
//	    char  ifr_name[IFNAMSIZ];
//	    short ifr_flags;
//	    ...
//	};
//
// We only use the interface name and flags. The padding keeps the Go struct
// large enough for the kernel's ifreq layout on Linux.
type ifreq struct {
	Name  [unix.IFNAMSIZ]byte
	Flags uint16
	_pad  [40 - unix.IFNAMSIZ - 2]byte
}

// Open connects this process to an existing TAP device such as tap0.
// This file is complete from chapter 1; the protocol work starts above it.
func Open(name string) (*os.File, error) {
	// /dev/net/tun is the control device for both TUN and TAP interfaces.
	// Opening it gives us a file descriptor, but it is not attached to tap0 yet.
	fd, err := unix.Open(devicePath, unix.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	// IFF_TAP means "give me Ethernet frames" instead of IP packets.
	// IFF_NO_PI means "do not prepend Linux's 4-byte packet information header".
	// With IFF_NO_PI, each Read returns bytes that start at the Ethernet header:
	//
	//	DstMAC(6) SrcMAC(6) EtherType(2) Payload(...)
	req := ifreq{Flags: unix.IFF_TAP | unix.IFF_NO_PI}
	copy(req.Name[:], name)

	// TUNSETIFF binds this file descriptor to the named interface.
	// unsafe.Pointer is used only here because ioctl requires passing the address
	// of the C-compatible ifreq memory block to the kernel.
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(&req))); errno != 0 {
		_ = unix.Close(fd)
		return nil, errno
	}

	// After TUNSETIFF succeeds, Read and Write on this os.File exchange Ethernet
	// frames with tap0.
	return os.NewFile(uintptr(fd), devicePath), nil
}
