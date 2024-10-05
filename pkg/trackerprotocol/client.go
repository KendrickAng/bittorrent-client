package trackerprotocol

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const (
	peerDialTimeout = 30 * time.Second
)

// Client stores the state of a single client connection to a single peer.
type Client struct {
	readConn     net.Conn
	writeConn    net.Conn
	handshaker   *Handshaker
	peerID       [20]byte
	infoHash     [20]byte
	handshake    *Handshake
	bitfield     Bitfield
	isChoked     bool
	isInterested bool
}

func NewClient(readConn net.Conn, writeConn net.Conn,
	handshaker *Handshaker,
	peerID [20]byte,
	infoHash [20]byte) *Client {

	return &Client{
		readConn:   readConn,
		writeConn:  writeConn,
		handshaker: handshaker,
		peerID:     peerID,
		infoHash:   infoHash,
		handshake:  nil,
		bitfield:   nil,
		// connections start out choked and not interested.
		isChoked:     true,
		isInterested: false,
	}
}

func (c *Client) Init() error {
	handshake, err := c.doHandshake(c.peerID, c.infoHash)
	if err != nil {
		return err
	}
	println("handshake complete", c.String())

	bf, err := receiveBitfield(c.readConn)
	if err != nil {
		return err
	}
	println("bitfield received", c.String(), bf)

	c.handshake = handshake
	c.bitfield = bf.Bitfield
	return nil
}

func (c *Client) IsChoked() bool {
	return c.isChoked
}

func (c *Client) SetChoked(isChoked bool) {
	c.isChoked = isChoked
}

func (c *Client) GetBitfield() Bitfield {
	return c.bitfield
}

func (c *Client) SetBitfield(bf Bitfield) {
	c.bitfield = bf
}

func (c *Client) ReceiveMessage() (*Message, error) {
	return Deserialize(c.readConn)
}

func (c *Client) ReceiveUnchokeMessage() (*MessageUnchoke, error) {
	_, err := receiveMessageOfType(c.readConn, MsgUnchoke)
	return &MessageUnchoke{}, err
}

func (c *Client) SendInterestedMessage() error {
	_, err := c.writeConn.Write(MessageInterested{}.Encode())
	return err
}

// SendRequestMessage sends a request to peer to download a section of a piece of data.
// pieceIndex: integer specifying the zero-based piece pieceIndex
// begin: integer specifying the zero-based byte offset within the piece
// requestLength: integer specifying the requested requestLength.
func (c *Client) SendRequestMessage(index, begin, length uint32) error {
	b := MessageRequest{
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

func (c *Client) ReceivePieceMessage() (*MessagePiece, error) {
	msg, err := receiveMessageOfType(c.readConn, MsgPiece)
	if err != nil {
		return nil, err
	}

	pieceIndex := binary.BigEndian.Uint32(msg.Payload[0:4])
	pieceBegin := binary.BigEndian.Uint32(msg.Payload[4:8])
	block := msg.Payload[8:]

	return &MessagePiece{
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

func (c *Client) doHandshake(peerID [20]byte, infoHash [20]byte) (*Handshake, error) {
	if err := c.handshaker.SendHandshake(peerID, infoHash); err != nil {
		return nil, err
	}

	handshake, err := c.handshaker.ReceiveHandshake()
	if err != nil {
		return nil, err
	}

	// If both sides don't send the same info hash value, sever the connection.
	if infoHash != handshake.InfoHash {
		return nil, errors.Join(errors.New("different info hash value"), c.Close())
	}

	return handshake, nil
}

func receiveMessage(conn net.Conn) (*Message, error) {
	return Deserialize(conn)
}

func receiveMessageOfType(conn net.Conn, id MessageID) (*Message, error) {
	msg, err := receiveMessage(conn)
	if err != nil {
		return nil, err
	}

	if msg.ID != id {
		return nil, fmt.Errorf("expected %s, got %s", id, msg.ID.String())
	}

	return msg, nil
}

func receiveBitfield(conn net.Conn) (*MessageBitfield, error) {
	msg, err := receiveMessageOfType(conn, MsgBitfield)
	if err != nil {
		return nil, err
	}
	return &MessageBitfield{Bitfield: msg.Payload}, nil
}
