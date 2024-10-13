package btclient

import (
	"context"
	"example.com/btclient/pkg/bittorrent/client"
	"example.com/btclient/pkg/bittorrent/torrentfile"
	"os"
	"os/signal"
	"syscall"
)

func Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Parse flags
	flags, err := GetFlags()
	if err != nil {
		return err
	}

	// Read .torrent file
	file, err := os.Open(flags.TorrentFileName)
	if err != nil {
		return err
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	// Decode bencoded file
	bencodedData, err := torrentfile.ReadTorrentFile(file)
	if err != nil {
		return err
	}
	torrent, err := bencodedData.Simplify()
	if err != nil {
		return err
	}

	// Handle (blocking)
	handler, err := client.NewClient(torrent)
	if err != nil {
		return err
	}
	if _, err := handler.Handle(ctx); err != nil {
		return err
	}
	defer handler.Close()

	// Wait until SIGINT is given, or the handler succeeds
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		defer cancel()
		<-signals
	}()

	return nil
}
