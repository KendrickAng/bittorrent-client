package btclient

import (
	"flag"
	"fmt"
	"strings"
)

var (
	flagTorrent = flag.String("torrent", "tears-of-steel.torrent",
		"The absolute path to the .torrent file of the torrent to download, e.g. 'mymovie.torrent'.")
)

type Flags struct {
	TorrentFileName string
}

func GetFlags() (Flags, error) {
	// Retrieve flags
	flags := Flags{
		TorrentFileName: strings.TrimSpace(*flagTorrent),
	}

	// Parse flags
	if !strings.HasSuffix(flags.TorrentFileName, ".torrent") {
		return Flags{}, fmt.Errorf("file must end with .torrent, got %s", flags.TorrentFileName)
	}

	return flags, nil
}

func init() {
	flag.Parse()
}
