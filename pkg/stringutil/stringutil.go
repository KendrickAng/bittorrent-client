package stringutil

import (
	"fmt"
)

func SplitChunksOf20(s string) ([][20]byte, error) {
	if len(s)%20 != 0 {
		return nil, fmt.Errorf("string length is not a multiple of 20, got %d", len(s))
	}

	splitChunks := SplitChunks(s, 20)

	var chunks [][20]byte
	for i := 0; i < len(splitChunks); i++ {
		// sanity-checking
		if len(splitChunks[i]) != 20 {
			return nil, fmt.Errorf("chunk size is not 20, got %d", len(splitChunks[i]))
		}

		buf := make([]byte, 20)
		copy(buf, []byte(splitChunks[i]))
		chunks = append(chunks, [20]byte(buf))
	}

	return chunks, nil
}

func SplitChunks(s string, chunkSize int) []string {
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
