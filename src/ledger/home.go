package ledger

import (
	"log"
	"os"
)

var HomeDir string = ""

func SetHomeDir(path string) {
	HomeDir = path

	if path == "" {
		log.Fatal("can't set HomeDir to empty string")
	}

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}
