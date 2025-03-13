package ledger

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"github.com/fxamacker/cbor/v2"
)

type Ledger struct {
	Changes  []ChangeSet // the first entry is the genesis change set
	Head     ChangeSetHash
	Snapshot *ValidationContext
}

// TODO: should validation be moved from ReadLedger() to DecodeLedger()?
func DecodeLedger(bytes []byte) (*Ledger, error) {
	changes, err := decodeLedgerChanges(bytes)
	if err != nil {
		return nil, err
	}

	var head ChangeSetHash

	l := &Ledger{changes, head, nil}

	l.syncHead()

	return l, nil
}

func (l *Ledger) Encode() []byte {
	return encodeLedgerChanges(l.Changes)
}

func (l *Ledger) syncHead() {
	l.Head = l.Changes[len(l.Changes)-1].Hash()
}

// creates an unsigned change set
func (l *Ledger) NewChangeSet(actions ...Action) *ChangeSet {
	cs := &ChangeSet{
		Parent: l.Head,
		Actions: actions,
		Signatures: []Signature{},
	}

	return cs
}

func makeLedgerPath(projectPath string) string {
	return projectPath + "/ledger"
}

func getLedgerPath(genesis *ChangeSet) string {
	h := StringifyProjectHash(genesis.Hash())	

	projectPath := HomeDir + "/" + h

	if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	return makeLedgerPath(projectPath)
}

func ReadLedger(validateAssets bool) (*Ledger, error) {
	g, err := LookupGenesisChangeSet()
	if err != nil {
		return nil, err
	}

	gHash := g.Hash()

	ledgerPath := getLedgerPath(g)

	l, ok := readLedger(ledgerPath)
	if !ok {
		l = &Ledger{[]ChangeSet{*g}, gHash, nil}
		if err := l.ValidateAll(validateAssets); err != nil {
			return nil, err
		}

		l.Write()
	} else {
		if err := l.ValidateAll(validateAssets); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	l.syncHead()

	return l, nil
}

func readLedger(ledgerPath string) (*Ledger, bool) {
	if _, err := os.Stat(ledgerPath); err == nil {
		dat, err := os.ReadFile(ledgerPath)

		if err == nil {
			l, err := DecodeLedger(dat)	
			if err == nil {
				return l, true
			} else {
				fmt.Println("Failed to decode ledger")
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}

	return nil, false
}

func (l *Ledger) ApplyAll(m ResourceManager) {
	for _, c := range l.Changes {
		c.Apply(m)
	}
}

func (l *Ledger) KeepChangeSets(until int) {
	l.Changes = l.Changes[0:until+1]

	l.syncHead()
}

func (l *Ledger) AppendChangeSet(cs *ChangeSet, validateAssets bool) error {
	backup := l.Changes

	l.Changes = append(backup, *cs)

	if err := l.ValidateAll(validateAssets); err != nil {
		l.Changes = backup
		return err
	}

	l.syncHead()

	return nil
}

func (l *Ledger) GetChangeSet(h ChangeSetHash) (*ChangeSet, bool) {
	// start from end
	for i := len(l.Changes); i >= 0; i-- {
		if i > 0 && i == len(l.Changes) {
			if IsSameChangeSetHash(h, l.Head) {
				return &(l.Changes[i-1]), true
			}
		} else if i < len(l.Changes) {
			if i > 1 && IsSameChangeSetHash(h, l.Changes[i].Parent) {
				return &(l.Changes[i-1]), true
			}
		} else if i > 0 {
			if (IsSameChangeSetHash(h, l.Changes[i].Hash())) {
				return &(l.Changes[i]), true
			}
		}
	}

	return nil, false
}

func (l *Ledger) GetChangeSetHashes() *ChangeSetHashes {
	hashes := make([]ChangeSetHash, len(l.Changes))

	for i := 0; i < len(l.Changes); i++ {
		if i+1 == len(l.Changes) {
			hashes[i] = l.Head
		} else if i+1 < len(l.Changes) {
			hashes[i] = l.Changes[i+1].Parent
		} else {
			hashes[i] = l.Changes[i].Hash()
		}
	}

	return &ChangeSetHashes{
		hashes,
	}
}

func (l *Ledger) ValidateAll(validateAssets bool) error {
	rootSignatures := l.Changes[0].Signatures[:]

	genesisBytes, err := l.Changes[0].Encode(true)
	if err != nil {
		return err
	}

	for _, s := range rootSignatures {
		if !s.Verify(genesisBytes) {
			return errors.New("invalid root signature")
		}
	}

	rootUsers := l.Changes[0].CollectSigners()

	// create validation context
	context := newValidationContext(validateAssets, rootUsers)

	// replay all the changes

	head := []byte{}

	for i, c := range l.Changes {
		// check that the Parent corresponds
		if !bytes.Equal(c.Parent, head) {
			return errors.New("Invalid change set head, expected " + StringifyChangeSetHash(head) + ", got " + StringifyChangeSetHash(c.Parent))
		}

		// first validate that the signatures correspond
		signers := []PubKey{}

		if i == 0 {
			signers = rootUsers
		} else {
			for _, s := range c.Signatures {
				cbs, err := c.Encode(true)
				if err != nil {
					return err
				}

				if !s.Verify(cbs) {
					return errors.New("invalid change set signatures")
				}
			}

			signers = c.CollectSigners()
		}

		// check that all the actions can actually be taken by the signers
		userPolicies, err := context.getSignatoryPermissions(signers)
		if err != nil {
			return err
		}

		for _, a := range c.Actions {
			allowed := false
			for _, policy := range userPolicies {
				if policy.AllowsAll(a.GetResources(), a.GetCategory(), a.GetName()) {
					allowed = true
				}
			}

			if !allowed {
				return errors.New("merged policy of all signers doesn't allow " + a.GetCategory() + ":" + a.GetName())
			}
		}

		if err := c.Apply(context); err != nil {
			return err
		}

		head = c.Hash()
	}

	return nil
}

func (l *Ledger) Write() {
	ledgerPath := getLedgerPath(&(l.Changes[0]))

	bytes := l.Encode()

	err := os.WriteFile(ledgerPath, bytes, 0644)

	if err != nil {
		log.Fatal(err)
	}
}

func decodeLedgerChanges(bytes []byte) ([]ChangeSet, error) {
	v := []ChangeSetCbor{}

	err := cbor.Unmarshal(bytes, &v)

	if err != nil {
		return nil, err
	}

	changes := make([]ChangeSet, len(v))

	for i, c := range v {
		cc, err := c.convertToChangeSet()

		if err != nil {
			return nil, err
		}

		changes[i] = *cc
	}

	return changes, nil
}

func encodeLedgerChanges(changes []ChangeSet) []byte {
	v := make([]ChangeSetCbor, len(changes))

	for i, c := range changes {
		v[i] = c.convertToChangeSetCbor(false)
	}

	bs, err := cbor.Marshal(v)

	if err != nil {
		log.Fatal(err)
	}

	return bs
}