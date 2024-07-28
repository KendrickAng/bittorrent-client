package bencodeutil

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io"
)

type TrackerResponse struct {
	FailureReason string `bencode:"failure reason,omitempty"`
	Interval      int    `bencode:"interval"`
	Peers         string `bencode:"peers"`
}

type Peer struct {
	PeerID string `bencode:"peer id"`
	IP     string `bencode:"ip"`
	Port   int    `bencode:"port"`
}

func UnmarshalTrackerResponse(b []byte) (*TrackerResponse, error) {
	var response TrackerResponse
	if err := bencode.Unmarshal(bytes.NewReader(b), &response); err != nil {
		return nil, err
	}

	// TODO move this out
	buffer := bytes.NewBuffer([]byte(response.Peers))

	buffer2 := make([]byte, 6)
	_, err := buffer.Read(buffer2)
	for err != io.EOF {
		ipAddr := binary.BigEndian.Uint32(buffer2[:4])
		port := binary.BigEndian.Uint16(buffer2[4:])
		fmt.Printf("IP: %012d, Port: %d\n", ipAddr, port)

		_, err = buffer.Read(buffer2)
	}

	return &response, nil
}
