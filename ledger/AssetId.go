package ledger

import (
	"errors"
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