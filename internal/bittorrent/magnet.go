package bittorrent

import (
	"encoding/base32"
	"encoding/hex"
	"errors"
	"example.com/btclient/internal/preconditions"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Magnet represents a BitTorrent magnet link.
// Only the V1 magnet link format is supported.
// v1: magnet:?xt=urn:btih:<info-hash>&dn=<name>&tr=<tracker-url>&x.pe=<peer-address>
// v2: magnet:?xt=urn:btmh:<tagged-info-hash>&dn=<name>&tr=<tracker-url>&x.pe=<peer-address>
type Magnet struct {
	// The original magnet link (v1 format only).
	uri *url.URL

	// REQUIRED. The info hash of the info dictionary, but hex-encoded (length 40).
	// Clients should also support the 32-character base32-encoded info-hash.
	infoHashString string
	// OPTIONAL. Display name that may be used by the client to display while waiting for metadata.
	displayName string
	// OPTIONAL. Tracker URLs.
	trackers []*url.URL
	// OPTIONAL. Peer address. Addr may be
	peers []peerAddress
}

func (m *Magnet) TrackerUrls() []*url.URL {
	return m.trackers
}

func (m *Magnet) InfoHash() ([20]byte, error) {
	var b []byte
	var err error
	if len(m.infoHashString) == 40 {
		b, err = hex.DecodeString(m.infoHashString)
	} else if len(m.infoHashString) == 32 {
		b, err = base32.StdEncoding.DecodeString(m.infoHashString)
	} else {
		panic(fmt.Errorf("got unexpected info hash of length %d", len(m.infoHashString)))
	}
	if err != nil {
		return [20]byte{}, err
	}
	return [20]byte(b), err
}

func (m *Magnet) DisplayName() string {
	return m.displayName
}

func (m *Magnet) Validate() error {
	if len(m.infoHashString) != 40 && len(m.infoHashString) != 32 {
		return fmt.Errorf("only info hashes of length 40/32 are supported, got %d", len(m.infoHashString))
	}
	if len(m.trackers) <= 0 {
		return fmt.Errorf("at least one tracker must be specified, DHT is not supported yet")
	}
	return nil
}

// Peer Address expressed as hostname:port, ipv4-literal:port or [ipv6-literal]:port.
// For initiating a direct metadata transfer between two clients, reducing the need for external peers.
// Should only be included if the client can discover its public IP address and determine its reachability.
type peerAddress struct {
	// The part before the ':'
	prefix string
	port   uint16
}

func ParseMagnet(magnet string) (*Magnet, error) {
	u, err := url.Parse(magnet)
	if err != nil {
		return nil, err
	}

	// retrieve info hash
	xt := u.Query().Get("xt")
	if strings.HasPrefix(xt, "urn:btmh") {
		return nil, errors.New("v2 magnet link not supported")
	} else if !strings.HasPrefix(xt, "urn:btih") {
		return nil, fmt.Errorf("unknown magnet link %s", xt)
	}
	infoHash := strings.TrimPrefix(xt, "urn:btih:")

	// retrieve display name
	displayName := u.Query().Get("dn")

	// retrieve trackers (if any)
	trackerStrings := u.Query()["tr"]
	trackers := make([]*url.URL, len(trackerStrings))
	for i, trackerString := range trackerStrings {
		tracker, err := url.Parse(trackerString)
		if err != nil {
			return nil, err
		}
		trackers[i] = tracker
	}

	// retrieve peer addresses
	peerAddrStrings := u.Query()["x.pe"]
	peers := make([]peerAddress, len(peerAddrStrings))
	for i, peerAddrString := range peerAddrStrings {
		tokens := strings.Split(peerAddrString, ":")
		preconditions.CheckArgument(len(tokens) == 2, "invalid peer address")
		prefix := strings.TrimFunc(tokens[0], func(r rune) bool { return r == '[' || r == ']' })
		port, err := strconv.ParseUint(tokens[1], 10, 16)
		if err != nil {
			return nil, err
		}
		peers[i] = peerAddress{
			prefix: prefix,
			port:   uint16(port),
		}
	}

	// create and validate magnet
	m := &Magnet{uri: u,
		infoHashString: infoHash,
		displayName:    displayName,
		trackers:       trackers,
		peers:          peers,
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}

	return m, nil
}
