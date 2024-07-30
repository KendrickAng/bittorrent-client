package bencodeutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jackpal/bencode-go"
	"net"
)

const (
	compactPeerBytesLen = 6
)

type TrackerResponse struct {
	FailureReason   string
	RefreshInterval int
	Peers           []Peer
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
	FailureReason string `bencode:"failure reason,omitempty"`
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"`
}

func UnmarshalTrackerResponse(b []byte) (*TrackerResponse, error) {
	var rawResponse rawTrackerResponse
	if err := bencode.Unmarshal(bytes.NewReader(b), &rawResponse); err != nil {
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
		peers[i].IP = net.IP(rawPeers[start : start+4])
		peers[i].Port = binary.BigEndian.Uint16(rawPeers[start+4 : end])
	}

	return &TrackerResponse{
		FailureReason:   rawResponse.FailureReason,
		RefreshInterval: rawResponse.Interval,
		Peers:           peers,
	}, nil
}
