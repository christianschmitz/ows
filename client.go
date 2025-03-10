package main

import (
    "fmt"
    "log"
    "os"
    "github.com/spf13/cobra"
    "ows/ledger"
)

func main() {
    initializeHomeDir()
    
    root := configureCLI()

    root.Execute()
}

func initializeHomeDir() {
    path, exists := os.LookupEnv("HOME")

    if (!exists) {
        log.Fatal("env variable HOME not set")
    }

    ledger.SetHomeDir(path + "/.ows/client")
}

func configureCLI() *cobra.Command {
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

            _ = getKeyPair()

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

            c.Ledger.Persist()
        },
    })

    root.AddCommand(tasks)

    return root
}

func getSyncedLedgerClient() *ledger.LedgerClient {
    l := ledger.ReadLedger()

    c := ledger.NewLedgerClient(l)
    c.Sync()

    return c
}

func readLedger() *ledger.Ledger {
    c := getSyncedLedgerClient()

    return c.Ledger
}

func getKeyPair() *ledger.KeyPair {
    p, err := ledger.ReadKeyPair(ledger.HomeDir + "/key", true)

    if err != nil {
        log.Fatal(err)
    }

    return p
}