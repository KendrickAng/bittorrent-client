package tracker

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net/netip"
)

// Response represents information from a http response from a BitTorrent tracker.
type Response struct {
	// Human-readable error message as to why the request failed.
	FailureReason string

	// Interval in seconds that the client should wait between sending regular re-requests to the tracker.
	RefreshInterval int

	// List of (IP, Port), each representing a peer.
	Peers []netip.AddrPort
}

// readResponse reads and returns a BitTorrent tracker response from r.
func readResponse(r *bufio.Reader) (*Response, error) {
	var rawResponse rawTrackerResponse
	if err := bencode.Unmarshal(r, &rawResponse); err != nil {
		return nil, err
	}

	rawPeers := []byte(rawResponse.Peers)
	if len(rawPeers)%compactPeerBytesLen != 0 {
		return nil, fmt.Errorf("malformed peers list, got %d bytes", len(rawPeers))
	}

	n := len(rawPeers) / compactPeerBytesLen
	peers := make([]netip.AddrPort, n)
	for i := 0; i < n; i++ {
		start := i * compactPeerBytesLen
		end := start + compactPeerBytesLen
		addr := ([4]byte)(rawPeers[start : start+4]) // 4 bytes for integer
		port := binary.BigEndian.Uint16(rawPeers[start+4 : end])
		peers[i] = netip.AddrPortFrom(netip.AddrFrom4(addr), port)
	}

	return &Response{
		FailureReason:   rawResponse.FailureReason,
		RefreshInterval: rawResponse.Interval,
		Peers:           peers,
	}, nil
}

// Bencoded response received when connecting to a tracker.
type rawTrackerResponse struct {
	// Human-readable error message as to why the request failed.
	FailureReason string `bencode:"failure reason,omitempty"`

	// Interval in seconds that the client should wait between sending regular re-requests to the tracker.
	Interval int `bencode:"interval"`

	// List of (IP, Port), each representing a peer.
	Peers string `bencode:"peers"`
}
