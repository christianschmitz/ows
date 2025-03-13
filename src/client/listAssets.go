package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

func listAssets(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		log.Fatal("expected 0 args")
	}

	c := getSyncedLedgerClient()

	assets, err := c.GetAssets()
	if err != nil {
		log.Fatal(err)
	}

	for _, a := range assets {
		fmt.Println(a)
	}
}
