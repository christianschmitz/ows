package ledger

import (
	"errors"
	"log"
	"os"
	"github.com/fxamacker/cbor/v2"
)

type Ledger struct {
	Genesis GenesisSet
	Changes []ChangeSet
	Head ChangeSetHash
}

// TODO: validate
func DecodeLedger(g *GenesisSet, bytes []byte) (*Ledger, error) {
	changes, err := decodeLedgerChanges(bytes)

	if err != nil {
		return nil, err
	}

	var head ChangeSetHash

	l := &Ledger{*g, changes, head}

	l.syncHead()

	return l, nil
}

func (l *Ledger) Encode() []byte {
	return encodeLedgerChanges(l.Changes)
}

func (l *Ledger) syncHead() {
	if (len(l.Changes) > 0) {
		l.Head = l.Changes[len(l.Changes)-1].Hash()
	} else {
		l.Head = l.Genesis.Hash()
	}
}

func getHomePath(isClient bool) string {
    path, exists := os.LookupEnv("HOME")

    if exists {
		if (isClient) {
			path = path + "/.ows/client"
		} else {
			path = path + "/.ows/node"
		}
    } else {
		if (isClient) {
			log.Fatal("no home path set for client")
		} else {
        	path = "/ows"
		}
    }

    err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
        log.Fatal(err)
	}

    return path
}

func getLedgerPath(g *GenesisSet, isClient bool) string {
	h := StringifyChangeSetHash(g.Hash())

	root := getHomePath(isClient)

	projectPath := root + "/" + h

	if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	ledgerPath := projectPath + "/ledger"

	return ledgerPath
}

func ReadLedger(isClient bool) *Ledger {
	g := LookupGenesisSet()

	ledgerPath := getLedgerPath(g, isClient)

	if _, err := os.Stat(ledgerPath); err == nil {
		dat, err := os.ReadFile(ledgerPath)

		if err == nil {
			l, err := DecodeLedger(g, dat)	

			if err == nil {
				return l
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Println(err)
	}

	return &Ledger{*g, []ChangeSet{}, g.Hash()}
}

func (l *Ledger) ApplyAll(m ResourceManager) {
	l.Genesis.Apply(m)

	for _, c := range l.Changes {
		c.Apply(m)
	}
}

func (l *Ledger) KeepChangeSets(until int) {
	l.Changes = l.Changes[0:until]

	l.syncHead()
}

// TODO: validate
func (l *Ledger) AppendChangeSet(cs *ChangeSet) {
	l.Changes = append(l.Changes, *cs)

	l.syncHead()
}

func (l *Ledger) GetNodeAddresses() map[ResourceId]string {
	m := map[ResourceId]string{}

	for i, a := range l.Genesis.Actions {
		if ac, ok := a.(*AddCompute); ok {
			id := GenerateGenesisResourceId(i)
			m[id] = ac.Address
		} else if rc, ok := a.(*RemoveCompute); ok {
			delete(m, rc.Id)
		}
	}

	for _, cs := range l.Changes {
		for i, a := range cs.Actions {
			if ac, ok := a.(*AddCompute); ok {
				id := GenerateResourceId(cs.Parent[:], i)
				m[id] = ac.Address
			} else if rc, ok := a.(*RemoveCompute); ok {
				delete(m, rc.Id)
			}
		}
	}

	return m
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
	hashes := make([]ChangeSetHash, 1 + len(l.Changes))

	for i := 0; i < 1 + len(l.Changes); i++ {
		if i < len(l.Changes) {
			hashes[i] = l.Changes[i].Parent
		} else if i == 0 {
			hashes[i] = l.Genesis.Hash()
		} else {
			hashes[i] = l.Changes[i-1].Hash()
		}
	}

	return &ChangeSetHashes{
		hashes,
	}
}

func (l *Ledger) Persist(isClient bool) {
	ledgerPath := getLedgerPath(&(l.Genesis), isClient)

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
		v[i] = c.convertToChangeSetCbor()
	}

	bs, err := cbor.Marshal(v)

	if err != nil {
		log.Fatal(err)
	}

	return bs
}