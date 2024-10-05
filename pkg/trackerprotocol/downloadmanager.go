package trackerprotocol

import (
	"context"
	"example.com/btclient/pkg/bencodeutil"
	"fmt"
)

const (
	// Pipelining is encouraged by downloaders to maximise download efficiency.
	maxDownloadThreads uint = 200
	maxResultThreads   uint = 1000
	maxRequestLength        = 16384 // 2 ^ 14 (16kiB)
)

type DownloadManager struct {
	torrent       *bencodeutil.SimpleTorrentFile
	clients       []*Client
	reconstructer *Reconstructer
	// Channel for receiving pieces of the downloaded torrent.
	pieceResultChannel chan *pieceResult
	done               chan struct{}
}

type pieceRequest struct {
	// Index identifying the piece to download.
	pieceIndex int
	// Size of a piece, in bytes.
	pieceLength int
	// Bytes to download in a single request message, in bytes.
	requestLength     int
	expectedPieceHash [20]byte
}

type pieceResult struct {
	piece []byte
	index int
	hash  [20]byte
}

func NewDownloadManager(torrent *bencodeutil.SimpleTorrentFile, clients []*Client) (*DownloadManager, error) {
	return &DownloadManager{
		torrent:       torrent,
		clients:       clients,
		reconstructer: NewReconstructer(torrent.PieceHashes),
		done:          make(chan struct{}),
	}, nil
}

func (d *DownloadManager) Start(ctx context.Context) error {
	// split pieces into pieces of work
	downloadTasks := createDownloadTasks(d.torrent)

	// TODO make this multi-threaded and download from multiple clients
	// handle with clients
	results, err := NewDownloadWorker(d.clients[0]).Start(ctx, downloadTasks)
	if err != nil {
		return err
	}

	// continue until user cancels or download completes
	final, err := d.reconstructer.Reconstruct(results)
	if err != nil {
		return err
	}
	fmt.Printf("COMPLETED! Got %s\n", string(final))

	return nil
}

func createDownloadTasks(torrent *bencodeutil.SimpleTorrentFile) []pieceRequest {
	var downloadTasks []pieceRequest

	for i, pieceHash := range torrent.PieceHashes {
		// Last piece may be smaller than piece length
		pieceLength := torrent.PieceLength
		if i == len(torrent.PieceHashes)-1 {
			pieceLength = torrent.Length % torrent.PieceLength
		}

		downloadTasks = append(downloadTasks, pieceRequest{
			pieceIndex:        i,
			requestLength:     maxRequestLength,
			pieceLength:       pieceLength,
			expectedPieceHash: pieceHash,
		})
	}

	return downloadTasks
}
