package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"ows/actions"
)

func listGateways(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		log.Fatal("unexpected args")
	}

	c := getSyncedLedgerClient()
	gateways := actions.ListGateways(c.Ledger)

	for id, config := range gateways {
		fmt.Printf("%s %d\n", id, config.Port)
	}
}
