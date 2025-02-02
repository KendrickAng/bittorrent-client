package main

import (
	"context"
	"log"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatalf("BTClient error: %v", err)
	}
}
