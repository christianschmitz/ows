package ledger

import (
	"crypto/ed25519"
)

type Signature struct {
	Key   PubKey   `cbor:"0,keyasint"`
	Bytes [64]byte `cbor:"1,keyasint"`
}

func (s Signature) Verify(message []byte) bool {
	return ed25519.Verify(s.Key[:], message, s.Bytes[:])
}