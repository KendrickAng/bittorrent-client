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
		torrent:            torrent,
		clients:            clients,
		reconstructer:      NewReconstructer(torrent.PieceHashes),
		pieceResultChannel: make(chan *pieceResult, maxResultThreads),
		done:               make(chan struct{}),
	}, nil
}

func (d *DownloadManager) Start(ctx context.Context) error {
	// split pieces into pieces of work
	downloadTasks := createDownloadTasks(d.torrent)
	pieceDownloadChannel := make(chan *pieceRequest, len(downloadTasks))
	for _, task := range downloadTasks {
		pieceDownloadChannel <- &task
	}
	defer close(pieceDownloadChannel)

	// TODO for now just use one goroutine
	// handle with clients
	worker := NewDownloadWorker(d.clients[0], pieceDownloadChannel, d.pieceResultChannel)
	if err := worker.Start(ctx); err != nil {
		return err
	}

	// Create one handler per peer
	//for _, peer := range d.clients {
	//	go func(peer *Client, requests chan *pieceRequest, results chan *pieceResult) {
	//		worker := NewDownloadWorker(peer, requests, results)
	//		if err := worker.Start(ctx); err != nil {
	//			println("worker ", peer.String(), " error ", err.Error())
	//		}
	//	}(peer, pieceDownloadChannel, d.pieceResultChannel)
	//}

	// continue until user cancels or download completes
	var results []*pieceResult
	for result := range d.pieceResultChannel {
		results = append(results, result)
	}

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
		downloadTasks = append(downloadTasks, pieceRequest{
			pieceIndex:        i,
			requestLength:     maxRequestLength,
			pieceLength:       torrent.PieceLength,
			expectedPieceHash: pieceHash,
		})
	}

	return downloadTasks
}
