package tcp

type State uint8

const (
	StateClosed State = iota
	StateListen
	StateSynReceived
	StateEstablished
	StateLastAck
)
