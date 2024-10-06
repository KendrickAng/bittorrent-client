package trackerprotocol

import (
	"errors"
)

// Reconstructer gathers the pieces of a torrent file and reconstructs the downloaded file.
type Reconstructer struct {
	pieceHashes [][20]byte
	pieces      [][]byte
	received    int
}

func NewReconstructer(pieceHashes [][20]byte) *Reconstructer {
	return &Reconstructer{
		pieceHashes: pieceHashes,
		pieces:      make([][]byte, len(pieceHashes)),
		received:    0,
	}
}

func (r *Reconstructer) Reconstruct(resultsChan chan *pieceResult) ([]byte, error) {
	pieces := make([][]byte, len(r.pieceHashes))
	for result := range resultsChan {
		if r.pieceHashes[result.index] != result.hash {
			return nil, errors.New("invalid piece hash")
		}
		pieces[result.index] = result.piece
	}

	// flatten
	var original []byte
	for _, f := range pieces {
		original = append(original, f...)
	}

	return original, nil
}
