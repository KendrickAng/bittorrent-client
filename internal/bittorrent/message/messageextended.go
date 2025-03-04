package message

import (
	"bytes"
	"example.com/btclient/internal/preconditions"
	"fmt"
	"github.com/jackpal/bencode-go"
)

const (
	EMessageIDHandshake uint8 = 0
	EMessageIDMagnet    uint8 = 20 // arbitrary

	ENameUTMetadata string = "ut_metadata"
)

// See: https://www.bittorrent.org/beps/bep_0010.html.
type ExtendedMessage struct {
	ExtendedMessageID uint8
	ExtensionHeader   ExtensionHeader
	UTMetadata        UTMetadata
}

type ExtensionHeader struct {
	// Maps names of extensions to an extended message ID for each extension message.
	// The extension message IDs are the IDs used to send the extension messages to
	// the peer sending this handshake, i.e. IDs are local to peers.
	SupportedExtensionMessages map[string]int `bencode:"m,omitempty"`

	// Handshake message fields (included in BEP 10)

	// OPTIONAL. Allows each side to learn about the TCP port of the other side.
	LocalTcpListenPort uint16 `bencode:"p,omitempty"`
	// OPTIONAL. Client name and version (UTF-8).
	// More reliable for identifying the client than using peer ID encoding.
	Version string `bencode:"v,omitempty"`
	// OPTIONAL. Compact representation of the IP address the peer sees you as.
	// I.e. the receiver's external IP address, without port (IPV4/IPV6)
	YourIP string `bencode:"yourip,omitempty"`
	// OPTIONAL. If the peer has an IPV6 interface, this is the compact representation of it.
	// Clients may prefer to connect back via this interface (4 bytes).
	IPV4 string `bencode:"ipv4,omitempty"`
	// OPTIONAL. If the peer has an IPV4 interface, this is the compact representation of it.
	// Clients may prefer to connect back via this interface (16 bytes).
	IPV6 string `bencode:"ipv6,omitempty"`
	// OPTIONAL. Number of outstanding request messages this client supports without dropping any.
	Reqq int `bencode:"reqq,omitempty"`

	// Other extension-specific fields (not included in BEP 10)

	// Number of bytes of the info-dictionary part of the .torrent file.
	MetadataSize int `bencode:"metadata_size,omitempty"`
}

func (e ExtensionHeader) ExtensionMessageID(extensionName string) uint8 {
	val, ok := e.SupportedExtensionMessages[extensionName]
	preconditions.CheckArgument(ok, fmt.Sprintf("extension header does not have %s", extensionName))
	return uint8(val)
}

func NewExtensionHandshakeMsg() ExtendedMessage {
	return ExtendedMessage{
		ExtendedMessageID: EMessageIDHandshake,
		ExtensionHeader: ExtensionHeader{
			SupportedExtensionMessages: map[string]int{
				ENameUTMetadata: int(EMessageIDMagnet),
			},
		},
	}
}

func (m ExtendedMessage) EncodeHandshake() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := bencode.Marshal(buf, m.ExtensionHeader); err != nil {
		return nil, err
	}
	payload := make([]byte, 1+buf.Len())
	payload[0] = m.ExtendedMessageID
	copy(payload[1:], buf.Bytes())
	return createMessageWithPayload(MsgExtended, payload), nil
}

func (m ExtendedMessage) DecodeHandshake(msg *Message) (*ExtendedMessage, error) {
	preconditions.CheckArgument(msg.ID == MsgExtended, "invalid message extended")

	extendedMessageID := msg.Payload[0]
	preconditions.CheckArgument(extendedMessageID == EMessageIDHandshake, "invalid ExtendedMessageID")
	extensionHeaders := msg.Payload[1:]

	var e ExtensionHeader
	if err := bencode.Unmarshal(bytes.NewReader(extensionHeaders), &e); err != nil {
		return nil, err
	}

	return &ExtendedMessage{
		ExtendedMessageID: extendedMessageID,
		ExtensionHeader:   e,
	}, nil
}
