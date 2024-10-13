package client

import (
	"bytes"
	"context"
	"errors"
	"example.com/btclient/pkg/bittorrent/handshake"
	"example.com/btclient/pkg/bittorrent/peer"
	"example.com/btclient/pkg/bittorrent/torrentfile"
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TcpClient represents a torrent downloader that uses TCP for data download from peers.
type TcpClient struct {
	Client
}

func (h *TcpClient) Download(ctx context.Context, torrent *torrentfile.SimpleTorrentFile) (resp *Response, err error) {
	// Parse tracker response
	trackerResp, err := h.tracker.FetchTorrentMetadata(torrent)
	if err != nil {
		return nil, err
	}
	if len(trackerResp.Peers) == 0 {
		return nil, errors.New("no peers found")
	}
	println("parsed tracker response")

	// Connect to available peers
	clientsCh := make(chan *peer.Client, len(trackerResp.Peers))
	println("attempting connection to", len(trackerResp.Peers), "peers")
	wg := new(sync.WaitGroup)
	for _, addrPort := range trackerResp.Peers {
		wg.Add(1)

		go func(peer2 netip.AddrPort) {
			defer wg.Done()

			// dial peer
			conn, err := net.DialTimeout("tcp", addrPort.String(), 30*time.Second)
			if err != nil {
				println("error creating client for peer", peer2.String())
				return
			}
			println("dialed", conn.RemoteAddr().String())

			// create client to peer
			client := peer.NewClient(conn, conn, handshake.NewHandshaker(conn), torrent.PeerID, torrent.InfoHash)
			if err := client.Init(); err != nil {
				println("error creating client for peer", peer2.String())
				return
			}

			clientsCh <- client
			fmt.Printf("created client for %s\n", peer2.String())
		}(addrPort)
	}
	wg.Wait()
	close(clientsCh) // close channel so we don't loop over it infinitely

	// Convert peers channel into peers queue
	var clients []*peer.Client
	for client := range clientsCh {
		clients = append(clients, client)
	}
	fmt.Printf("found %d peers\n", len(clients))

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

	for _, btclient := range clients {
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
						// TODO: fix the bug where the last piece has an invalid hash.
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
