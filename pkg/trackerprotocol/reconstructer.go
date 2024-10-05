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

func (r *Reconstructer) Reconstruct(results []*pieceResult) ([]byte, error) {
	var final []byte
	for i, result := range results {
		if r.pieceHashes[i] != result.hash {
			return nil, errors.New("invalid piece hash")
		}
		final = append(final, result.piece...)
	}
	return final, nil
}
