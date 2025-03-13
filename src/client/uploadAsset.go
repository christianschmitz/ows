package main

import (
	"fmt"
	"log"
	"os"
	"github.com/spf13/cobra"
)

func uploadAssets(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		log.Fatal("expected at least 1 arg")
	}

	c := getSyncedLedgerClient()

	for _, arg := range args {
		bs, err := os.ReadFile(arg)
		if err != nil {
			log.Fatal(err)
		}

		id, err := c.UploadFile(bs)  
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s: %s\n", arg, id)
	}
}