package main

import (
	"flag"
	"fmt"
	"slices"
	"strings"
	"sync"
)

const (
	typeMagnet  string = "magnet"
	typeTorrent string = "torrent"
)

var (
	parseFlags = sync.OnceFunc(flag.Parse)

	acceptedTypes = []string{typeMagnet, typeTorrent}

	flagType = flag.String("type", typeTorrent,
		fmt.Sprintf("Whether to parse the input as a torrent file or a magnet link. Accepted values: %s", strings.Join(acceptedTypes, ",")))
)

type Flags struct {
	FileName string
	Type     string
}

func (f Flags) IsInputMagnetLink() bool {
	return f.Type == typeMagnet
}

func (f Flags) IsInputTorrentFile() bool {
	return f.Type == typeTorrent
}

func getFlags() (Flags, error) {
	parseFlags()

	// Retrieve flags
	flags := Flags{
		FileName: flag.Arg(0),
		Type:     strings.TrimSpace(*flagType),
	}

	if err := validate(flags); err != nil {
		return Flags{}, err
	}

	return flags, nil
}

func validate(f Flags) error {
	if !slices.Contains(acceptedTypes, f.Type) {
		return fmt.Errorf("invalid input %s, only %v is supported", f.Type, acceptedTypes)
	}
	return nil
}
