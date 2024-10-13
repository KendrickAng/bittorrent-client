package tracker

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net"
)

type Response struct {
	// Human-readable error message as to why the request failed.
	FailureReason string

	// Interval in seconds that the client should wait between sending regular re-requests to the tracker.
	RefreshInterval int

	// List of (IP, Port), each representing a peer.
	Peers []Peer
}

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p *Peer) String() string {
	return fmt.Sprintf("%s:%d", p.IP, p.Port)
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

// ReadResponse reads and returns a BitTorrent tracker response from r.
func ReadResponse(r *bufio.Reader) (*Response, error) {
	var rawResponse rawTrackerResponse
	if err := bencode.Unmarshal(r, &rawResponse); err != nil {
		return nil, err
	}

	rawPeers := []byte(rawResponse.Peers)
	if len(rawPeers)%compactPeerBytesLen != 0 {
		return nil, fmt.Errorf("malformed peers list, got %d bytes", len(rawPeers))
	}

	n := len(rawPeers) / compactPeerBytesLen
	peers := make([]Peer, n)
	for i := 0; i < n; i++ {
		start := i * compactPeerBytesLen
		end := start + compactPeerBytesLen
		peers[i].IP = rawPeers[start : start+4] // 4 bytes for integer
		peers[i].Port = binary.BigEndian.Uint16(rawPeers[start+4 : end])
	}

	return &Response{
		FailureReason:   rawResponse.FailureReason,
		RefreshInterval: rawResponse.Interval,
		Peers:           peers,
	}, nil
}
