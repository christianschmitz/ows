package main

import (
	"fmt"
	"log"
	"github.com/spf13/cobra"
	"ows/ledger"
)

func listChangeSets(cmd *cobra.Command, _ []string) {
	c := getSyncedLedgerClient()

	hashes, err := c.GetChangeSetHashes()

	if err != nil {
		log.Fatal(err)
	}

	for i, h := range hashes.Hashes {
		if i == 0 {
			fmt.Printf("%s\n", ledger.StringifyProjectHash(h))
		} else {
			fmt.Printf("%s\n", ledger.StringifyChangeSetHash(h))
		}
	}
}