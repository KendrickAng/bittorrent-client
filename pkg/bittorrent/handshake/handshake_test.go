package handshake

import (
	"bytes"
	"example.com/btclient/pkg/stringutil"
	"io"
	"net"
	"testing"
)

func TestHandshake_SendHandshake(t *testing.T) {
	// Arrange
	reader, writer := net.Pipe()
	handshaker := NewHandshaker(writer)

	extensionBits, err := stringutil.Random8Bytes()
	if err != nil {
		t.Fatal(err)
	}
	peerId, err := stringutil.Random20Bytes()
	if err != nil {
		t.Fatal(err)
	}
	infoHash, err := stringutil.Random20Bytes()
	if err != nil {
		t.Fatal(err)
	}

	// Act
	go func() {
		if err := handshaker.SendHandshake(extensionBits, peerId, infoHash); err != nil {
			t.Fatal(err)
		}

		writer.Close()
		reader.Close()
	}()

	// Assert
	b, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if len(b) != 68 {
		t.Fatal("wrong length")
	}
	pstrlen := int(b[0])
	pstr := string(b[1 : 1+pstrlen])
	actualExtensionBits := b[1+pstrlen : 1+pstrlen+8]
	actualInfoHash := b[1+pstrlen+8 : 1+pstrlen+8+20]
	actualPeerId := b[1+pstrlen+8+20 : 1+pstrlen+8+20+20]
	if pstrlen != 19 {
		t.Fatal("wrong pstrlen")
	}
	if pstr != "BitTorrent protocol" {
		t.Fatal("wrong pstr")
	}
	if !bytes.Equal(actualExtensionBits, extensionBits[:]) {
		t.Fatal("wrong extension bits")
	}
	if !bytes.Equal(infoHash[:], actualInfoHash[:]) {
		t.Fatal("wrong infohash")
	}
	if !bytes.Equal(peerId[:], actualPeerId[:]) {
		t.Fatal("wrong peerid")
	}
}
