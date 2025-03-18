package ledger

import (
	"fmt"
	"log"

	"golang.org/x/crypto/blake2b"
)

// Takes the ChangeSetID of the first change set, and changes to bech32 prefix
// to "project".
func (l *Ledger) ProjectID() ProjectID {
	bs, err := l.Changes[0].ID().encode()
	if err != nil {
		panic(fmt.Sprintf("invalid change set id (%v)", err))
	}

	return ProjectID(encodeBech32(ProjectIDPrefix, bs))
}

func (cs *ChangeSet) ID() ChangeSetID {
	bs := cs.Encode()
	hash := digestShort(bs)

	return decodeChangeSetID(hash)
}

func GenerateAssetID(bs []byte) AssetID {
	hash := digestShort(bs)

	return AssetID(encodeBech32(AssetIDPrefix, hash))
}

func (k PublicKey) UserID() UserID {
	return k.id(UserIDPrefix)
}

// Node ids are based directly on the node's public keys (instead of basing it
// on the prev ChangeSetID and AddNode action index).
func (k PublicKey) NodeID() NodeID {
	return k.id(NodeIDPrefix)
}

func (k PublicKey) id(prefix string) ResourceID {
	hash := digestShort(k)

	return ResourceID(encodeBech32(prefix, hash))
}

// Validate an id with a bech32 prefix
func ValidateID(id string, expectedPrefix string) error {
	prefix, _, err := decodeBech32(id)
	if err != nil {
		return fmt.Errorf("invalid id %s (%v)", id, err)
	}

	if prefix != expectedPrefix {
		return fmt.Errorf("invalid id %s, expected %s prefix, got %s prefix", id, expectedPrefix, prefix)
	}

	return nil
}

func newResourceIDGenerator(prev ChangeSetID, actionIndex uint) ResourceIDGenerator {
	return func(prefix string) ResourceID {
		return generateResourceId(prefix, prev, actionIndex)
	}
}

// Creates a resource id string by hashing a concatenation of the Prev
// ChangeSetID hash, and the little endian encoding of the action index.
//
// The current ChangeSetID can't be used because it isn't known yet.
func generateResourceId(prefix string, prev ChangeSetID, actionIndex uint) ResourceID {
	prevBytes, err := prev.encode()
	if err != nil {
		panic(fmt.Sprintf("invalid change set id %s format (%v)", prev, err))
	}

	indexBytes := encodeActionIndexLE(actionIndex)

	hash := digestShort(append(prevBytes, indexBytes...))

	id := encodeBech32(prefix, hash)

	return ResourceID(id)
}

// Custom integer little endian encoding function.
//
// Less cumbersome to use than the encoding functions in standard binary package
func encodeActionIndexLE(index uint) []byte {
	indexBytes := []byte{}

	// custom little endian encoding
	for index >= 256 {
		indexBytes = append(indexBytes, byte(index%256))
		index = index / 256
	}

	indexBytes = append(indexBytes, byte(index))

	return indexBytes
}

// Number of bytes returned by digestShort().
const shortDigestSize = 16

// Blake2b is faster than Sha3 and allows generating shorter digests, which are
// more readable and easier to use.
//
// Hash collision risk of using a short digest is low because each ledger is
// private and doesn't contain that many entries (unlike a public blockchain).
// A hash collision in OWS also wouldn't result in any financial risk.
func digestShort(bs []byte) []byte {
	hasher, err := blake2b.New(shortDigestSize, nil)
	if err != nil {
		log.Fatal(err)
	}

	hasher.Write(bs)
	hash := hasher.Sum(nil)

	return hash
}
