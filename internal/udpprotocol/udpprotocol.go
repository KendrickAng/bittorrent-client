package udpprotocol

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/exp/rand"
	"net"
	"time"
)

const (
	connectProtocolId int64 = 0x41727101980
	connectAction     int32 = 0
)

func Connect(conn *net.Conn) error {
	// Choose a random 32-bit int random transaction ID.
	tId := randInt32()

	// Fill the connect request structure.
	connectPacket, err := buildConnectPacket(tId)
	if err != nil {
		return err
	}

	tracker := *conn

	// Read the packet pre-emptively.
	go func() {
		// TODO
		buf := make([]byte, 1024)
		n, err := tracker.Read(buf)
		if err != nil {
			fmt.Printf("Error reading from tracker: %v\n", err)
		}
		fmt.Printf("Read %d bytes from %s\n", n, tracker.RemoteAddr().String())
		fmt.Println(string(buf))
	}()

	// Send the packet.
	fmt.Printf("Sending connection packet (no-op for now): %+v\n", connectPacket)
	n, err := tracker.Write(connectPacket)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes to %s\n", n, tracker.RemoteAddr().String())

	time.Sleep(time.Minute * 1)

	return nil
}

func randInt32() int32 {
	return rand.New(rand.NewSource(uint64(time.Now().UnixNano()))).Int31()
}

func buildConnectPacket(transactionId int32) ([]byte, error) {
	return buildPacket(16, map[int]any{
		0:  connectProtocolId,
		8:  connectAction,
		12: transactionId,
	})
}

func buildPacket(byteSize int, offsetToVal map[int]any) ([]byte, error) {
	buf := make([]byte, byteSize)

	for offset, val := range offsetToVal {
		switch t := val.(type) {
		case int64:
			binary.BigEndian.PutUint64(buf[offset:offset+8], uint64(t))
		case int32:
			binary.BigEndian.PutUint32(buf[offset:offset+4], uint32(t))
		case string:
			s := val.(string)
			buf, err := writeString(buf[offset:offset+len(s)], s)
			if err != nil {
				return buf, err
			}
		default:
			panic(fmt.Errorf("unsupported type: %v", t))
		}
	}

	return buf, nil
}

func writeString(arr []byte, s string) ([]byte, error) {
	if len(arr) < len(s) {
		return nil, fmt.Errorf("attemping to copy string of length %d to buffer of length %d", len(s), len(arr))
	}
	buffer := make([]byte, len(arr))
	copy(buffer, []byte(s))
	return buffer, nil
}
