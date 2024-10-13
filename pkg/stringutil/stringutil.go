package stringutil

import (
	"example.com/btclient/pkg/preconditions"
	"fmt"
)

// ChunksOf20 splits s into a slice of byte arrays of size 20 each.
func ChunksOf20(s string) ([][20]byte, error) {
	if len(s)%20 != 0 {
		return nil, fmt.Errorf("string length is not a multiple of 20, got %d", len(s))
	}

	splitChunks := chunksOfN(s, 20)
	chunks := make([][20]byte, len(splitChunks))
	for i, chunk := range splitChunks {
		preconditions.CheckArgumentf(len(chunk) == 20, "chunk size must be 20, got %d", len(chunk))
		chunks[i] = ([20]byte)([]byte(chunk))
	}

	return chunks, nil
}

func chunksOfN(s string, chunkSize int) []string {
	var chunks []string
	for len(s) > 0 {
		if len(s) >= chunkSize {
			chunks = append(chunks, s[:chunkSize])
			s = s[chunkSize:]
		} else {
			chunks = append(chunks, s)
			s = ""
		}
	}
	return chunks
}
