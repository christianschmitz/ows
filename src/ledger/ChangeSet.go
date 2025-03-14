package ledger

import (
	"bytes"
	"github.com/fxamacker/cbor/v2"
	"log"
)

type ChangeSet struct {
	Parent     ChangeSetHash
	Actions    []Action
	Signatures []Signature
}

func DecodeChangeSet(bytes []byte) (*ChangeSet, error) {
	v := ChangeSetCbor{}

	err := cbor.Unmarshal(bytes, &v)
	if err != nil {
		return nil, err
	}

	return v.convertToChangeSet()
}

func (c *ChangeSet) Apply(m ResourceManager) error {
	for i, a := range c.Actions {
		err := a.Apply(m, func(prefix string) string {
			return GenerateResourceId(prefix, c.Parent, i)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *ChangeSet) Encode(forSigning bool) ([]byte, error) {
	bytes, err := cbor.Marshal(c.convertToChangeSetCbor(forSigning))
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (g *ChangeSet) EncodeToString() string {
	bytes, err := g.Encode(false)
	if err != nil {
		log.Fatal(err)
	}

	return StringifyCompactBytes(bytes)
}

func (c *ChangeSet) Hash() ChangeSetHash {
	bytes, err := c.Encode(false)
	if err != nil {
		log.Fatal(err)
	}

	hash := DigestCompact(bytes)

	return hash[:]
}

func (c *ChangeSet) convertToChangeSetCbor(forSigning bool) ChangeSetCbor {
	actions := make([]ActionCbor, len(c.Actions))

	for i, a := range c.Actions {
		h := NewActionHelper(a)
		actions[i] = h.convertToActionCbor()
	}

	signatures := c.Signatures[:]

	if forSigning {
		signatures = []Signature{}
	}

	return ChangeSetCbor{c.Parent[:], actions, signatures}
}

func (c *ChangeSet) CollectSigners() []PubKey {
	signatures := c.Signatures[:]
	users := []PubKey{}

	for _, s := range signatures {
		unique := true

		for _, pk := range users {
			if bytes.Equal(pk[:], s.Key[:]) {
				unique = false
				break
			}
		}

		if unique {
			users = append(users, s.Key)
		}
	}

	return users
}
