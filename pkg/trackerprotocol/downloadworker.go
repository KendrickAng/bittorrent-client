package trackerprotocol

import (
	"context"
	"errors"
	"example.com/btclient/pkg/hashutil"
	"math"
)

// DownloadWorker handles the download of a single piece of data in the torrent.
// A torrent is split into many pieces for download.
type DownloadWorker struct {
	client *Client
}

func NewDownloadWorker(client *Client) *DownloadWorker {
	return &DownloadWorker{
		client: client,
	}
}

// Start starts the worker downloading available pieces from a client.
func (d *DownloadWorker) Start(ctx context.Context, requests []pieceRequest) ([]*pieceResult, error) {
	pieceResults := make([]*pieceResult, len(requests))

	for i, req := range requests {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// If choked, try to unchoke ourselves
		if d.client.IsChoked() {
			if err := d.client.SendInterestedMessage(); err != nil {
				return nil, err
			}
			if _, err := d.client.ReceiveUnchokeMessage(); err != nil {
				return nil, err
			}
			d.client.SetChoked(false)
			println(d.client.String(), "unchoked")
		}

		remainingBytes := req.pieceLength
		numRequests := int(math.Ceil(float64(req.pieceLength) / float64(req.requestLength)))
		index := uint32(req.pieceIndex)
		blocks := make([][]byte, numRequests)

		for i := 0; i < numRequests; i++ {
			begin := uint32(req.requestLength * i)
			reqLength := uint32(min(req.requestLength, remainingBytes))

			if err := d.client.SendRequestMessage(index, begin, reqLength); err != nil {
				return nil, err
			}

		Inner:
			for {
				msg, err := d.client.ReceiveMessage()
				if err != nil {
					return nil, err
				}
				switch msg.ID {
				case MsgKeepAlive:
					println("keep alive")
				case MsgChoke:
					d.client.SetChoked(true)
				case MsgUnchoke:
					d.client.SetChoked(false)
				case MsgBitfield:
					d.client.SetBitfield(msg.AsMsgBitfield().Bitfield)
				case MsgPiece:
					piece := msg.AsMsgPiece()
					println("piece", index, ":", begin, "of length", len(piece.Block), "of total", req.pieceLength)
					if piece.Begin != begin || piece.Index != index {
						return nil, errors.New("invalid piece")
					}
					blocks[i] = piece.Block
					remainingBytes -= len(piece.Block)
					break Inner
				default:
					panic(msg.ID)
				}
			}
		}

		var finalBlocks []byte
		for _, block := range blocks {
			finalBlocks = append(finalBlocks, block...)
		}

		pieceResults[i] = &pieceResult{
			piece: finalBlocks,
			index: req.pieceIndex,
			hash:  hashutil.BTHash(finalBlocks),
		}
	}

	return pieceResults, nil
}
