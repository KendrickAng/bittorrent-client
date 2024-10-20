package bittorrent

import (
	"net/url"
	"testing"
)

func TestParseMagnet(t *testing.T) {
	magnetLink := "magnet:?xt=urn:btih:d69f91e6b2ae4c542468d1073a71d4ea13879a7f&dn=sample.torrent&tr=http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce"
	magnet, err := ParseMagnet(magnetLink)
	if err != nil {
		t.Fatal(err)
	}

	if magnet.infoHashString != "d69f91e6b2ae4c542468d1073a71d4ea13879a7f" {
		t.Fatal("incorrect infohash", magnet.infoHashString)
	}

	if magnet.displayName != "sample.torrent" {
		t.Fatal("incorrect displayname", magnet.displayName)
	}

	if len(magnet.trackers) != 1 {
		t.Fatal("incorrect trackers", len(magnet.trackers))
	}

	u, err := url.QueryUnescape("http%3A%2F%2Fbittorrent-test-tracker.codecrafters.io%2Fannounce")
	if err != nil {
		t.Fatal(err)
	}
	if magnet.trackers[0].String() != u {
		t.Fatal("incorrect trackers", magnet.trackers[0].String())
	}
}
