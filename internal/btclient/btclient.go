package btclient

import (
	"example.com/btclient/pkg/bencodeutil"
	"example.com/btclient/pkg/closelogger"
	"fmt"
	"os"
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

	fmt.Printf("%+v\n", bencodedData)

	return nil
}
