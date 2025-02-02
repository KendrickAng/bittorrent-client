package bittorrent

import "math"

// ExtensionBits represents the bits sent during a BitTorrent handshake with a peer.
type ExtensionBits [8]byte

// extension represents a particular extension supported by BitTorrent.
type extension int

const (
	// ExtensionProtocolBit represents http://bittorrent.org/beps/bep_0009.html.
	ExtensionProtocolBit extension = 20
)

func NewExtensionBits(extensions ...extension) ExtensionBits {
	extensionBits := ExtensionBits(make([]byte, 8))
	for _, extension := range extensions {
		extensionBits.setBit(int(extension))
	}
	return extensionBits
}

// hasBit returns true if the bit at position n is set. Counting starts at 0 from the right.
func (e *ExtensionBits) hasBit(n int) bool {
	byteIdx := 8 - int(math.Ceil(float64(n)/8.0))
	hasBit := e[byteIdx] & (1 << (n % 8))
	return hasBit != 0
}

func (e *ExtensionBits) setBit(n int) {
	byteIdx := 8 - int(math.Ceil(float64(n)/8.0))
	e[byteIdx] |= 1 << (n % 8)
}

// HasExtensionProtocolBit returns true if a peer supports the extension Protocol (BEP 10).
// See: https://www.bittorrent.org/beps/bep_0010.html.
func (e *ExtensionBits) HasExtensionProtocolBit() bool {
	return e.hasBit(int(ExtensionProtocolBit))
}
