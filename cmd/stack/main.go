package main

import (
	"log"

	"tcpip-go/internal/stack"
	"tcpip-go/internal/tap"
)

const (
	tapName  = "tap0"
	localMAC = "02:00:00:00:00:02"
	localIP  = "192.168.100.2"
)

func main() {
	cfg, err := stack.ConfigFromStrings(localMAC, localIP)
	if err != nil {
		log.Fatal(err)
	}
	dev, err := tap.Open(tapName)
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Printf("stack started: dev=%s mac=%s ip=%s udp=%d tcp=%d", tapName, cfg.MAC, cfg.IP, stack.UDPEchoPort, stack.TCPEchoPort)
	if err := stack.New(cfg).Run(dev); err != nil {
		log.Fatal(err)
	}
}
