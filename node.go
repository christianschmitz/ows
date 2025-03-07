package main

import (
    "fmt"
    "cws/ledger"
)

func main() {
    _ = ledger.GetEnvGenesisChangeSet()

    fmt.Println("Genesis change-set OK")
}
