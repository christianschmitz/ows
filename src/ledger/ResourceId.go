package ledger

import (
	"errors"
	"log"
)

type ResourceIdGenerator = func (prefix string) string

func GenerateGlobalResourceId() string {
	return "*"
}

func GenerateResourceId(prefix string, parentId []byte, index int) string {
	if (index < 0) {
		log.Fatal("invalid negative index in GenerateResourceId")
	}

	indexBytes := []byte{}

	// custom little endian encoding
	for index >= 256 {
		indexBytes = append(indexBytes, byte(index%256))
		index = index/256
	}

	indexBytes = append(indexBytes, byte(index))

	resourceIdBytes := DigestCompact(append(parentId, indexBytes...))

	return StringifyHumanReadableBytes(prefix, resourceIdBytes[:])
}

func ValidateResourceId(id string, expectedPrefix string) error {
	bs, err := ParseHumanReadableBytes(id, expectedPrefix)
	if err != nil {
		return err
	}

	if StringifyHumanReadableBytes(expectedPrefix, bs) != id {
		return errors.New("bech32 roundtrip error")
	}

	return nil
}