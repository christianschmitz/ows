package main

import (
    "fmt"
    "log"
    "github.com/spf13/cobra"
    "ows/ledger"
)

func main() {
    root := &cobra.Command{
		Use:   "ows",
		Short: "Open Web Services CLI",
	}

    root.AddCommand(&cobra.Command{
        Use: "init",
        Short: "Create genesis config",
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                log.Fatal("expected 1 arg")    
            }

            g := ledger.NewGenesisSet(
                ledger.NewAddCompute(
                    args[0],
                ),
            )

            fmt.Println(g.EncodeBase64())
		},
    })

    root.AddCommand(&cobra.Command{
        Use: "nodes",
        Short: "List node addresses",
        Run: func(cmd *cobra.Command, _ []string) {
            l := readLedger()
            m := l.GetNodeAddresses()

            for id, addr := range m {
                fmt.Printf("%s %s\n", ledger.StringifyResourceId(id), addr)
            }
        },
    })

    root.AddCommand(&cobra.Command{
        Use: "hashes",
        Short: "List config change hashes (including genesis)",
        Run: func(cmd *cobra.Command, _ []string) {
            c := getSyncedLedgerClient()

            hashes := c.GetChangeSetHashes()

            for _, h := range hashes.Hashes {
                fmt.Printf("%s\n", ledger.StringifyChangeSetHash(h))
            }
        },
    })

    tasks := &cobra.Command{
        Use: "tasks",
        Short: "Manage tasks",
    }
    tasks.AddCommand(&cobra.Command{
        Use: "add",
        Short: "Create a new task",
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                log.Fatal("expected 1 arg")
            }

            c := getSyncedLedgerClient()

            cs := &ledger.ChangeSet{
                Parent: c.Ledger.Head,
                Signatures: []ledger.Signature{}, // TODO: sign
                Actions: []ledger.Action{
                    &ledger.AddTask{"nodejs", args[0]},
                },
            }

            c.PublishChangeSet(cs)

            c.Ledger.AppendChangeSet(cs)

            c.Ledger.Persist(true)
        },
    })
    root.AddCommand(tasks)

    root.Execute()
}

func getSyncedLedgerClient() *ledger.LedgerClient {
    l := ledger.ReadLedger(true)

    c := ledger.NewLedgerClient(l)
    c.Sync(true)

    return c
}

func readLedger() *ledger.Ledger {
    c := getSyncedLedgerClient()

    return c.Ledger
}
