package trackerprotocol

import (
	"slices"
	"strings"
)

// Bitfield represents the pieces that a peer has available for download.
type Bitfield []byte

func (b Bitfield) HasBit(bitIdx int) bool {
	byteIdx := bitIdx / 8
	bitOffset := bitIdx % 8
	a := b[byteIdx] >> (7 - bitOffset)
	return (a | 0) != 0
}

func (b Bitfield) SetBit(bitIdx int) {
	byteIdx := bitIdx / 8
	bitOffset := bitIdx % 8
	b[byteIdx] |= 1 << (7 - bitOffset)
}

func (b Bitfield) String() string {
	var s []string
	for _, bb := range b {
		s = append(s, byteToString(bb))
	}
	return strings.Join(s, " ")
}

func byteToString(b uint8) string {
	s := make([]string, 8)
	for i := 7; i >= 0; i-- {
		if b&1 == 1 {
			s[i] = "1"
		} else {
			s[i] = "0"
		}
		b = b >> 1
	}
	s = slices.Insert(s, len(s)/2, " ")
	return strings.Join(s, "")
}
