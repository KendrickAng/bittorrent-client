package bittorrent

import "testing"

func TestNewExtensionBits(t *testing.T) {
	ext := NewExtensionBits(ExtensionProtocolBit)
	extByteArr := [8]byte(ext)
	res := int(extByteArr[5] & 0x10)

	if res != 0x10 {
		t.Fatal()
	}
}

func TestExtensionBits_HasExtensionProtocolBit(t *testing.T) {
	ext := NewExtensionBits(ExtensionProtocolBit)

	if !ext.HasExtensionProtocolBit() {
		t.Fatal()
	}
}
