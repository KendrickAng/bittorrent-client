// Package tracker provides client implementations for BitTorrent trackers.
// See: https://www.bittorrent.org/beps/bep_0003.html.
package tracker

import (
	"net/url"
)

const (
	// Tracker Compact Peer Lists.
	// See: https://www.bittorrent.org/beps/bep_0023.html.
	compactPeerBytesLen = 6
)

type Tracker interface {
	FetchTorrentMetadata(request FetchTorrentMetadataRequest) (*Response, error)
}

type FetchTorrentMetadataRequest struct {
	TrackerUrl *url.URL
	InfoHash   [20]byte
	PeerID     [20]byte
	Left       int
}
