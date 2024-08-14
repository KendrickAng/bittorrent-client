package trackerprotocol

import (
	"context"
	"example.com/btclient/pkg/bencodeutil"
	"fmt"
)

const (
	// Pipelining is encouraged by downloaders to maximise download efficiency.
	maxDownloadThreads uint = 200

	maxRequestLength = 16384 // 2 ^ 14 (16kiB)
)

type DownloadManager struct {
	torrent       *bencodeutil.SimpleTorrentFile
	clients       []*Client
	reconstructer *Reconstructer
	// Channel for sending in pieces of the torrent to download.
	pieceDownloadChannel chan *pieceRequest
	// Channel for receiving pieces of the downloaded torrent.
	pieceResultChannel chan *pieceResult
	done               chan struct{}
}

type pieceRequest struct {
	index             int
	length            int
	expectedPieceHash [20]byte
}

type pieceResult struct {
	piece []byte
	index int
	hash  [20]byte
}

func NewDownloadManager(torrent *bencodeutil.SimpleTorrentFile, clients []*Client) (*DownloadManager, error) {
	return &DownloadManager{
		torrent:              torrent,
		clients:              clients,
		reconstructer:        NewReconstructer(torrent.PieceHashes),
		pieceDownloadChannel: make(chan *pieceRequest, maxDownloadThreads),
		// no upper limit on processing download results
		pieceResultChannel: make(chan *pieceResult),
		done:               make(chan struct{}),
	}, nil
}

func (d *DownloadManager) Start(ctx context.Context) error {
	// split pieces into pieces of work
	downloadTasks := createDownloadTasks(d.torrent)
	for _, task := range downloadTasks {
		d.pieceDownloadChannel <- &task
	}

	// handle with clients
	for _, peer := range d.clients {
		peer := peer
		requests := d.pieceDownloadChannel
		results := d.pieceResultChannel
		go func() {
			worker := NewDownloadWorker(peer, requests, results)
			if err := worker.Start(ctx); err != nil {
				fmt.Printf("worker %s encountered error %s", peer.String(), err.Error())
			}
		}()
	}

	// continue until user cancels or download completes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-d.done:
			return nil
		case pieceResult := <-d.pieceResultChannel:
			isDone, err := d.reconstructer.Reconstruct(pieceResult.piece, pieceResult.index)
			if err != nil {
				fmt.Printf("error reconstructing piece with index %d", pieceResult.index)
			}
			if isDone {
				close(d.done)
			}
		}
	}
}

func createDownloadTasks(torrent *bencodeutil.SimpleTorrentFile) []pieceRequest {
	var downloadTasks []pieceRequest

	for i, pieceHash := range torrent.PieceHashes {
		downloadTasks = append(downloadTasks, pieceRequest{
			index:             i,
			length:            maxRequestLength,
			expectedPieceHash: pieceHash,
		})
	}

	return downloadTasks
}
