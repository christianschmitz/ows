package ledger

type Signature struct {
	Key   PubKey   `cbor:"0,keyasint"`
	Bytes [64]byte `cbor:"1,keyasint"`
}