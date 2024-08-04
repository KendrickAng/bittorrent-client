package trackerprotocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// All non-keepalive messages start with a single byte.
type messageID uint8

func (m messageID) String() string {
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
	}
	return fmt.Sprintf("unknown(%d)", uint8(m))
}

const (
	MsgChoke         messageID = 0 // no payload
	MsgUnchoke       messageID = 1 // no payload
	MsgInterested    messageID = 2 // no payload
	MsgNotInterested messageID = 3 // no payload
	MsgHave          messageID = 4
	MsgBitfield      messageID = 5
	MsgRequest       messageID = 6
	MsgPiece         messageID = 7
	MsgCancel        messageID = 8
	MsgKeepAlive     messageID = 100 // arbitrary
)

// Message contains the ID and payload of a message
type Message struct {
	ID      messageID
	Payload []byte
}

func (m *Message) Serialize() []byte {
	if m.ID == MsgKeepAlive {
		return make([]byte, 0)
	}

	numBytes := 4 + 1 + len(m.Payload) // length + type + payload
	buffer := make([]byte, numBytes)

	// Total length of message = payload + id
	binary.BigEndian.PutUint32(buffer[:4], uint32(len(m.Payload)+1))

	// Message type
	buffer[4] = byte(m.ID)

	// Message payload
	copy(buffer[5:], m.Payload)

	return buffer
}

func (m *Message) String() string {
	return fmt.Sprintf("ID: %s, Payload: %d bytes\n", m.ID.String(), len(m.Payload))
}

func Deserialize(r io.Reader) (*Message, error) {
	// Read length of message
	lengthBuffer := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuffer)
	if err != nil {
		return nil, err
	}

	// Get length of message
	length := binary.BigEndian.Uint32(lengthBuffer[:])

	// KeepAlive messages have a length of zero
	if length == 0 {
		return mKeepAlive(), nil
	}

	// Non-keepalive messages start with a single byte (msg type)
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      messageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}, nil
}

func mKeepAlive() *Message {
	return &Message{
		ID:      MsgKeepAlive,
		Payload: nil,
	}
}
