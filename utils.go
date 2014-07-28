package main

import (
	"os"
)

func fileIsReadable(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
