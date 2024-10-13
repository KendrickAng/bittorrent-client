package message

import "fmt"

const (
	MsgChoke         Type = 0 // no payload
	MsgUnchoke       Type = 1 // no payload
	MsgInterested    Type = 2 // no payload
	MsgNotInterested Type = 3 // no payload
	MsgHave          Type = 4
	MsgBitfield      Type = 5
	MsgRequest       Type = 6
	MsgPiece         Type = 7
	MsgCancel        Type = 8
	MsgKeepAlive     Type = 100 // arbitrary
)

// Type represents the type of peer message.
type Type uint8

func (m Type) String() string {
	switch m {
	case MsgChoke:
		return "choke"
	case MsgUnchoke:
		return "unchoke"
	case MsgInterested:
		return "interested"
	case MsgNotInterested:
		return "not interested"
	case MsgHave:
		return "have"
	case MsgBitfield:
		return "bitfield"
	case MsgRequest:
		return "request"
	case MsgCancel:
		return "cancel"
	case MsgPiece:
		return "piece"
	case MsgKeepAlive:
		return "keep-alive"
	}
	return fmt.Sprintf("unknown(%d)", uint8(m))
}
