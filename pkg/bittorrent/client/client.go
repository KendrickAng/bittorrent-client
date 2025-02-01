package client

import (
	"context"
	"errors"
	"example.com/btclient/pkg/bittorrent/peer"
	"example.com/btclient/pkg/bittorrent/torrentfile"
	"example.com/btclient/pkg/bittorrent/tracker"
)

// Client represents a BitTorrent client that downloads a piece of data specified by a Metainfo (.torrent) file.
//
// A Client is higher-level than [DataTransfer] and handles details like torrent file reading and validation.
type Client struct {
	torrent      *torrentfile.SimpleTorrentFile
	tracker      tracker.Tracker
	dataTransfer DataTransfer
}

// DataTransfer is an interface that represents the ability to download a torrent with a particular schema.
// For example, being able to download over TCP or UDP.
type DataTransfer interface {
	// Download downloads a torrent file, returning a [Response] for the provided [torrentfile.SimpleTorrentFile].
	Download(ctx context.Context, torrent *torrentfile.SimpleTorrentFile) (*Response, error)
}

// TODO refactor this to accept a io.Reader.
func NewClient(torrent torrentfile.SimpleTorrentFile, connPool *peer.Pool) (*Client, error) {
	if len(torrent.PieceHashes) <= 0 {
		return nil, errors.New("torrent should have pieces to download")
	}
	if torrent.Length <= 0 {
		return nil, errors.New("torrent length should be greater than zero")
	}

	tcpClient := NewTcpClient(connPool)

	return &Client{torrent: &torrent, dataTransfer: tcpClient}, nil
}

func (h *Client) Handle(ctx context.Context) (*Response, error) {
	return h.dataTransfer.Download(ctx, h.torrent)
}

func (h *Client) Close() {
	// nothing yet
}
