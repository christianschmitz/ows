package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"ows/actions"
)

func listNodes(cmd *cobra.Command, args []string) {
	l := readLedger()
	m := actions.ListNodes(l)

	for id, addr := range m {
		fmt.Printf("%s %s\n", id, addr)
	}
}
