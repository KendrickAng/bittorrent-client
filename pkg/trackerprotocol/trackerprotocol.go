package trackerprotocol

import (
	"errors"
	"example.com/btclient/pkg/bencodeutil"
	"example.com/btclient/pkg/preconditions"
	"example.com/btclient/pkg/udpprotocol"
	"fmt"
	"net"
	"net/url"
	"time"
)

const (
	dialTimeout = 30 * time.Second
)

type Handler struct {
	announceUrl *url.URL
	torrent     *bencodeutil.SimpleTorrentFile

	// HTTP
	httpListener net.Listener
}

func NewHandler(torrent bencodeutil.SimpleTorrentFile) (*Handler, error) {
	preconditions.CheckArgument(len(torrent.Announce) > 0)

	// Parse announce announceUrl
	announceUrl, err := url.Parse(torrent.Announce)
	if err != nil {
		return nil, err
	}

	return &Handler{announceUrl: announceUrl, torrent: &torrent}, nil
}

func (h *Handler) Handle() error {
	switch scheme := h.announceUrl.Scheme; scheme {

	case "udp":
		// Dial connection to announce url
		conn, err := net.DialTimeout(
			h.announceUrl.Scheme,
			fmt.Sprintf("%s:%s", h.announceUrl.Hostname(), h.announceUrl.Port()),
			dialTimeout,
		)
		if err != nil {
			return err
		}

		// Initialize UDP client
		if err := udpprotocol.Connect(&conn); err != nil {
			return err
		}
		return errors.New("udp scheme is not fully supported yet")

	case "http":
		return h.handleHttp()

	default:
		return fmt.Errorf("unsupported scheme: %s", scheme)

	}
}

func (h *Handler) Close() {
	// nothing yet
}
