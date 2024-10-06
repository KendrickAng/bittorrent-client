package trackerprotocol

import (
	"context"
	"example.com/btclient/pkg/bencodeutil"
	"fmt"
	"sync"
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
	downloadTasksChan := make(chan pieceRequest, len(downloadTasks))
	for _, downloadTask := range downloadTasks {
		downloadTasksChan <- downloadTask
	}

	// start a goroutine for each client to download from
	downloadResultsChan := make(chan *pieceResult, len(downloadTasks))
	wg := new(sync.WaitGroup)
	wg.Add(len(downloadTasks))
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		wg.Wait()
		close(downloadResultsChan)
		cancel()
	}()

	for _, client := range d.clients {
		go func(ctx context.Context, client *Client) {
			for {
				select {
				case <-ctx.Done():
					return
				case downloadTask := <-downloadTasksChan:
					result, err := NewDownloadWorker(client).Start(ctx, downloadTask)
					if err != nil {
						downloadTasksChan <- downloadTask
					} else {
						downloadResultsChan <- result
						wg.Done()
					}
				}
			}
		}(ctx, client)
	}

	// TODO save file to disk once download is completed
	// TODO optimize by writing parts to disk as they arrive, as well
	// blocking reconstruct that completes when wait group is done
	final, err := d.reconstructer.Reconstruct(downloadResultsChan)
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
