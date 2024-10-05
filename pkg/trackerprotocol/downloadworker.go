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

// TODO: refactor this into a state machine e.g. switch statement
// Start starts the worker downloading available pieces from a client.
func (d *DownloadWorker) Start(ctx context.Context) error {
	for req := range d.requests {
		// If choked, try to unchoke ourselves
		if d.client.IsChoked() {
			if err := d.client.SendInterestedMessage(); err != nil {
				d.handleError(err, req)
				continue
			}
			if _, err := d.client.ReceiveUnchokeMessage(); err != nil {
				d.handleError(err, req)
				continue
			}
			d.client.SetChoked(false)
			println(d.client.String(), "unchoked")
		}

		numRequests := int(math.Ceil(float64(req.pieceLength) / float64(req.requestLength)))
		index := uint32(req.pieceIndex)
		blocks := make([][]byte, numRequests)
		//reqLength := uint32(req.pieceLength) TODO for some reason using this makes it not work.

		for i := 0; i < numRequests; i++ {
			begin := uint32(req.requestLength * i)

			if err := d.client.SendRequestMessage(index, begin, maxRequestLength); err != nil {
				return err
			}

		Inner:
			for {
				msg, err := d.client.ReceiveMessage()
				if err != nil {
					return err
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
						return errors.New("invalid piece")
					}
					blocks[i] = piece.Block
					break Inner
				default:
					panic(msg.ID)
				}
			}
		}

		// Unchoked, start downloading pieces
		//blocks := make([][]byte, numRequests)
		//for i := 0; i < numRequests; i++ {
		//	// Last request may be smaller than the others
		//	if i == numRequests-1 {
		//		// TODO
		//	}
		//	index := uint32(req.pieceLength)
		//	begin := uint32(req.requestLength * i)
		//	length := uint32(req.pieceLength)
		//	if err := d.client.SendRequestMessage(index, begin, length); err != nil {
		//		d.handleError(err, req)
		//		continue
		//	}
		//	pieceMsg, err := d.client.ReceivePieceMessage()
		//	if err != nil {
		//		d.handleError(err, req)
		//		continue
		//	}
		//	println(d.client.String(), " received piece message")
		//	if pieceMsg.Begin != begin || pieceMsg.Index != index {
		//		d.handleError(errors.New("invalid piece"), req)
		//		continue
		//	}
		//	blocks[i] = pieceMsg.Block
		//}

		var finalBlocks []byte
		for _, block := range blocks {
			finalBlocks = append(finalBlocks, block...)
		}

		d.results <- &pieceResult{
			piece: finalBlocks,
			index: req.pieceIndex,
			hash:  hashutil.BTHash(finalBlocks),
		}
	}

	return nil
}

func (d *DownloadWorker) handleError(err error, req *pieceRequest) {
	println("worker ", d.client.String(), " error ", err.Error())
	d.requests <- req
}
