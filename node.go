package main

import (
    "fmt"
    "cws/ledger"
)

func main() {
    l := ledger.ReadLedger()

    fmt.Printf("Ledger OK, has %d changes \n", len(l.Changes))
}
