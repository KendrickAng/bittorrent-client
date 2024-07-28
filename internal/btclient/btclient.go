package btclient

import (
	"example.com/btclient/pkg/bencodeutil"
	"example.com/btclient/pkg/closelogger"
	"example.com/btclient/pkg/trackerprotocol"
	"os"
	"os/signal"
	"syscall"
)

func Run() error {
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
	defer closelogger.CloseOrLog(file, flags.TorrentFileName)

	// Decode bencoded file
	bencodedData, err := bencodeutil.Unmarshal(file)
	if err != nil {
		return err
	}
	torrent, err := bencodedData.Simplify()
	if err != nil {
		return err
	}

	// Handle
	handler, err := trackerprotocol.NewHandler(torrent)
	if err != nil {
		return err
	}
	if err := handler.Handle(); err != nil {
		return err
	}
	defer handler.Close()

	// Wait until SIGINT is given
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	<-signals

	return nil
}
