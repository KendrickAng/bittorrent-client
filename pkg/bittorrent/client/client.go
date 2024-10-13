package client

import (
	"context"
	"errors"
	"example.com/btclient/pkg/bittorrent/torrentfile"
	"example.com/btclient/pkg/bittorrent/tracker"
	"fmt"
	"time"
)

const (
	dialTimeout = 30 * time.Second
)

// Client represents a BitTorrent client that downloads a piece of data specified by a Metainfo (.torrent) file.
//
// A Client is higher-level than [DataTransfer] and handles details like torrent file reading and validation.
type Client struct {
	torrent         *torrentfile.SimpleTorrentFile
	tracker         tracker.Tracker
	dataTransferrer DataTransfer
}

type Config struct {
}

// DataTransfer is an interface that represents the ability to download a torrent with a particular schema.
// For example, being able to download over TCP or UDP.
type DataTransfer interface {
	// Download downloads a torrent file, returning a [Response] for the provided [torrentfile.SimpleTorrentFile].
	Download(ctx context.Context, torrent *torrentfile.SimpleTorrentFile) (*Response, error)
}

// TODO refactor this to accept a io.Reader.
func NewClient(torrent torrentfile.SimpleTorrentFile) (*Client, error) {
	if len(torrent.PieceHashes) <= 0 {
		return nil, errors.New("torrent should have pieces to download")
	}
	if torrent.Length <= 0 {
		return nil, errors.New("torrent length should be greater than zero")
	}

	return &Client{torrent: &torrent}, nil
}

func (h *Client) Handle(ctx context.Context) (*Response, error) {
	switch scheme := h.torrent.Announce.Scheme; scheme {
	case "http":
		h.tracker = tracker.DefaultHttpClient
		h.dataTransferrer = &TcpClient{Client: *h}
		return h.dataTransferrer.Download(ctx, h.torrent)
	case "udp":
		panic("udp scheme is not fully supported yet")
	default:
		panic(fmt.Errorf("unsupported scheme: %s", scheme))
	}
}

func (h *Client) Close() {
	// nothing yet
}
