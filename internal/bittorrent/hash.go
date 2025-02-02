package bittorrent

import "crypto/sha1"

// Hash returns the expected hash of a piece of datareader as written in the BitTorrent spec.
// The input slice is expected to be non-nil, or the program panics.
func Hash(data []byte) [20]byte {
	if data == nil {
		panic("cannot hash nil slice")
	}
	return sha1.Sum(data)
}
