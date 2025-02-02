package client

import (
	"bytes"
	"context"
	"example.com/btclient/internal/bittorrent/peer"
	"example.com/btclient/internal/bittorrent/torrentfile"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TcpClient represents a torrent downloader that uses TCP for datareader download from peers.
type TcpClient struct {
	connectionPool *peer.Pool
}

func NewTcpClient(connectionPool *peer.Pool) *TcpClient {
	return &TcpClient{connectionPool: connectionPool}
}

func (h *TcpClient) Download(ctx context.Context, torrent *torrentfile.SimpleTorrentFile) (resp *Response, err error) {
	// split pieces into pieces of work
	downloadTasks := createDownloadTasks(torrent)
	downloadTasksChan := make(chan pieceRequest, len(downloadTasks))
	for _, downloadTask := range downloadTasks {
		downloadTasksChan <- downloadTask
	}

	// start a goroutine for each client to download from
	downloadResultsChan := make(chan *pieceResult, len(downloadTasks))
	wg2 := new(sync.WaitGroup)
	wg2.Add(len(downloadTasks))
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		wg2.Wait()
		println("download completed")
		close(downloadResultsChan)
		cancel()
	}()

	for _, btclient := range h.connectionPool.GetClients() {
		go func(ctx context.Context, btclient *peer.Client) {
			for {
				select {
				case <-ctx.Done():
					return
				case downloadTask := <-downloadTasksChan:
					// skip if client doesn't have the piece
					if !btclient.Bitfield.HasBit(downloadTask.pieceIndex) {
						downloadTasksChan <- downloadTask
						time.Sleep(1 * time.Second) // prevent starvation
						continue
					}

					// have client download the piece
					result, err := newDownloadWorker(btclient).start(ctx, downloadTask)
					if err != nil {
						downloadTasksChan <- downloadTask
					} else if !bytes.Equal(torrent.PieceHashes[result.index][:], result.hash[:]) {
						println("invalid piece hash for piece", result.index)
						downloadTasksChan <- downloadTask
					} else {
						downloadResultsChan <- result
						wg2.Done()
					}
				}
			}
		}(ctx, btclient)
	}

	// TODO save file to disk once download is completed
	// TODO optimize by writing parts to disk as they arrive, as well
	// blocking reconstruct that completes when wait group is done
	pieceHashes := torrent.PieceHashes
	pieces := make([][]byte, len(pieceHashes))
	for result := range downloadResultsChan {
		if pieceHashes[result.index] != result.hash {
			return nil, fmt.Errorf("invalid hash: expected: %x, got: %x", result.hash, pieceHashes[result.index])
		}
		pieces[result.index] = result.piece
	}

	// flatten
	var original []byte
	for _, f := range pieces {
		original = append(original, f...)
	}

	// write the downloaded bytes to disk
	f, err := os.OpenFile(torrent.Name, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	n, err := f.Write(original)
	if err != nil {
		return nil, err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	absPath, err := filepath.Abs(filepath.Join(cwd, f.Name()))
	if err != nil {
		return nil, err
	}
	fmt.Printf("wrote %d bytes to %s\n", n, absPath)

	return &Response{
		NumDownloadedBytes: n,
	}, nil
}

func createDownloadTasks(torrent *torrentfile.SimpleTorrentFile) []pieceRequest {
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
