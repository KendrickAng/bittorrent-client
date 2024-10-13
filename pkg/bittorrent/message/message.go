package message

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Message contains the ID and payload of a message
type Message struct {
	ID      Type
	Payload []byte
	// Total requestLength of the message in bytes.
	Length uint32
}

func (m *Message) AsMsgBitfield() *BitfieldMessage {
	return BitfieldMessage{}.Decode(m)
}

func (m *Message) AsMsgPiece() *PieceMessage {
	return PieceMessage{}.Decode(m)
}

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
	n, err := r.Read(lengthBuffer)
	if n == 0 {
		return nil, errors.New("unexpected 0 length read")
	}
	if err != nil {
		return nil, err
	}

	// Get requestLength of message
	length := binary.BigEndian.Uint32(lengthBuffer[:])

	// KeepAlive messages have a requestLength of zero
	if length == 0 {
		return &DefaultKeepAliveMessage, nil
	}

	// Non-keepalive messages start with a single byte (msg type)
	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      Type(messageBuf[0]),
		Length:  length,
		Payload: messageBuf[1:],
	}, nil
}

func createMessageWithPayload(id Type, payload []byte) []byte {
	msg := make([]byte, len(payload)+1+4) // len + message id + payload
	msgLen := 1 + len(payload)
	binary.BigEndian.PutUint32(msg, uint32(msgLen))
	msg[4] = uint8(id)
	copy(msg[5:], payload)
	return msg
}
