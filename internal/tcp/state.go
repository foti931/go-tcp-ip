package tcp

type State uint8

const (
	StateClosed State = iota
	StateListen
	StateSynReceived
	StateEstablished
	StateCloseWait
	StateLastAck
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateListen:
		return "LISTEN"
	case StateSynReceived:
		return "SYN_RECEIVED"
	case StateEstablished:
		return "ESTABLISHED"
	case StateCloseWait:
		return "CLOSE_WAIT"
	case StateLastAck:
		return "LAST_ACK"
	default:
		return "UNKNOWN"
	}
}
