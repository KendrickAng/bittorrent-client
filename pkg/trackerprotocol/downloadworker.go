package trackerprotocol

import (
	"context"
	"errors"
)

type DownloadWorker struct {
	client   *Client
	requests chan *pieceRequest
	results  chan *pieceResult
}

func NewDownloadWorker(client *Client, requests chan *pieceRequest, results chan *pieceResult) *DownloadWorker {
	return &DownloadWorker{
		client:   client,
		requests: requests,
		results:  results,
	}
}

// Starts the worker downloading available pieces from a client.
func (d *DownloadWorker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-d.requests:
			// TODO implement
			return errors.New("DOWNLOAD WORKER NOT IMPLEMENTED YET!!")
		}
	}
}
