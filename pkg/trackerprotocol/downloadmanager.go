package trackerprotocol

import (
	"bytes"
	"context"
	"example.com/btclient/pkg/bencodeutil"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
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
		torrent: torrent,
		clients: clients,
		done:    make(chan struct{}),
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
					// skip if client doesn't have the piece
					if !client.bitfield.HasBit(downloadTask.pieceIndex) {
						downloadTasksChan <- downloadTask
						time.Sleep(1 * time.Second) // prevent starvation
						continue
					}

					// have client download the piece
					result, err := NewDownloadWorker(client).Start(ctx, downloadTask)
					if err != nil {
						downloadTasksChan <- downloadTask
					} else if !bytes.Equal(d.torrent.PieceHashes[result.index][:], result.hash[:]) {
						// TODO: fix the bug where the last piece has an invalid hash.
						println("invalid piece hash for piece", result.index)
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
	pieceHashes := d.torrent.PieceHashes
	pieces := make([][]byte, len(pieceHashes))
	for result := range downloadResultsChan {
		if pieceHashes[result.index] != result.hash {
			return fmt.Errorf("invalid hash: expected: %x, got: %x", result.hash, pieceHashes[result.index])
		}
		pieces[result.index] = result.piece
	}

	// flatten
	var original []byte
	for _, f := range pieces {
		original = append(original, f...)
	}

	// write the downloaded bytes to disk
	f, err := os.OpenFile(d.torrent.Name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	n, err := f.Write(original)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	absPath, err := filepath.Abs(filepath.Join(cwd, f.Name()))
	if err != nil {
		return err
	}
	fmt.Printf("wrote %d bytes to %s\n", n, absPath)

	return nil
}

func createDownloadTasks(torrent *bencodeutil.SimpleTorrentFile) []pieceRequest {
	var downloadTasks []pieceRequest

	// TODO: this logic should be tested
	for i, pieceHash := range torrent.PieceHashes {
		pieceLength := torrent.PieceLength

		// Last piece may be smaller than piece length
		if (i == len(torrent.PieceHashes)-1) && (torrent.Length%torrent.PieceLength != 0) {
			pieceLength = torrent.Length % torrent.PieceLength
		}

		request := createDownloadTask(i, pieceLength, pieceHash)
		downloadTasks = append(downloadTasks, request)
	}

	return downloadTasks
}

func createDownloadTask(pieceIndex int, pieceLength int, expectedPieceHash [20]byte) pieceRequest {
	return pieceRequest{
		pieceIndex:        pieceIndex,
		requestLength:     maxRequestLength,
		pieceLength:       pieceLength,
		expectedPieceHash: expectedPieceHash,
	}
}
