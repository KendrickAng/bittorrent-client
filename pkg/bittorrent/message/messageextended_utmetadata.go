package message

import (
	"bytes"
	"example.com/btclient/pkg/preconditions"
	"github.com/jackpal/bencode-go"
)

const (
	UTMetadataRequest UTMetadataType = 0
	UTMetadataData    UTMetadataType = 1
	UTMetadataReject  UTMetadataType = 2

	UTMetadataBlockSize = 16384 // 16 KiB
)

type UTMetadataType int

type UTMetadata struct {
	// An unrecognized message ID must be ignored.
	MsgType UTMetadataType `bencode:"msg_type,omitempty"`

	// Fields from the Request type

	// Indicates which part of the metadata the message refers to.
	// Metadata is handled in blocks of 16KiB (16384 bytes), indexed starting from 0.
	// All blocks are 16KiB except the last block, which may be smaller.
	Piece int `bencode:"piece,omitempty"`

	// Fields from the Data type

	// Total size of the metadata piece in this message.
	// The metadata piece is appended to the bencoded dictionary and is not part of it.
	// However, it is part of the Extension Message and is included in the length prefix.
	// May be less than 16KiB if the piece is the last piece of metadata.
	TotalSize int `bencode:"total_size,omitempty"`

	// Fields from the Reject type
}

func NewUTMetadataRequestMsg(pieceIdx int, headers ExtensionHeader) *ExtendedMessage {
	return &ExtendedMessage{
		// We should always communicate to the client with the message ID they declared during extension handshake.
		ExtendedMessageID: headers.ExtensionMessageID(ENameUTMetadata),
		UTMetadata: UTMetadata{
			MsgType: UTMetadataRequest,
			Piece:   pieceIdx,
		},
	}
}

func (m ExtendedMessage) EncodeUTMetadata() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bencode.Marshal(buf, m.UTMetadata); err != nil {
		return nil, err
	}
	payload := make([]byte, 1+buf.Len())
	payload[0] = m.ExtendedMessageID
	copy(payload[1:], buf.Bytes())
	return createMessageWithPayload(MsgExtended, payload), nil
}

func (m ExtendedMessage) DecodeUTMetadata(msg *Message) (*ExtendedMessage, error) {
	preconditions.CheckArgument(msg.ID == MsgExtended, "invalid message extended")

	extendedMessageID := msg.Payload[0]
	utMetadata := msg.Payload[1:]

	// client should always communicate to us with the message ID we declared to them during extension handshake.
	preconditions.CheckArgument(extendedMessageID == EMessageIDMagnet, "incorrect ut_metadata extension")

	var u UTMetadata
	if err := bencode.Unmarshal(bytes.NewReader(utMetadata), &u); err != nil {
		return nil, err
	}

	return &ExtendedMessage{
		ExtendedMessageID: EMessageIDMagnet,
		UTMetadata:        u,
	}, nil
}
