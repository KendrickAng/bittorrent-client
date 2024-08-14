package trackerprotocol

import (
	"context"
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
)

const (
	startPort = 6881
	endPort   = 6889
)

func (h *Handler) handleHttp(ctx context.Context) error {
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

	// Connect to available peers
	var clients []*Client
	for _, peer := range trackerResp.Peers {
		client, err := NewClient(peer, peerID, h.torrent.InfoHash)
		if err != nil {
			fmt.Printf("Error creating client for peer %s: %s\n", peer.String(), err)
			return err
		}
		clients = append(clients, client)

		fmt.Printf("Created client for %s\n", peer.String())
	}

	// Start download manager
	manager, err := NewDownloadManager(h.torrent, clients)
	if err != nil {
		return err
	}
	if err := manager.Start(ctx); err != nil {
		return err
	}

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
