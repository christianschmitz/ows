package ledger

import (
	"errors"
	"log"
	"os"
)

const ASSET_ID_PREFIX = "asset"

func GenerateAssetId(content []byte) string {
	idBytes := DigestCompact(content)

	return stringifyAssetId(idBytes)
}

func stringifyAssetId(idBytes []byte) string {
	return StringifyHumanReadableBytes(ASSET_ID_PREFIX, idBytes)
}

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