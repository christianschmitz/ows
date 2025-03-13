package ledger

import (
	"bytes"
	"strings"
)

// can be 0 bytes long (Genesis) or 32 bytes long
type ChangeSetHash = []byte

const CHANGE_SET_HASH_PREFIX = "changes"
const PROJECT_HASH_PREFIX = "project"

func StringifyChangeSetHash(h ChangeSetHash) string {
	return StringifyHumanReadableBytes(CHANGE_SET_HASH_PREFIX, h)
}

func StringifyProjectHash(h ChangeSetHash) string {
	return StringifyHumanReadableBytes(PROJECT_HASH_PREFIX, h)
}

func ParseChangeSetHash(h string) (ChangeSetHash, error) {
	if strings.HasPrefix(h, "/") {
		h = h[1:]
	}

	hash, err := ParseHumanReadableBytes(h, CHANGE_SET_HASH_PREFIX)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func ParseProjectHash(h string) (ChangeSetHash, error) {
	if strings.HasPrefix(h, "/") {
		h = h[1:]
	}

	hash, err := ParseHumanReadableBytes(h, PROJECT_HASH_PREFIX)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func IsSameChangeSetHash(a ChangeSetHash, b ChangeSetHash) bool {
	return bytes.Equal(a, b)
}
