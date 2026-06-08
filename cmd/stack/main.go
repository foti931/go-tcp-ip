package main

import (
	"log"

	"tcpip-go/internal/ethernet"
	"tcpip-go/internal/ipv4"
	"tcpip-go/internal/stack"
	"tcpip-go/internal/tap"
)

const (
	tapName = "tap0"
)

func main() {
	cfg := stack.Config{
		MAC: ethernet.MAC{0x02, 0x00, 0x00, 0x00, 0x00, 0x02},
		IP:  ipv4.Addr{192, 168, 100, 2},
		Log: true,
	}

	dev, err := tap.Open(tapName)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	log.Printf("stack started: dev=%s mac=%s ip=%s", tapName, cfg.MAC, cfg.IP)
	if err := stack.New(cfg).Run(dev); err != nil {
		log.Fatal(err)
	}
}
