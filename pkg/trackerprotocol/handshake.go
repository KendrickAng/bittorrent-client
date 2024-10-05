package trackerprotocol

import (
	"fmt"
	"io"
	"net"
)

const (
	btProtocolID string = "BitTorrent protocol"
)

type Handshaker struct {
	peerConn net.Conn
}

type Handshake struct {
	ProtocolID string
	PeerID     [20]byte
	InfoHash   [20]byte
}

func NewHandshaker(conn net.Conn) *Handshaker {
	return &Handshaker{
		peerConn: conn,
	}
}

func (h *Handshaker) SendHandshake(peerID [20]byte, infoHash [20]byte) error {
	peerID = [20]byte{0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9}
	handshake := buildHandshake(btProtocolID, peerID, infoHash)
	_, err := h.peerConn.Write(handshake)
	return err
}

func (h *Handshaker) ReceiveHandshake() (*Handshake, error) {
	// Read requestLength of protocol ID
	lengthBuffer := make([]byte, 1)
	_, err := io.ReadFull(h.peerConn, lengthBuffer)
	if err != nil {
		return nil, err
	}

	// Read protocol ID
	protocolIDBuffer := make([]byte, int(lengthBuffer[0]))
	_, err = io.ReadFull(h.peerConn, protocolIDBuffer)
	if err != nil {
		return nil, err
	}
	if string(protocolIDBuffer) != btProtocolID {
		return nil, fmt.Errorf("expected peer protocol ID '%s', got '%s'", btProtocolID, string(protocolIDBuffer))
	}

	// Read extension bits, info hash, peer ID
	buf := make([]byte, 8+20+20)
	_, err = io.ReadFull(h.peerConn, buf)
	if err != nil {
		return nil, err
	}

	var extensionBits [20]byte
	var infoHash [20]byte
	var peerID [20]byte
	copy(extensionBits[:], buf[:8])
	copy(infoHash[:], buf[8:28])
	copy(peerID[:], buf[28:48])

	return &Handshake{
		ProtocolID: string(protocolIDBuffer),
		PeerID:     peerID,
		InfoHash:   infoHash,
	}, nil
}

func buildHandshake(protocolID string, peerID [20]byte, bencodedInfoHash [20]byte) []byte {
	buf := make([]byte, len(protocolID)+49)
	buf[0] = byte(len(protocolID))
	ptr := 1 // first byte taken by '19'
	ptr += copy(buf[ptr:], protocolID)
	ptr += copy(buf[ptr:], make([]byte, 8))
	ptr += copy(buf[ptr:], bencodedInfoHash[:])
	ptr += copy(buf[ptr:], peerID[:])
	return buf
}
