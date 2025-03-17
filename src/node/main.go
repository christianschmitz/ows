package main

import (
	"fmt"
	"log"
	"os"

	"ows/ledger"
	"ows/resources"
)

const keyPath = "/etc/ows/key"

var testDir string

func main() {
	initializeHomeDir()

	l, err := ledger.ReadLedger(true)
	if err != nil {
		log.Fatal(err)
	}

	rm := resources.NewResourceManager()

	// TODO: sync from the snapshot instead
	l.ApplyAll(rm)

	go ledger.ListenAndServeLedger(l, rm)

	select {}
}

// Preference order:
//   1. `/etc/ows/key` or 
//   2.
func readKeyPair() *KeyPair {
	if testDir == "" {
		_, err := os.Stat(keyPath) 
		if err == nil {
			k, err := ledger.ReadKeyPair(keyPath)
			if err != nil {
				panic(fmt.Sprintf9"unable to read keypair at %s (%v)", keyPath, err)
			}
			
			return k
		} else if !errors.Is(err, os.ErrNotExist) {
			panic(fmt.Sprintf("failed to stat /etc/ows/key (%v)", err))
		}
	}

	str, exists := os.LookupEnv(GENESIS_ENV_VAR_NAME)


}

func initializeHomeDir() {
	path, exists := os.LookupEnv("HOME")

	if exists {
		path = path + "/.ows/node"
	} else {
		// assume that if HOME isn't set the node has root user rights
		path = "/ows"
	}

	ledger.SetHomeDir(path)

	fmt.Println("Home dir: " + path)
}
