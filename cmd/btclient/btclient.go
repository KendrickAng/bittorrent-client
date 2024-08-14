package main

import (
	"context"
	"errors"
	"example.com/btclient/internal/btclient"
)

func main() {
	if err := btclient.Run(context.Background()); err != nil {
		panic(errors.Join(errors.New("BTClient error"), err))
	}
}
