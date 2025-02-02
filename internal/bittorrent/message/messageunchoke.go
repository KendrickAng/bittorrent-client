package message

type UnchokeMessage struct{}

func (m UnchokeMessage) Encode() []byte {
	return createMessageWithPayload(MsgUnchoke, []byte{})
}
