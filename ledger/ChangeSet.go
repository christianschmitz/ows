package ledger

import (
	"log"
	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/sha3"
)

type ChangeSet struct {
	Parent     ChangeSetHash
	Signatures []Signature
	Actions    []Action
}

type ChangeSetCbor struct {
	Parent     ChangeSetHash `cbor:"0,keyasint"`
	Signatures []Signature   `cbor:"1,keyasint"`
	Actions    []ActionCbor  `cbor:"2,keyasint"`
}

func GenerateResourceId(parentId []byte, i int) ResourceId {
	return sha3.Sum256(append(parentId, byte(i)))
}

func DecodeChangeSet(bytes []byte) (*ChangeSet, error) {
	v := ChangeSetCbor{}

	err := cbor.Unmarshal(bytes, &v)

	if err != nil {
		return nil, err
	}

	return v.convertToChangeSet()
}

func (c *ChangeSet) Apply(m ResourceManager) {
	for i, a := range c.Actions {
		a.Apply(m, func () ResourceId {
			return GenerateResourceId(c.Parent[:], i)
		})
	}
}

func (c *ChangeSet) Encode() []byte {
	bytes, err := cbor.Marshal(c.convertToChangeSetCbor())

	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func (c *ChangeSet) Hash() ChangeSetHash {
	return sha3.Sum256(c.Encode())
}

func (c *ChangeSet) convertToChangeSetCbor() ChangeSetCbor {
	actions := make([]ActionCbor, len(c.Actions))

	for i, a := range c.Actions {
		h := NewActionHelper(a)
		actions[i] = h.convertToActionCbor()
	}

	return ChangeSetCbor{c.Parent, c.Signatures, actions}
}

func (c ChangeSetCbor) convertToChangeSet() (*ChangeSet, error) {
	actions := make([]Action, len(c.Actions))

	for i, a := range c.Actions {
		var err error
		actions[i], err = a.convertToAction()

		if err != nil {
			return nil, err
		}
	}

	return &ChangeSet{
		c.Parent,
		c.Signatures,
		actions,
	}, nil
}