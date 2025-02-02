package tracker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
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

func (h *HttpClient) FetchTorrentMetadata(req FetchTorrentMetadataRequest) (*Response, error) {
	if req.TrackerUrl.Scheme != "http" {
		return nil, fmt.Errorf("only http is supported, got scheme %s", req.TrackerUrl.Scheme)
	}

	trackerResponse, err := fetchTorrentMetadataFromTracker(req.TrackerUrl, req.InfoHash, req.PeerID, req.Left)
	if err != nil {
		return nil, err
	}

	return readResponse(bufio.NewReader(bytes.NewReader(trackerResponse)))
}

func fetchTorrentMetadataFromTracker(trackerUrl *url.URL,
	infoHash, peerID [20]byte, left int) (data []byte, error error) {

	for port := minTrackerPort; port <= maxTrackerPort; port++ {
		trackerUrl := buildTrackerURL(trackerUrl, infoHash, peerID, left, port)

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

func buildTrackerURL(trackerUrl *url.URL,
	infoHash [20]byte,
	peerID [20]byte,
	left int,
	port int,
) *url.URL {

	announceUrlCopy := *trackerUrl
	params := url.Values{
		"info_hash":  []string{string(infoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(port)},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(left)},
	}
	announceUrlCopy.RawQuery = params.Encode()
	return &announceUrlCopy
}
