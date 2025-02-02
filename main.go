package main

import (
	"context"
	"errors"
)

func main() {
	if err := run(context.Background()); err != nil {
		panic(errors.Join(errors.New("BTClient error"), err))
	}
}
