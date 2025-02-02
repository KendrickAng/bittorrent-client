package main

import (
	"errors"
	"flag"
	"strings"
)

var (
	flagTorrent = flag.String("torrent", "",
		"The absolute path to the .torrent file of the torrent to download, e.g. 'mymovie.torrent'.")
	flagMagnet = flag.String("magnet", "",
		"The absolute path to the file containing the magnet link to download.")
)

type Flags struct {
	TorrentFileName string
	MagnetFileName  string
}

func parseFlags() (Flags, error) {
	// Retrieve flags
	flags := Flags{
		TorrentFileName: strings.TrimSpace(*flagTorrent),
		MagnetFileName:  strings.TrimSpace(*flagMagnet),
	}

	if flags.MagnetFileName == "" && flags.TorrentFileName == "" {
		return Flags{}, errors.New("--torrent or --magnet is required")
	}

	if flags.TorrentFileName != "" && flags.MagnetFileName != "" {
		return Flags{}, errors.New("only one of --torrent and --magnet is accepted")
	}

	return flags, nil
}

func init() {
	flag.Parse()
}
