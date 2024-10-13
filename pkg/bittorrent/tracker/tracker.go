// Package tracker provides client implementations for BitTorrent trackers.
// See: https://www.bittorrent.org/beps/bep_0003.html.
package tracker

import (
	"example.com/btclient/pkg/bittorrent/torrentfile"
)

const (
	// Tracker Compact Peer Lists.
	// See: https://www.bittorrent.org/beps/bep_0023.html.
	compactPeerBytesLen = 6
)

type Tracker interface {
	FetchTorrentMetadata(torrent *torrentfile.SimpleTorrentFile) (*Response, error)
}
