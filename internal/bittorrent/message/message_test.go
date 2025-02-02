package message

import (
	"bytes"
	"fmt"
	"github.com/jackpal/bencode-go"
	"reflect"
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

func TestMessageExtended_Encoode(t *testing.T) {
	// Arrange
	msg := ExtendedMessage{
		ExtendedMessageID: EMessageIDHandshake,
		ExtensionHeader: ExtensionHeader{
			SupportedExtensionMessages: map[string]int{
				"ut_metadata": 3,
			},
			MetadataSize: 31235,
		},
	}

	// Act
	msgExtended, err := msg.EncodeHandshake()
	if err != nil {
		t.Fatal(err)
	}

	// Assert
	buf := new(bytes.Buffer)
	if err := bencode.Marshal(buf, msg.ExtensionHeader); err != nil {
		t.Fatal(err)
	}
	expectedBytes := append([]byte{
		0, 0, 0, byte(buf.Len() + 2),
		uint8(MsgExtended),
		uint8(EMessageIDHandshake),
	},
		buf.Bytes()...)
	if !bytes.Equal(msgExtended, expectedBytes) {
		t.Fatal("incorrect bytes, got", msgExtended)
	}
}

func TestExtendedMessage_Decode(t *testing.T) {
	// taken from wireshark
	msg, err := Deserialize(bytes.NewReader([]byte{0x0, 0x0, 0x0, 0x63, 0x14, 0x0, 0x64, 0x31, 0x3a, 0x6d, 0x64, 0x31, 0x31, 0x3a, 0x75, 0x74, 0x5f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x69, 0x31, 0x65, 0x36, 0x3a, 0x75, 0x74, 0x5f, 0x70, 0x65, 0x78, 0x69, 0x32, 0x65, 0x65, 0x31, 0x33, 0x3a, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x73, 0x69, 0x7a, 0x65, 0x69, 0x31, 0x33, 0x32, 0x65, 0x34, 0x3a, 0x72, 0x65, 0x71, 0x71, 0x69, 0x32, 0x35, 0x30, 0x65, 0x31, 0x3a, 0x76, 0x31, 0x30, 0x3a, 0x52, 0x61, 0x69, 0x6e, 0x20, 0x30, 0x2e, 0x30, 0x2e, 0x30, 0x36, 0x3a, 0x79, 0x6f, 0x75, 0x72, 0x69, 0x70, 0x34, 0x3a, 0x9a, 0xcd, 0x10, 0x25, 0x65}))
	if err != nil {
		t.Fatal(err)
	}

	decodedMsg, err := ExtendedMessage{}.DecodeHandshake(msg)
	if err != nil {
		t.Fatal(err)
	}

	if decodedMsg.ExtendedMessageID != EMessageIDHandshake {
		t.Fatal("incorrect ExtendedMessageID, got", decodedMsg.ExtendedMessageID)
	}
	if !reflect.DeepEqual(decodedMsg, &ExtendedMessage{
		ExtendedMessageID: EMessageIDHandshake,
		ExtensionHeader: ExtensionHeader{
			SupportedExtensionMessages: map[string]int{
				"ut_metadata": 1,
				"ut_pex":      2,
			},
			MetadataSize: 132,
			Reqq:         250,
			Version:      "Rain 0.0.0",
			YourIP:       string([]byte{0x9A, 0xCD, 0x10, 0x25}),
		},
	}) {
		t.Fatal(fmt.Sprintf("incorrect message, got %+v", decodedMsg))
	}
}
