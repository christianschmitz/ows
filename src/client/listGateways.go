package main

import (
	"fmt"
	"log"
	"github.com/spf13/cobra"
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