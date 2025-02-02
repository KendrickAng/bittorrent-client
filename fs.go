package main

import (
	"embed"
)

//go:embed data/*
var fs embed.FS

const (
	dataDir = "data"
)

func readData(filename string) ([]byte, error) {
	return fs.ReadFile(dataDir + "/" + filename)
}
