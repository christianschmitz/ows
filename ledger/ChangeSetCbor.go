package ledger

import (
)

// also used by GenesisSet
type ChangeSetCbor struct {
	Parent     []byte       `cbor:"0,keyasint"`
	Actions    []ActionCbor `cbor:"1,keyasint"`
	Signatures []Signature  `cbor:"2,keyasint,omitempty"`
}

func (c ChangeSetCbor) convertToChangeSet() (*ChangeSet, error) {
	actions := make([]Action, len(c.Actions))

	for i, a := range c.Actions {
		var err error
		ac, err := a.convertToAction()
		if err != nil {
			return nil, err
		}

		actions[i] = ac
	}

	return &ChangeSet{
		c.Parent,
		actions,
		c.Signatures,
	}, nil
}