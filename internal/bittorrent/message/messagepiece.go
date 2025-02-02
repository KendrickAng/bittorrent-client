package message

import "encoding/binary"

type PieceMessage struct {
	// The zero-based piece pieceIndex.
	Index uint32

	// The zero-based byte offset within the piece.
	Begin uint32

	// The piece itself.
	Block []byte
}

func (m PieceMessage) Encode() []byte {
	payload := make([]byte, 8+len(m.Block))
	binary.BigEndian.PutUint32(payload[0:4], m.Index)
	binary.BigEndian.PutUint32(payload[4:8], m.Begin)
	copy(payload[8:], m.Block)
	return createMessageWithPayload(MsgPiece, payload)
}

func (m PieceMessage) Decode(msg *Message) *PieceMessage {
	if msg.ID != MsgPiece {
		panic("invalid message piece")
	}
	return &PieceMessage{
		Index: binary.BigEndian.Uint32(msg.Payload[0:4]),
		Begin: binary.BigEndian.Uint32(msg.Payload[4:8]),
		Block: msg.Payload[8:],
	}
}
