package stack

import (
	"errors"
	"fmt"
	"io"
	"log"

	"tcpip-go/internal/arp"
	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/icmp"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/tcp"
	"tcpip-go/internal/udp"
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
	cfg     Config
	tcpConn tcpConn
	ipID    uint16
}

type tcpConn struct {
	state      tcp.State
	remoteMAC  ethernet.MAC
	remoteIP   ipv4.Addr
	remotePort uint16
	localPort  uint16
	localSeq   uint32
	remoteSeq  uint32
}

func New(cfg Config) *Stack {
	return &Stack{
		cfg: cfg,
		tcpConn: tcpConn{
			state:    tcp.StateListen,
			localSeq: 1000,
		},
	}
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
			if s.cfg.Log {
				log.Printf("drop frame: %v", err)
			}
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
	frame, err := ethernet.Parse(b)
	if err != nil {
		return nil, err
	}
	switch frame.EtherType {
	case ethernet.TypeARP:
		return s.handleARP(frame)
	case ethernet.TypeIPv4:
		return s.handleIPv4(frame)
	default:
		return nil, nil
	}
}

func (s *Stack) handleARP(frame ethernet.Frame) ([]byte, error) {
	req, err := arp.Parse(frame.Payload)
	if err != nil {
		return nil, err
	}
	rep, ok := arp.Reply(req, s.cfg.MAC, s.cfg.IP)
	if !ok {
		return nil, nil
	}
	payload, err := arp.Marshal(rep)
	if err != nil {
		return nil, err
	}
	return ethernet.Marshal(ethernet.Frame{
		Dst:       req.SenderMAC,
		Src:       s.cfg.MAC,
		EtherType: ethernet.TypeARP,
		Payload:   payload,
	})
}

func (s *Stack) handleIPv4(frame ethernet.Frame) ([]byte, error) {
	pkt, err := ipv4.Parse(frame.Payload)
	if err != nil {
		return nil, err
	}
	if pkt.Dst != s.cfg.IP {
		return nil, nil
	}
	switch pkt.Protocol {
	case ipv4.ProtocolICMP:
		return s.handleICMP(frame, pkt)
	case ipv4.ProtocolUDP:
		return s.handleUDP(frame, pkt)
	case ipv4.ProtocolTCP:
		return s.handleTCP(frame, pkt)
	default:
		return nil, nil
	}
}

func (s *Stack) handleICMP(frame ethernet.Frame, pkt ipv4.Packet) ([]byte, error) {
	msg, err := icmp.Parse(pkt.Payload)
	if err != nil {
		return nil, err
	}
	rep, ok := icmp.EchoReply(msg)
	if !ok {
		return nil, nil
	}
	payload, err := icmp.Marshal(rep)
	if err != nil {
		return nil, err
	}
	return s.marshalIPv4(frame.Src, pkt.Src, ipv4.ProtocolICMP, payload)
}

func (s *Stack) handleUDP(frame ethernet.Frame, pkt ipv4.Packet) ([]byte, error) {
	d, err := udp.Parse(pkt.Payload, pkt.Src, pkt.Dst, true)
	if err != nil {
		return nil, err
	}
	if d.DstPort != UDPEchoPort {
		return nil, nil
	}
	payload, err := udp.Marshal(udp.Datagram{
		SrcPort: UDPEchoPort,
		DstPort: d.SrcPort,
		Payload: d.Payload,
	}, s.cfg.IP, pkt.Src)
	if err != nil {
		return nil, err
	}
	return s.marshalIPv4(frame.Src, pkt.Src, ipv4.ProtocolUDP, payload)
}

