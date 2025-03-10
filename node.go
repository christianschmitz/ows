package main

import (
    "fmt"
    "os"
    "ows/ledger"
    "ows/resources"
)

func main() {
    initializeHomeDir()

    l := ledger.ReadLedger()

    rm := resources.NewResourceManager()

    l.ApplyAll(rm)

    fmt.Printf("Ledger OK, has %d changes \n", len(l.Changes))

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
}