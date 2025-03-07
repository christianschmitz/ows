package main

import (
    "fmt"
    "log"
    "github.com/spf13/cobra"
    "cws/ledger"
)

func main() {
    cmd := &cobra.Command{
		Use:   "cws",
		Short: "CWS Client CLI",
		Run: func(cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                log.Fatal("expected 1 arg")    
            }

            g := ledger.NewGenesisChangeSet(
                ledger.NewAddCompute(
                    args[0],
                ),
            )

            fmt.Println(g.EncodeBase64())
		},
	}

    cmd.Execute()
}

