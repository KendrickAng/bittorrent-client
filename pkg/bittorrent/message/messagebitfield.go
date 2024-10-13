package message

import "example.com/btclient/pkg/bittorrent"

type BitfieldMessage struct {
	Bitfield bittorrent.Bitfield
}

func (m BitfieldMessage) Decode(msg *Message) *BitfieldMessage {
	if msg.ID != MsgBitfield {
		panic("invalid message bitfield")
	}
	return &BitfieldMessage{
		Bitfield: msg.Payload,
	}
}
