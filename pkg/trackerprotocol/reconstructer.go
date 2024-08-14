package trackerprotocol

import (
	"example.com/btclient/pkg/hashutil"
	"fmt"
)

// Reconstructer gathers the pieces of a torrent file and reconstructs the downloaded file.
type Reconstructer struct {
	pieceHashes [][20]byte
	pieces      [][]byte
}

func NewReconstructer(pieceHashes [][20]byte) *Reconstructer {
	return &Reconstructer{
		pieceHashes: pieceHashes,
		pieces:      make([][]byte, len(pieceHashes)),
	}
}

func (r *Reconstructer) Reconstruct(piece []byte, pieceIdx int) (bool, error) {
	actualPieceHash := hashutil.BTHash(piece)
	expectedPieceHash := r.pieceHashes[pieceIdx]
	if actualPieceHash != expectedPieceHash {
		return false, fmt.Errorf("invalid piece hash for index %d", pieceIdx)
	}
	return true, nil
}
