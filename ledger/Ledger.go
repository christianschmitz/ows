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
}

// TODO: validate
func DecodeLedger(g *GenesisSet, bytes []byte) *Ledger {
	return &Ledger{*g, decodeLedgerChanges(bytes)}
}

func GetHomePath() string {
    path, exists := os.LookupEnv("HOME")

    if exists {
        path = path + "/.cws"
    } else {
        path = "/cws"
    }

    err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
        log.Fatal(err)
	}

    return path
}

func ReadLedger() *Ledger {
	g := LookupGenesisSet()

	h := StringifyHash(g.Hash())

	root := GetHomePath()

	projectPath := root + "/" + h

	if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	ledgerPath := projectPath + "/ledger"
	if _, err := os.Stat(ledgerPath); err == nil {
		dat, err := os.ReadFile(ledgerPath)

		if err == nil {
			return DecodeLedger(g, dat)	
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		log.Println(err)
	}

	return &Ledger{*g, []ChangeSet{}}
}

func decodeLedgerChanges(bytes []byte) []ChangeSet {
	v := []ChangeSetCbor{}

	err := cbor.Unmarshal(bytes, &v)

	if err != nil {
		log.Fatal(err)
	}

	changes := make([]ChangeSet, len(v))

	for i, c := range v {
		changes[i] = c.convertToChangeSet()
	}

	return changes
}