func (s *Stack) handleTCP(frame ethernet.Frame, pkt ipv4.Packet) ([]byte, error) {
	seg, err := tcp.Parse(pkt.Payload, pkt.Src, pkt.Dst)
	if err != nil {
		return nil, err
	}
	if seg.DstPort != TCPEchoPort {
		return nil, nil
	}
	c := &s.tcpConn
	switch c.state {
	case tcp.StateListen:
		if seg.Flags&tcp.FlagSYN == 0 {
			return nil, nil
		}
		c.remoteMAC = frame.Src
		c.remoteIP = pkt.Src
		c.remotePort = seg.SrcPort
		c.localPort = seg.DstPort
		c.remoteSeq = seg.Seq + 1
		c.state = tcp.StateSynReceived
		return s.replyTCP(tcp.FlagSYN|tcp.FlagACK, nil)
	case tcp.StateSynReceived:
		if !samePeer(c, frame.Src, pkt.Src, seg.SrcPort) {
			return nil, nil
		}
		if seg.Flags&tcp.FlagACK != 0 && seg.Ack == c.localSeq {
			c.state = tcp.StateEstablished
		}
		return nil, nil
	case tcp.StateEstablished:
		if !samePeer(c, frame.Src, pkt.Src, seg.SrcPort) {
			return nil, nil
		}
		if seg.Flags&tcp.FlagFIN != 0 {
			c.remoteSeq = seg.Seq + 1
			c.state = tcp.StateLastAck
			return s.replyTCP(tcp.FlagACK|tcp.FlagFIN, nil)
		}
		if len(seg.Payload) == 0 {
			return nil, nil
		}
		if seg.Seq != c.remoteSeq {
			return s.replyTCP(tcp.FlagACK, nil)
		}
		c.remoteSeq += uint32(len(seg.Payload))
		return s.replyTCP(tcp.FlagACK|tcp.FlagPSH, seg.Payload)
	case tcp.StateCloseWait:
		return nil, nil
	case tcp.StateLastAck:
		if seg.Flags&tcp.FlagACK != 0 {
			c.state = tcp.StateListen
			c.localSeq = 1000
		}
		return nil, nil
	default:
		return nil, errors.New("stack: unknown tcp state")
	}
}

func samePeer(c *tcpConn, mac ethernet.MAC, ip ipv4.Addr, port uint16) bool {
	return c.remoteMAC == mac && c.remoteIP == ip && c.remotePort == port
}

func (s *Stack) replyTCP(flags uint8, payload []byte) ([]byte, error) {
	c := &s.tcpConn
	seg, err := tcp.Marshal(tcp.Segment{
		SrcPort: c.localPort,
		DstPort: c.remotePort,
		Seq:     c.localSeq,
		Ack:     c.remoteSeq,
		Flags:   flags,
		Window:  tcp.DefaultWin,
		Payload: payload,
	}, s.cfg.IP, c.remoteIP)
	if err != nil {
		return nil, err
	}
	advance := uint32(len(payload))
	if flags&tcp.FlagSYN != 0 {
		advance++
	}
	if flags&tcp.FlagFIN != 0 {
		advance++
	}
	c.localSeq += advance
	return s.marshalIPv4(c.remoteMAC, c.remoteIP, ipv4.ProtocolTCP, seg)
}

func (s *Stack) marshalIPv4(dstMAC ethernet.MAC, dstIP ipv4.Addr, proto uint8, payload []byte) ([]byte, error) {
	s.ipID++
	ipb, err := ipv4.Marshal(ipv4.Packet{
		ID:       s.ipID,
		TTL:      ipv4.DefaultTTL,
		Protocol: proto,
		Src:      s.cfg.IP,
		Dst:      dstIP,
		Payload:  payload,
	})
	if err != nil {
		return nil, err
	}
	return ethernet.Marshal(ethernet.Frame{
		Dst:       dstMAC,
		Src:       s.cfg.MAC,
		EtherType: ethernet.TypeIPv4,
		Payload:   ipb,
	})
}

func ConfigFromStrings(mac, ip string) (Config, error) {
	m, err := ethernet.ParseMAC(mac)
	if err != nil {
		return Config{}, err
	}
	a, err := ipv4.ParseAddr(ip)
	if err != nil {
		return Config{}, err
	}
	if a == (ipv4.Addr{}) || m == (ethernet.MAC{}) {
		return Config{}, fmt.Errorf("stack: invalid zero config")
	}
	return Config{MAC: m, IP: a, Log: true}, nil
}
