package trackerprotocol

import (
	"encoding/binary"
	"errors"
	"example.com/btclient/pkg/bencodeutil"
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
	conn         net.Conn
	peerID       [20]byte
	infoHash     [20]byte
	handshake    *Handshake
	bitfield     Bitfield
	isChoked     bool
	isInterested bool
}

func NewClient(peer bencodeutil.Peer, peerID [20]byte, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), peerDialTimeout)
	if err != nil {
		return nil, err
	}
	if err = conn.SetDeadline(time.Time{}); err != nil { // no I/O deadline
		return nil, err
	}

	handshake, err := doHandshake(conn, peerID, infoHash)
	if err != nil {
		return nil, err
	}

	bf, err := receiveBitfield(conn)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:      conn,
		peerID:    peerID,
		infoHash:  infoHash,
		handshake: handshake,
		bitfield:  bf.Bitfield,
		// connections start out choked and not interested.
		isChoked:     true,
		isInterested: false,
	}, nil
}

func (c *Client) ReceiveUnchokeMessage() (*MessageUnchoke, error) {
	_, err := receiveMessageOfType(c.conn, MsgUnchoke)
	return &MessageUnchoke{}, err
}

func (c *Client) SendInterestedMessage() error {
	_, err := c.conn.Write(MessageInterested{}.Encode())
	return err
}

func (c *Client) SendRequestMessage(index, begin, length uint32) error {
	_, err := c.conn.Write(MessageRequest{
		Index:  index,
		Begin:  begin,
		Length: length,
	}.Encode())
	return err
}

func (c *Client) ReceivePieceMessage() (*MessagePiece, error) {
	msg, err := receiveMessageOfType(c.conn, MsgPiece)
	if err != nil {
		return nil, err
	}

	pieceIndex, err := c.readInteger()
	if err != nil {
		return nil, err
	}

	pieceBegin, err := c.readInteger()
	if err != nil {
		return nil, err
	}

	blockLength := msg.Length - 9 // message id + index + begin = 9
	block, err := c.readBytes(int(blockLength))
	if err != nil {
		return nil, err
	}

	return &MessagePiece{
		Index: pieceIndex,
		Begin: pieceBegin,
		Block: block,
	}, nil
}

func (c *Client) String() string {
	return c.conn.RemoteAddr().String()
}

func (c *Client) Close() error {
	return c.conn.Close()
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
	_, err := io.ReadFull(c.conn, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func doHandshake(conn net.Conn, peerID [20]byte, infoHash [20]byte) (*Handshake, error) {
	handshaker := NewHandshaker(conn)
	if err := handshaker.SendHandshake(peerID, infoHash); err != nil {
		return nil, err
	}

	handshake, err := handshaker.ReceiveHandshake()
	if err != nil {
		return nil, err
	}

	// If both sides don't send the same info hash value, sever the connection.
	if infoHash != handshake.InfoHash {
		return nil, errors.Join(errors.New("different info hash value"), conn.Close())
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
