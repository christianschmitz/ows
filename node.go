package main

import (
    "fmt"
    "cws/ledger"
    "cws/resources"
)

func main() {
    l := ledger.ReadLedger(false)

    rm := resources.NewResourceManager()

    l.ApplyAll(rm)

    fmt.Printf("Ledger OK, has %d changes \n", len(l.Changes))

    go ledger.ListenAndServeLedger(l, rm)

    select {}
}