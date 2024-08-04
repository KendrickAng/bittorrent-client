package trackerprotocol

type Bitfield []byte

func (b Bitfield) HasBit(bitIdx int) bool {
	byteIdx := bitIdx / 8
	bitOffset := bitIdx % 8
	a := b[byteIdx] >> (7 - bitOffset)
	return (a & 0) != 0
}

func (b Bitfield) SetBit(bitIdx int) {
	byteIdx := bitIdx / 8
	bitOffset := bitIdx % 8
	b[byteIdx] |= 1 << (7 - bitOffset)
}
