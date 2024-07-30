package trackerprotocol

import (
	"errors"
	"example.com/btclient/pkg/bencodeutil"
	"example.com/btclient/pkg/closelogger"
	"fmt"
	"golang.org/x/exp/rand"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	startPort = 6881
	endPort   = 6889

	peerDialTimeout        = 10 * time.Second
	btProtocolID    string = "BitTorrent protocol"
)

func (h *Handler) handleHttp() error {
	// Reserve port for this application
	port, err := h.reservePort()
	if err != nil {
		return err
	}

	// Generate a random peer ID
	peerID, err := random20Bytes()
	if err != nil {
		return err
	}

	// Build GET request to tracker
	trackerURL, err := h.buildTrackerURL(peerID, port)
	if err != nil {
		return err
	}

	// Send GET request to tracker
	resp, err := http.Get(trackerURL)
	if err != nil {
		return err
	}
	defer closelogger.CloseOrLog(resp.Body, "Tracker GET response body")

	// Retrieve tracker response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Parse tracker response
	trackerResp, err := bencodeutil.UnmarshalTrackerResponse(body)
	if err != nil {
		return err
	}
	if len(trackerResp.Peers) == 0 {
		return errors.New("no peers found")
	}
	fmt.Printf("Received tracker response: %+v\n", trackerResp)

	// TODO handle tracker refresh interval

	// TODO peer concurrency

	for _, peer := range trackerResp.Peers {
		go func() {
			// Dial peer
			fmt.Println("Connecting to peer " + peer.String())

			conn, err := net.DialTimeout("tcp", peer.String(), peerDialTimeout)
			if err != nil {
				return
			}
			defer conn.Close()

			fmt.Printf("Establish TCP client to peer: %+v\n", peer)

			// Initiate handshake with peer
			handshake := buildHandshake(btProtocolID, peerID, h.torrent.InfoHash)
			_, err = conn.Write(handshake)
			if err != nil {
				return
			}

			// Receive message from peer
			buf := make([]byte, 1024)
			_, err = conn.Read(buf)
			if err != nil {
				return
			}
			fmt.Println("RECEIVED: " + string(buf))
		}()
	}
	//peer := trackerResp.Peers[0]

	return nil
}

func (h *Handler) reservePort() (int, error) {
	for port := startPort; port <= endPort; port++ {
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			h.httpListener = listener
			return port, nil
		}
	}
	return -1, errors.New("could not find free port")
}

func (h *Handler) buildTrackerURL(peerID [20]byte, port int) (string, error) {
	base, err := h.announceUrl.Parse(h.torrent.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(h.torrent.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(h.torrent.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

type Handshake struct {
	protocolID       string
	bencodedInfoHash [20]byte
	peerID           [20]byte
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

func random20Bytes() ([20]byte, error) {
	var bb [20]byte

	b, err := randomBytes(20)
	if err != nil {
		return bb, err
	}

	copy(bb[:], b)
	return bb, nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
