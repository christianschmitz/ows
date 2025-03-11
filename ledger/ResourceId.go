package ledger

import (
	"errors"
)

type ResourceId = [32]byte

type ResourceIdGenerator = func () ResourceId

const RESOURCE_ID_PREFIX = "resource"

func GenerateGlobalResourceId() ResourceId {
	globalResourceId := [32]byte{}

	for i, _ := range globalResourceId {
		globalResourceId[i] = 0
	}

	return globalResourceId
}

func StringifyResourceId(id ResourceId) string {
	return StringifyHumanReadableBytes(RESOURCE_ID_PREFIX, id[:])
}

func ParseResourceId(str string) (ResourceId, error) {
	bs, err := ParseHumanReadableBytes(str, RESOURCE_ID_PREFIX)
	if err != nil {
		return [32]byte{}, err
	}

	if len(bs) != 32 {
		return [32]byte{}, errors.New("unexpected number of resourec id bytes")
	}

	rId := [32]byte{}

	for i, b := range rId {
		rId[i] = b
	}

	return rId, nil
}