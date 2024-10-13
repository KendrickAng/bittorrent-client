package torrentfile

// SimpleTorrentFile represents a simplified, flattened version of [TorrentFile].
type SimpleTorrentFile struct {
	Announce string
	// SHA-1 hash of the entire bencoded info dict.
	InfoHash [20]byte
	// Hash of each piece.
	PieceHashes [][20]byte
	// Number of bytes in each piece.
	PieceLength int
	// Length of the file in bytes.
	Length int
	// May be empty in multi-file mode.
	Name string
}
