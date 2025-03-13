package main

import (
	"fmt"
	"log"
	"os"
	"ows/actions"
	"ows/ledger"
	"ows/resources"
)

var _ActionsInitialized = actions.InitializeActions()

func main() {
	initializeHomeDir()

	l, err := ledger.ReadLedger(true)
	if err != nil {
		log.Fatal(err)
	}

	rm := resources.NewResourceManager()

	l.ApplyAll(rm)

	go ledger.ListenAndServeLedger(l, rm)

	select {}
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
