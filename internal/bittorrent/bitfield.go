package bittorrent

import (
	"slices"
	"strings"
)

// Bitfield represents the pieces that a peer has available for download, as a sequence of bits.
// The first byte of the bitfield corresponds to indices 0 - 7 from high bit to low bit, respectively.
// The next one 8-15, etc. Spare bits at the end are set to zero.
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

func (b Bitfield) Validate() error {
	// TODO throw an error if the Bitfield is incorrect.
	// From the docs: A Bitfield of the wrong length is considered an error.
	// Clients should drop the connection if they receive bitfields that are not of the correct size,
	// or if the Bitfield has any of the spare bits set.
	return nil
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
