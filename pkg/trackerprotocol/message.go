package trackerprotocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

// All non-keepalive messages start with a single byte.
type MessageID uint8

func (m MessageID) String() string {
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
	MsgChoke         MessageID = 0 // no payload
	MsgUnchoke       MessageID = 1 // no payload
	MsgInterested    MessageID = 2 // no payload
	MsgNotInterested MessageID = 3 // no payload
	MsgHave          MessageID = 4
	MsgBitfield      MessageID = 5
	MsgRequest       MessageID = 6
	MsgPiece         MessageID = 7
	MsgCancel        MessageID = 8
	MsgKeepAlive     MessageID = 100 // arbitrary

	MaxRequestLength = 16384 // 2 ^ 14 (16kiB)
)

// Message contains the ID and payload of a message
type Message struct {
	ID      MessageID
	Payload []byte
	// Total requestLength of the message in bytes.
	Length uint32
}

type MessageUnchoke struct{}

func (m MessageUnchoke) Encode() []byte {
	return createMessageWithPayload(MsgUnchoke, []byte{})
}

type MessageInterested struct{}

func (m MessageInterested) Encode() []byte {
	return createMessageWithPayload(MsgInterested, []byte{})
}

type MessageBitfield struct {
	Bitfield Bitfield
}

type MessageRequest struct {
	// The zero-based piece pieceIndex.
	Index uint32

	// The zero-based byte offset within the piece.
	Begin uint32

	// The requested requestLength of the piece.
	// Length is generally a power of two unless it gets truncated by the end of the file.
	// Current implementations use 2^14 (16kiB), except close connections, which use more.
	Length uint32
}

func (m MessageRequest) Encode() []byte {
	msg := make([]byte, 12)
	binary.BigEndian.PutUint32(msg[0:4], m.Index)
	binary.BigEndian.PutUint32(msg[4:8], m.Begin)
	binary.BigEndian.PutUint32(msg[8:12], m.Length)
	return createMessageWithPayload(MsgRequest, msg)
}

type MessagePiece struct {
	// The zero-based piece pieceIndex.
	Index uint32

	// The zero-based byte offset within the piece.
	Begin uint32

	// The piece itself.
	Block []byte
}

type MessageKeepAlive struct{}

func (m *Message) Serialize() []byte {
	if m.ID == MsgKeepAlive {
		return make([]byte, 0)
	}

	numBytes := 4 + 1 + len(m.Payload) // requestLength + type + payload
	buffer := make([]byte, numBytes)

	// Total requestLength of message = payload + id
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
	// Read requestLength of message
	lengthBuffer := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuffer)
	if err != nil {
		return nil, err
	}

	// Get requestLength of message
	length := binary.BigEndian.Uint32(lengthBuffer[:])

	// KeepAlive messages have a requestLength of zero
	if length == 0 {
		return buildKeepAliveMessage(), nil
	}

	// Non-keepalive messages start with a single byte (msg type)
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      MessageID(messageBuf[0]),
		Length:  length,
		Payload: messageBuf[1:],
	}, nil
}

func buildKeepAliveMessage() *Message {
	return &Message{
		ID:      MsgKeepAlive,
		Payload: nil,
	}
}

func createMessageWithPayload(id MessageID, payload []byte) []byte {
	msg := make([]byte, len(payload)+1+4) // len + message id + payload
	msgLen := 1 + len(payload)
	binary.BigEndian.PutUint32(msg, uint32(msgLen))
	msg[4] = uint8(id)
	copy(msg[5:], payload)
	return msg
}
