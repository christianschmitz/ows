package main

import (
	"errors"
	"log"
	"os"
)

func GetAssetsDir() string {
	assetsDir := HomeDir + "/assets"

	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	return assetsDir
}

func AssetExists(id string) bool {
	assetsDir := GetAssetsDir()

	path := assetsDir + "/" + id

	_, err := os.Stat(path)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		} else {
			log.Println(err)
			return false
		}
	} else {
		// TODO: should we check if file is corrupt or not?
		return true
	}
}
