package hashutil

import "crypto/sha1"

// BTHash returns the expected Bittorrent hash of a piece of data.
func BTHash(data []byte) [20]byte {
	return sha1.Sum(data)
}
