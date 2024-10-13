package message

import "encoding/binary"

type RequestMessage struct {
	// The zero-based piece pieceIndex.
	Index uint32

	// The zero-based byte offset within the piece.
	Begin uint32

	// The requested requestLength of the piece.
	// Length is generally a power of two unless it gets truncated by the end of the file.
	// Current implementations use 2^14 (16kiB), except close connections, which use more.
	Length uint32
}

func (m RequestMessage) Encode() []byte {
	msg := make([]byte, 12)
	binary.BigEndian.PutUint32(msg[0:4], m.Index)
	binary.BigEndian.PutUint32(msg[4:8], m.Begin)
	binary.BigEndian.PutUint32(msg[8:12], m.Length)
	return createMessageWithPayload(MsgRequest, msg)
}
