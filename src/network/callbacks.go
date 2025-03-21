package network

import (
	"ows/ledger"
)

// Implemented by nodeState
type Callbacks interface {
	AddAsset(bs []byte, isFromNode bool) (ledger.AssetID, error)
	AppendChangeSet(cs *ledger.ChangeSet) error
	Ledger() *ledger.Ledger
	ListAssets() []ledger.AssetID
	Rollback(p int) error
	OwnKeyPair() *ledger.KeyPair
}
