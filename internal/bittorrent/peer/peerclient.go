package peer

import (
	"bytes"
	"encoding/binary"
	"errors"
	"example.com/btclient/internal/bittorrent"
	"example.com/btclient/internal/bittorrent/handshake"
	"example.com/btclient/internal/bittorrent/message"
	"example.com/btclient/internal/bittorrent/torrentfile"
	"example.com/btclient/internal/preconditions"
	"fmt"
	"io"
	"math"
	"net"
)

// Client stores the state of a single client connection to a single peer.
type Client struct {
	readConn        net.Conn
	writeConn       net.Conn
	handshaker      *handshake.Handshaker
	extensions      bittorrent.ExtensionBits
	extensionHeader message.ExtensionHeader
	peerID          [20]byte
	infoHash        [20]byte
	InfoDict        *torrentfile.Info
	handshake       *handshake.Handshake
	Bitfield        bittorrent.Bitfield
	isChoked        bool
	isInterested    bool
}

func NewClient(readConn net.Conn, writeConn net.Conn,
	handshaker *handshake.Handshaker,
	extensionBits bittorrent.ExtensionBits,
	peerID [20]byte,
	infoHash [20]byte) *Client {

	return &Client{
		readConn:   readConn,
		writeConn:  writeConn,
		handshaker: handshaker,
		extensions: extensionBits,
		peerID:     peerID,
		infoHash:   infoHash,
		handshake:  nil,
		Bitfield:   nil,
		// connections start out choked and not interested.
		isChoked:     true,
		isInterested: false,
	}
}

func (c *Client) Init() error {
	hs, err := c.doHandshake(c.extensions, c.peerID, c.infoHash)
	if err != nil {
		return err
	}
	println("handshake complete", c.String())

	bf, err := receiveBitfield(c.readConn)
	if err != nil {
		return err
	}
	println("bitfield", c.String(), bf)

	bitfield := bf.Bitfield
	if err := bitfield.Validate(); err != nil {
		return err
	}

	if c.extensions.HasExtensionProtocolBit() {
		if !hs.Extensions.HasExtensionProtocolBit() {
			// Client doesn't support extension protocol
			println("extension protocol selected, but peer client does not support extension protocol")
		} else {
			// exchange supported extensions with peer
			extMsg, err := c.doExtensionHandshake(hs.Extensions)
			if err != nil {
				return err
			}
			c.extensionHeader = extMsg.ExtensionHeader

			// download info dictionary from peer
			numMetadataPieces := int(math.Ceil(float64(extMsg.ExtensionHeader.MetadataSize) / message.UTMetadataBlockSize))
			for i := 0; i < numMetadataPieces; i++ {
				reqMsg := message.NewUTMetadataRequestMsg(i, extMsg.ExtensionHeader)
				reqMsgBytes, err := reqMsg.EncodeUTMetadata()
				if err != nil {
					return err
				}
				_, err = c.writeConn.Write(reqMsgBytes)
				if err != nil {
					return err
				}

				extDataMsg, err := message.Deserialize(c.readConn)
				if err != nil {
					return err
				}
				dataMsg, err := message.ExtendedMessage{}.DecodeUTMetadata(extDataMsg)
				if err != nil {
					return err
				}

				switch dataMsg.UTMetadata.MsgType {
				case message.UTMetadataData:
					// Read the info dict from the datareader message
					infoDictBytes := extDataMsg.Payload[len(extDataMsg.Payload)-dataMsg.UTMetadata.TotalSize:]
					r := bytes.NewReader(infoDictBytes)
					infoDict, err := torrentfile.ReadInfoDict(r)
					if err != nil {
						return err
					}
					c.InfoDict = &infoDict
				case message.UTMetadataReject:
					println("received UTMetadataReject, aborting")
				default:
					return fmt.Errorf("unexpected message type: %v", dataMsg.UTMetadata.MsgType)
				}
			}
		}
	}

	c.handshake = hs
	c.Bitfield = bitfield
	return nil
}

func (c *Client) IsChoked() bool {
	return c.isChoked
}

func (c *Client) SetChoked(isChoked bool) {
	c.isChoked = isChoked
}

func (c *Client) GetBitfield() bittorrent.Bitfield {
	return c.Bitfield
}

func (c *Client) SetBitfield(bf bittorrent.Bitfield) {
	c.Bitfield = bf
}

