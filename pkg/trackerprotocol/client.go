package trackerprotocol

import (
	"example.com/btclient/pkg/bencodeutil"
	"fmt"
	"net"
	"time"
)

const (
	peerDialTimeout = 10 * time.Second
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
		bitfield:  bf,
		// connections start out choked and not interested.
		isChoked:     true,
		isInterested: false,
	}, nil
}

func (c *Client) Start() error {

	// TODO f

	return nil
}
func (c *Client) Close() error {
	return c.conn.Close()
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

	return &handshake, nil
}

func receiveMessage(conn net.Conn) (*Message, error) {
	msg, err := Deserialize(conn)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func receiveBitfield(conn net.Conn) (Bitfield, error) {
	msg, err := receiveMessage(conn)
	if err != nil {
		return nil, err
	}

	if msg.ID != MsgBitfield {
		return nil, fmt.Errorf("expected bitfield, got %s", msg.ID.String())
	}

	return Bitfield(msg.Payload), nil
}
