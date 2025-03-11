package ledger

import (
	"errors"
	"log"
	"os"
	"strconv"
	"golang.org/x/crypto/sha3"
)

type AssetId = [32]byte

const ASSET_ID_PREFIX = "asset"

func GenerateAssetId(content []byte) AssetId {
	return sha3.Sum256(content)
}

func StringifyAssetId(id AssetId) string {
	return StringifyHumanReadableBytes(ASSET_ID_PREFIX, id[:])
}

func ParseAssetId(h string) (AssetId, error) {
	rawId, err := ParseHumanReadableBytes(h, ASSET_ID_PREFIX)
	if err != nil {
		return [32]byte{}, err
	}

	if len(rawId) != 32 {
		return [32]byte{}, errors.New("asset id not exactly 32 bytes long, got " + strconv.Itoa(len(rawId)) + " bytes")
	}

	id := [32]byte{}

	for i, b := range rawId {
		id[i] = b
	}

	return id, nil
}

func AssetExists(id AssetId) bool {
	path := HomeDir + "/assets/" + StringifyAssetId(id)

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