package ledger

import (
	"bytes"
	"log"
	"strings"
)

type ChangeSetHash = [32]byte

func StringifyChangeSetHash(h ChangeSetHash) string {
	return StringifyBytes(h[:])
}

func ParseChangeSetHash(h string) ChangeSetHash {
	if (strings.HasPrefix(h, "/")) {
		h = h[1:]
	}
	
	bytes, err := ParseBytes(h)

	if err != nil {
		log.Fatal(err)
	}

	if len(bytes) != 32 {
		log.Fatal("hash not exactly 32 bytes long")
	}

	hash := [32]byte{}

	for i, b := range bytes {
		hash[i] = b
	}

	return hash
}

func IsSameChangeSetHash(a ChangeSetHash, b ChangeSetHash) bool {
	return bytes.Equal(a[:], b[:])
}