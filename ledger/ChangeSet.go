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

func (c ChangeSetCbor) convertToChangeSet() ChangeSet {
	actions := make([]Action, len(c.Actions))

	for i, a := range c.Actions {
		var err error
		actions[i], err = a.convertToAction()

		if err != nil {
			log.Fatal(err)
		}
	}

	return ChangeSet{
		c.Parent,
		c.Signatures,
		actions,
	}
}