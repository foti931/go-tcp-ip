package stack

import (
	"io"
	"log"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/hexdump"
	"tcpip-go/internal/ipv4"
)

const (
	UDPEchoPort = 9000
	TCPEchoPort = 8080
)

type Config struct {
	MAC ethernet.MAC
	IP  ipv4.Addr
	Log bool
}

type Stack struct {
	cfg Config
}

func New(cfg Config) *Stack {
	return &Stack{cfg: cfg}
}

func (s *Stack) Run(rw io.ReadWriter) error {
	buf := make([]byte, 1600)
	for {
		n, err := rw.Read(buf)
		if err != nil {
			return err
		}
		reply, err := s.HandleFrame(buf[:n])
		if err != nil {
			log.Printf("drop frame: %v", err)
			continue
		}
		if len(reply) == 0 {
			continue
		}
		if _, err := rw.Write(reply); err != nil {
			return err
		}
	}
}

func (s *Stack) HandleFrame(b []byte) ([]byte, error) {
	if s.cfg.Log {
		log.Printf("rx frame:\n%s", hexdump.Format(b))
	}

	// TODO(chapter-02):
	// ethernet.Parse して EtherType ごとに handleARP / handleIPv4 に振り分ける。
	//
	// TODO(chapter-03):
	// ARP Request に Reply を返す。
	//
	// TODO(chapter-05):
	// ICMP Echo Request に Reply を返す。
	//
	// TODO(chapter-06):
	// UDP port 9000 の payload を echo する。
	//
	// TODO(chapter-07/08/09):
	// TCP port 8080 で handshake / echo / close を扱う。
	return nil, nil
}
