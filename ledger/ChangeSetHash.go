package ledger

import "encoding/hex"

type ChangeSetHash = [32]byte

func StringifyHash(h ChangeSetHash) string {
	return hex.EncodeToString(h[:])
}