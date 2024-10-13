package message

import (
	"bytes"
	"testing"
)

func TestMessageUnchoke_Encode(t *testing.T) {
	// Arrange
	msg := UnchokeMessage{}

	// Act
	msgBytes := msg.Encode()

	// Assert
	// - Message length : 1
	// - Message Type: Unchoke (1)
	// Message taken from wireshark dump.
	if !bytes.Equal(msgBytes, []byte{0, 0, 0, 1, uint8(MsgUnchoke)}) {
		t.Fatal("incorrect bytes, got", msgBytes)
	}
}

func TestMessageInterested_Encode(t *testing.T) {
	// Arrange
	msg := InterestedMessage{}

	// Act
	msgBytes := msg.Encode()

	// Assert
	if !bytes.Equal(msgBytes, []byte{0, 0, 0, 1, uint8(MsgInterested)}) {
		t.Fatal("incorrect bytes, got", msgBytes)
	}
}

func TestMessageRequest_Encode(t *testing.T) {
	// Arrange
	msg := RequestMessage{
		Index:  1,
		Begin:  2,
		Length: 3,
	}

	// Act
	msgBytes := msg.Encode()

	// Assert
	expectedBytes := []byte{
		0, 0, 0, 13,
		uint8(MsgRequest),
		0, 0, 0, 1,
		0, 0, 0, 2,
		0, 0, 0, 3}
	if !bytes.Equal(msgBytes, expectedBytes) {
		t.Fatal("incorrect bytes, got", msgBytes)
	}
}

func TestMessagePiece_Encode(t *testing.T) {
	// Arrange
	msg := PieceMessage{
		Index: 1,
		Begin: 2,
		Block: []byte{3, 4, 5, 6, 7, 8},
	}

	// Act
	msgBytes := msg.Encode()

	// Assert
	expectedBytes := append([]byte{
		0, 0, 0, 15,
		uint8(MsgPiece),
		0, 0, 0, 1,
		0, 0, 0, 2,
	},
		msg.Block...)
	if !bytes.Equal(msgBytes, expectedBytes) {
		t.Fatal("incorrect bytes, got", msgBytes)
	}
}
