package ledger

import (
	"encoding/base64"
	"errors"
	"log"
	"os"
	"golang.org/x/crypto/sha3"
	"github.com/fxamacker/cbor/v2"
)

type Ledger struct {
	changes []ChangeSet
}

type ChangeSetHash = [32]byte

type ChangeSet interface {
	GetEntries() []ChangeEntry
	Hash() ChangeSetHash
}

type GenesisChangeSet struct {
	changes []ChangeEntry
}

type RegularChangeSet struct {
	prev ChangeSetHash
	signatures []Signature
	changes []ChangeEntry
}

type Signature struct {
	key PubKey
	bytes [64]byte
}

type PubKey = [32]byte

type ChangeEntry interface {
	// valid categories are: compute, permissions
	GetCategoryName() string
	GetActionName() string
}

type cborChangeEntry struct {
	Category   string `cbor:"0,keyasint"`
	Action     string `cbor:"1,keyasint"`
	Attributes []byte `cbor:"2,keyasint"`
}

type AddCompute struct {
	addr string `cbor:"0,keyasint"`
}

type RemoveCompute struct {
	addr string `cbor:"0,keyasint"`
}

type AddUser struct {
	// public Ed25519 32 byte key
	// TODO: like to policy
	key PubKey `cbor:"0,keyasint"`
}

func NewGenesisChangeSet(entries ...ChangeEntry) *GenesisChangeSet {
	return &GenesisChangeSet{entries}
}

// is encoded as list of bytes
func (g *GenesisChangeSet) Encode() []byte {
	n := len(g.changes)
	lst := make([][]byte, n)

	for i, c := range(g.changes) {
		lst[i] = EncodeChangeEntry(c)
	}

	bytes, err := cbor.Marshal(lst)

	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func (g *GenesisChangeSet) EncodeBase64() string {
	bytes := g.Encode()
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)
}

func (g *GenesisChangeSet) GetEntries() []ChangeEntry {
	return g.changes
}

func (g *GenesisChangeSet) Hash() ChangeSetHash {
	return sha3.Sum256(g.Encode())
}

func DecodeGenesisChangeSet(bytes []byte) (*GenesisChangeSet, error) {
	lst := [][]byte{}
	err := cbor.Unmarshal(bytes, &lst)

	if err != nil {
		return nil, err
	}

	n := len(lst)
	changes := make([]ChangeEntry, n)

	for i := 0; i < n; i++ {
		c, err := DecodeChangeEntry(lst[i])

		if err != nil {
			return nil, err
		}

		changes[i] = c
	}

	g := GenesisChangeSet{changes}

	return &g, nil
}

func GetEnvGenesisChangeSet() *GenesisChangeSet {
	str, exists := os.LookupEnv("CWS_GENESIS")

	// Check if the variable exists
	if !exists {
		log.Fatal("CWS_GENESIS is not set")
	}

	decodedBytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(str)
	if err != nil {
		log.Fatal(err)
	}

	g, err := DecodeGenesisChangeSet(decodedBytes)

	if err != nil {
		log.Fatal(err)
	}

	return g
}

func EncodeChangeEntry(c ChangeEntry) []byte {
	attrBytes, err := cbor.Marshal(c)

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := cbor.Marshal(cborChangeEntry{
		c.GetCategoryName(),
		c.GetActionName(),
		attrBytes,
	})

	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func DecodeChangeEntry(bytes []byte) (ChangeEntry, error) {
	var v cborChangeEntry

	err := cbor.Unmarshal(bytes, &v)

	if err != nil {
		return nil, err
	}

	category := v.Category
	action := v.Action
	attrBytes := v.Attributes

	switch category {
	case "compute":
		switch action {
		case "AddCompute":
			var c AddCompute
			err = cbor.Unmarshal(attrBytes, &c)
			return &c, err
		case "RemoveCompute":
			var c RemoveCompute
			err = cbor.Unmarshal(attrBytes, &c)
			return &c, err
		default:
			return nil, errors.New("invalid " + category + " action " + action)
		}
	case "permissions":
		switch action {
		case "AddUser":
			var c AddUser
			err = cbor.Unmarshal(attrBytes, &c)
			return &c, err
		default:
			return nil, errors.New("invalid " + category + " action " + action)
		}
	default:
		return nil, errors.New("invalid category " + category)
	}
}

func NewAddCompute(addr string) *AddCompute {
	return &AddCompute{addr}
}

func (c *AddCompute) GetCategoryName() string {
	return "compute"
}

func (c *AddCompute) GetActionName() string {
	return "AddCompute"
}

func (c *RemoveCompute) GetCategoryName() string {
	return "compute"
}

func (c *RemoveCompute) GetActionName() string {
	return "RemoveCompute"
}

func (c *AddUser) GetCategoryName() string {
	return "permissions"
}

func (c *AddUser) GetActionName() string {
	return "AddUser"
}
