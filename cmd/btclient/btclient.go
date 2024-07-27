package main

import (
	"example.com/btclient/internal/btclient"
	"log"
)

func main() {
	if err := btclient.Run(); err != nil {
		log.Fatalf("Error running btclient: %v", err)
	}
}
