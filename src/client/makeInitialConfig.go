package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"ows/actions"
	"ows/ledger"
)

func makeInitialConfig(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		log.Fatal("expected 1 arg")
	}

	keyPair := getKeyPair()

	g := ledger.NewGenesisChangeSet(
		actions.NewAddNode(
			args[0],
		),
	)

	// TODO: multi-sig
	signature, err := keyPair.SignChangeSet(g)
	if err != nil {
		log.Fatal(err)
	}

	g.Signatures = append(g.Signatures, signature)

	fmt.Println(g.EncodeToString())
}
