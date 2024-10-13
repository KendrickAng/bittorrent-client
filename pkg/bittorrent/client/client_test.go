package client

import (
	"bytes"
	"example.com/btclient/pkg/bittorrent/handshake"
	"example.com/btclient/pkg/bittorrent/message"
	"io"
	"net"
	"testing"
	"time"
)

func TestClient_IsChoked(t *testing.T) {
	// Arrange
	var a1 [20]byte
	var a2 [20]byte
	reader, writer := net.Pipe()
	client := NewClient(reader, writer, handshake.NewHandshaker(writer), a1, a2)
	defer client.Close()

	// Assert
	if client.IsChoked() != true {
		t.Fatal("client is not choked")
	}
}

func TestClient_ReceiveUnchokeMessage(t *testing.T) {
	// Arrange
	var a1 [20]byte
	var a2 [20]byte
	reader, writer := net.Pipe()
	writer.SetWriteDeadline(time.Now().Add(time.Minute * 3))
	reader.SetReadDeadline(time.Now().Add(time.Minute * 3))
	client := NewClient(reader, writer, handshake.NewHandshaker(writer), a1, a2)
	defer client.Close()

	go func() {
		if _, err := writer.Write(unchokeMessage()); err != nil {
			t.Error(err)
			return
		}

		writer.Close()
	}()

	// Act
	actualMsgUnchoke, err := client.ReceiveUnchokeMessage()
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	if actualMsgUnchoke.Encode()[0] != 5 {
		t.Fatal("unchoke message not equal")
	}
	if actualMsgUnchoke.Encode()[4] != uint8(message.MsgUnchoke) {
		t.Fatal("unchoke message not equal")
	}
}

func TestClient_Init(t *testing.T) {
	// TODO test handshake and Bitfield sending
}

func TestClient_SendInterestedMessage(t *testing.T) {
	// Arrange
	var a1 [20]byte
	var a2 [20]byte
	reader, writer := net.Pipe()
	client := NewClient(reader, writer, handshake.NewHandshaker(writer), a1, a2)
	defer client.Close()

	// Act
	go func() {
		if err := client.SendInterestedMessage(); err != nil {
			t.Error(err)
		}
		writer.Close()
	}()

	// Assert
	read, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(read, []byte{0, 0, 0, 1, uint8(message.MsgInterested)}) {
		t.Fatal("incorrect bytes")
	}
}

func TestClient_SendRequestMessage(t *testing.T) {
	// Arrange
	var a1 [20]byte
	var a2 [20]byte
	reader, writer := net.Pipe()
	client := NewClient(reader, writer, handshake.NewHandshaker(writer), a1, a2)
	defer client.Close()

	// Act
	go func() {
		if err := client.SendRequestMessage(0, 1, 2); err != nil {
			t.Error(err)
		}
		writer.Close()
	}()

	// Assert
	read, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(read, []byte{
		0, 0, 0, 13,
		uint8(message.MsgRequest),
		0, 0, 0, 0,
		0, 0, 0, 1,
		0, 0, 0, 2}) {
		t.Fatal("incorrect bytes")
	}
}

func TestClient_ReceivePieceMessage(t *testing.T) {
	// Arrange
	var a1 [20]byte
	var a2 [20]byte
	reader, writer := net.Pipe()
	client := NewClient(reader, writer, handshake.NewHandshaker(writer), a1, a2)
	defer client.Close()

	go func() {
		if _, err := writer.Write(message.PieceMessage{
			Index: 1,
			Begin: 2,
			Block: []byte{3, 4, 5, 6, 7, 8, 9},
		}.Encode()); err != nil {
			t.Error(err)
		}

		writer.Close()
	}()

	// Act
	pieceMsg, err := client.ReceivePieceMessage()
	if err != nil {
		t.Fatal(err)
	}
	if pieceMsg.Index != 1 {
		t.Fatal("incorrect index")
	}
	if pieceMsg.Begin != 2 {
		t.Fatal("incorrect begin")
	}
	if !bytes.Equal(pieceMsg.Block, []byte{3, 4, 5, 6, 7, 8, 9}) {
		t.Fatal("incorrect bytes")
	}
}

func unchokeMessage() []byte {
	return message.UnchokeMessage{}.Encode()
}
