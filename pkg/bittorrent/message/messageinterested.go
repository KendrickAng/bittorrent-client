package message

type InterestedMessage struct{}

func (m InterestedMessage) Encode() []byte {
	return createMessageWithPayload(MsgInterested, []byte{})
}
