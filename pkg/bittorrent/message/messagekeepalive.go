package message

var (
	DefaultKeepAliveMessage = Message{
		ID:      MsgKeepAlive,
		Payload: nil,
	}
)

type KeepAliveMessage struct{}
