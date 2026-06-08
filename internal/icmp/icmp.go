package icmp

import "errors"

const (
	HeaderLen       = 8
	TypeEchoReply   = 0
	TypeEchoRequest = 8
	CodeEcho        = 0
)

var ErrShortPacket = errors.New("icmp: packet too short")

type Message struct {
	Type       uint8
	Code       uint8
	Identifier uint16
	Sequence   uint16
	Payload    []byte
}

func Parse(b []byte) (Message, error) {
	// TODO(chapter-05):
	// Type/Code/Checksum/Identifier/Sequence/Payload を読む。
	return Message{}, ErrShortPacket
}

func Marshal(m Message) ([]byte, error) {
	// TODO(chapter-05):
	// ICMP message を組み立て、checksum を入れる。
	return nil, errors.New("icmp: TODO chapter-05 Marshal")
}

func EchoReply(req Message) (Message, bool) {
	// TODO(chapter-05):
	// Echo Request なら Type 0 の Reply を作る。
	return Message{}, false
}
