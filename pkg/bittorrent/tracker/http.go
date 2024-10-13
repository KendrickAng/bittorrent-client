package tracker

import (
	"bufio"
	"bytes"
	"errors"
	"example.com/btclient/pkg/bittorrent/torrentfile"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	minTrackerPort = 6881
	maxTrackerPort = 6889
)

var (
	DefaultHttpClient = &HttpClient{}
)

type HttpClient struct{}

func (h *HttpClient) FetchTorrentMetadata(torrent *torrentfile.SimpleTorrentFile) (*Response, error) {
	trackerResponse, err := fetchTorrentMetadataFromTracker(torrent)
	if err != nil {
		return nil, err
	}
	return readResponse(bufio.NewReader(bytes.NewReader(trackerResponse)))
}

func fetchTorrentMetadataFromTracker(torrent *torrentfile.SimpleTorrentFile) (data []byte, error error) {
	for port := minTrackerPort; port <= maxTrackerPort; port++ {
		trackerUrl := buildTrackerURL(torrent, port)

		// Send GET request to tracker
		resp, err := http.Get(trackerUrl.String())
		if err != nil {
			continue // retry with another port
		}
		defer func() {
			if e := resp.Body.Close(); e != nil && err == nil {
				err = e
			}
		}()

		// Read bytes response from tracker client
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		return data, nil
	}

	return nil, errors.New("failed to query tracker")
}

func buildTrackerURL(torrent *torrentfile.SimpleTorrentFile, port int) *url.URL {
	announceUrlCopy := *torrent.Announce
	params := url.Values{
		"info_hash":  []string{string(torrent.InfoHash[:])},
		"peer_id":    []string{string(torrent.PeerID[:])},
		"port":       []string{strconv.Itoa(port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(torrent.Length)},
	}
	announceUrlCopy.RawQuery = params.Encode()
	return &announceUrlCopy
}
