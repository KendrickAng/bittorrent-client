package trackerprotocol

import (
	"encoding/binary"
	"testing"
)

func TestBitfield_HasBit(t *testing.T) {
	// Arrange
	// 1000 0001 0100 0010
	bitfield := Bitfield([]byte{129, 66})

	if !bitfield.HasBit(0) {
		t.Fatal("invalid bit")
	}
	if !bitfield.HasBit(7) {
		t.Fatal("invalid bit")
	}
	if !bitfield.HasBit(9) {
		t.Fatal("invalid bit")
	}
	if !bitfield.HasBit(14) {
		t.Fatal("invalid bit")
	}

}

func TestBitfield_SetBit(t *testing.T) {
	// Arrange
	// 0000 0000 0000 0000
	bitfield := Bitfield([]byte{0, 0})

	// Act
	// 1000 0001 0100 0010
	bitfield.SetBit(0)
	bitfield.SetBit(7)
	bitfield.SetBit(9)
	bitfield.SetBit(14)

	// Assert
	if binary.BigEndian.Uint16(bitfield) != 33090 {
		t.Fatal("invalid bitfield, got", bitfield)
	}
}

func TestBitfield_String(t *testing.T) {
	if Bitfield([]byte{129, 66}).String() != "1000 0001 0100 0010" {
		t.Fatal("invalid bitfield")
	}
}
