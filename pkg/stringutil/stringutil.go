package stringutil

import (
	"crypto/rand"
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

// Random20Bytes generates a random byte array of 20 bytes.
func Random20Bytes() ([20]byte, error) {
	b, err := randomBytes(20)
	if err != nil {
		return [20]byte{}, err
	}
	return [20]byte(b), nil
}

// Random8Bytes generates a random byte array of 8 bytes.
func Random8Bytes() ([8]byte, error) {
	b, err := randomBytes(8)
	if err != nil {
		return [8]byte{}, err
	}
	return [8]byte(b), nil
}

func randomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}
