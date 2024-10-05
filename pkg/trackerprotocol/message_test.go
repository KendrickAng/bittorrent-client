package trackerprotocol

import (
	"bytes"
	"testing"
)

func TestMessageUnchoke_Encode(t *testing.T) {
	// Arrange
	msg := MessageUnchoke{}

	// Act
	msgBytes := msg.Encode()

	// Assert
	// - Message length : 1
	// - Message Type: Unchoke (1)
	// Message taken from wireshark dump.
	if len(msgBytes) != 5 {
		t.Fatal("incorrect length")
	}
	if !bytes.Equal(msgBytes, []byte{0, 0, 0, 1, uint8(MsgUnchoke)}) {
		t.Fatal("incorrect bytes, got", msgBytes)
	}
}

func TestMessageInterested_Encode(t *testing.T) {
	// Arrange
	msg := MessageInterested{}

	// Act
	msgBytes := msg.Encode()

	// Assert
	if len(msgBytes) != 5 {
		t.Fatal("incorrect length")
	}
	if !bytes.Equal(msgBytes, []byte{0, 0, 0, 1, uint8(MsgInterested)}) {
		t.Fatal("incorrect bytes, got", msgBytes)
	}
}

func TestMessageRequest_Encode(t *testing.T) {
	// TODO
}