func (c *Client) ReceiveMessage() (*message.Message, error) {
	return message.Deserialize(c.readConn)
}

func (c *Client) ReceiveUnchokeMessage() (*message.UnchokeMessage, error) {
	_, err := receiveMessageOfType(c.readConn, message.MsgUnchoke)
	return &message.UnchokeMessage{}, err
}

func (c *Client) SendInterestedMessage() error {
	_, err := c.writeConn.Write(message.InterestedMessage{}.Encode())
	return err
}

// SendRequestMessage sends a request to peer to download a section of a piece of datareader.
// pieceIndex: integer specifying the zero-based piece pieceIndex
// begin: integer specifying the zero-based byte offset within the piece
// requestLength: integer specifying the requested requestLength.
func (c *Client) SendRequestMessage(index, begin, length uint32) error {
	b := message.RequestMessage{
		Index:  index,
		Begin:  begin,
		Length: length,
	}.Encode()
	n, err := c.writeConn.Write(b)
	if n == 0 {
		return errors.New("failed to send request")
	}
	return err
}

// TODO we can probably remove this and the other Receive methods.
func (c *Client) ReceivePieceMessage() (*message.PieceMessage, error) {
	msg, err := receiveMessageOfType(c.readConn, message.MsgPiece)
	if err != nil {
		return nil, err
	}

	pieceIndex := binary.BigEndian.Uint32(msg.Payload[0:4])
	pieceBegin := binary.BigEndian.Uint32(msg.Payload[4:8])
	block := msg.Payload[8:]

	return &message.PieceMessage{
		Index: pieceIndex,
		Begin: pieceBegin,
		Block: block,
	}, nil
}

func (c *Client) String() string {
	if c.readConn.RemoteAddr().String() == c.writeConn.RemoteAddr().String() {
		return c.readConn.RemoteAddr().String()
	}
	return fmt.Sprintf("read: %s, write: %s", c.readConn.RemoteAddr().String(), c.writeConn.RemoteAddr().String())
}

func (c *Client) Close() error {
	if err := c.readConn.Close(); err != nil {
		return err
	}
	if err := c.writeConn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Client) readInteger() (uint32, error) {
	buf, err := c.readBytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func (c *Client) readBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := io.ReadFull(c.readConn, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// TODO Move this into handshake.go
func (c *Client) doHandshake(extensionBits bittorrent.ExtensionBits,
	peerID [20]byte,
	infoHash [20]byte,
) (*handshake.Handshake, error) {

	if err := c.handshaker.SendHandshake(extensionBits, peerID, infoHash); err != nil {
		return nil, err
	}

	hs, err := c.handshaker.ReceiveHandshake()
	if err != nil {
		return nil, err
	}

	// If both sides don't send the same info hash value, sever the connection.
	if infoHash != hs.InfoHash {
		return nil, errors.Join(errors.New("different info hash value"), c.Close())
	}

	return hs, nil
}

func (c *Client) doExtensionHandshake(ext bittorrent.ExtensionBits) (*message.ExtendedMessage, error) {
	preconditions.CheckArgument(ext.HasExtensionProtocolBit(), "no extension protocol bit")

	// TODO If the extension protocol is supported, the extension handshake message
	// should be send immediately after the standard BT handshake.
	// It is valid to send the handshake message more than once during the connection's lifetime.
	if err := c.handshaker.SendExtensionHandshake(); err != nil {
		return nil, err
	}

	extMsg, err := c.handshaker.ReceiveExtensionHandshake()
	if err != nil {
		return nil, err
	}

	return extMsg, nil
}

func receiveMessage(conn net.Conn) (*message.Message, error) {
	return message.Deserialize(conn)
}

func receiveMessageOfType(conn net.Conn, id message.Type) (*message.Message, error) {
	msg, err := receiveMessage(conn)
	if err != nil {
		return nil, err
	}

	if msg.ID != id {
		return nil, fmt.Errorf("expected %s, got %s", id, msg.ID.String())
	}

	return msg, nil
}

func receiveBitfield(conn net.Conn) (*message.BitfieldMessage, error) {
	msg, err := receiveMessageOfType(conn, message.MsgBitfield)
	if err != nil {
		return nil, err
	}
	return &message.BitfieldMessage{Bitfield: msg.Payload}, nil
}
