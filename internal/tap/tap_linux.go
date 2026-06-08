//go:build linux

package tap

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

const devicePath = "/dev/net/tun"

// ifreq は TUNSETIFF ioctl に渡す Linux 用の request 構造体です。
//
// Linux kernel は、おおよそ次の C の struct ifreq と同じメモリ配置を期待します。
//
//	struct ifreq {
//	    char  ifr_name[IFNAMSIZ];
//	    short ifr_flags;
//	    ...
//	};
//
// この教材では interface 名と flags だけを使います。
// _pad は kernel が読む ifreq のサイズに合わせるための詰め物です。
type ifreq struct {
	Name  [unix.IFNAMSIZ]byte
	Flags uint16
	_pad  [40 - unix.IFNAMSIZ - 2]byte
}

// Open は、このプロセスを tap0 のような TAP device に接続します。
// 第1章ではこの関数を入口にして、TAP が Ethernet frame を返すことを確認します。
func Open(name string) (*os.File, error) {
	// /dev/net/tun は TUN/TAP 共通の制御 device です。
	// open した直後の fd は、まだ tap0 には紐づいていません。
	fd, err := unix.Open(devicePath, unix.O_RDWR, 0)
	if err != nil {
		return nil, err
	}

	// IFF_TAP は「IP packet ではなく Ethernet frame を読み書きする」という指定です。
	// IFF_NO_PI は Linux 独自の 4 byte packet information header を付けない指定です。
	// これにより Read の結果は Ethernet header から始まります。
	//
	//	DstMAC(6) SrcMAC(6) EtherType(2) Payload(...)
	req := ifreq{Flags: unix.IFF_TAP | unix.IFF_NO_PI}
	copy(req.Name[:], name)

	// TUNSETIFF は fd と指定した interface 名を結びつける ioctl です。
	// ioctl は C 互換のメモリブロックのアドレスを kernel に渡す API なので、
	// この 1 箇所だけ unsafe.Pointer を使います。
	if _, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(unix.TUNSETIFF), uintptr(unsafe.Pointer(&req))); errno != 0 {
		_ = unix.Close(fd)
		return nil, errno
	}

	// ここから先は、この os.File に対する Read/Write が tap0 の Ethernet frame
	// の受信/送信になります。
	return os.NewFile(uintptr(fd), devicePath), nil
}
