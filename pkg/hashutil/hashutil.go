package hashutil

import "crypto/sha1"

// BTHash returns the expected Bittorrent hash of a piece of data.
// The input slice is expected to be non-nil.
func BTHash(data []byte) [20]byte {
	if data == nil {
		panic("cannot hash nil slice")
	}
	return sha1.Sum(data)
}
