package main

import (
	"errors"
	"example.com/btclient/internal/btclient"
)

func main() {
	if err := btclient.Run(); err != nil {
		panic(errors.Join(errors.New("BTClient error"), err))
	}
}
