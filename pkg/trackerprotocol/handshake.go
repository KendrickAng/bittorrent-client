package trackerprotocol

import (
	"errors"
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
	handshake := buildHandshake(btProtocolID, peerID, infoHash)
	_, err := h.peerConn.Write(handshake)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handshaker) ReceiveHandshake() (Handshake, error) {
	// TODO receive handshake
	return Handshake{}, errors.New("Not supported")
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
