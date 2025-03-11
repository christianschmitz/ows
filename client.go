package main

import (
    "fmt"
    "log"
    "os"
    "strconv"
    "strings"
    "github.com/spf13/cobra"
    "ows/actions"
    "ows/ledger"
    "ows/sync"
)

var _ActionsInitialized = actions.InitializeActions()

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

            keyPair := getKeyPair()

            g := ledger.NewGenesisChangeSet(
                actions.NewAddCompute(
                    args[0],
                ),
            )

            // TODO: multi-sig
            signature, err := keyPair.SignChangeSet(g)
            if err != nil {
                log.Fatal(err)
            }

            g.Signatures = append(g.Signatures, signature)

            fmt.Println(g.EncodeToString())
		},
    })

    root.AddCommand(&cobra.Command{
        Use: "nodes",
        Short: "List node addresses",
        Run: func(cmd *cobra.Command, _ []string) {
            l := readLedger()
            m := actions.GetNodeAddresses(l)

            for id, addr := range m {
                fmt.Printf("%s %s\n", id, addr)
            }
        },
    })

    root.AddCommand(&cobra.Command{
        Use: "hashes",
        Short: "List config change hashes (including genesis)",
        Run: func(cmd *cobra.Command, _ []string) {
            c := getSyncedLedgerClient()

            hashes, err := c.GetChangeSetHashes()

            if err != nil {
                log.Fatal(err)
            }

            for i, h := range hashes.Hashes {
                if i == 0 {
                    fmt.Printf("%s\n", ledger.StringifyProjectHash(h))
                } else {
                    fmt.Printf("%s\n", ledger.StringifyChangeSetHash(h))
                }
            }
        },
    })

    root.AddCommand(&cobra.Command{
        Use: "upload",
        Short: "Upload file (fails for directories)",
        Run: func(cmd *cobra.Command, args []string) {
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

                fmt.Printf("%s: %s\n", arg, ledger.StringifyAssetId(id))
            }
        },
    })

    gateways := &cobra.Command{
        Use: "gateways",
        Short: "Manage gateways",
    }

    gateways.AddCommand(&cobra.Command{
        Use: "add",
        Short: "Create a new gateway",
        Run: func (cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                log.Fatal("expected 1 arg")
            }

            c := getSyncedLedgerClient()
            key := getKeyPair()

            port, err := strconv.Atoi(args[0])
            if err != nil {
                log.Fatal(err)
            }

            cs := &ledger.ChangeSet{
                Parent: c.Ledger.Head,
                Actions: []ledger.Action{
                    actions.NewAddGateway(port),
                },
                Signatures: []ledger.Signature{},
            }

            signature, err := key.SignChangeSet(cs)
            if err != nil {
                log.Fatal(err)
            }

            cs.Signatures = append(cs.Signatures, signature)

            if err := c.PublishChangeSet(cs); err != nil {
                log.Fatal(err)
            }

            if err := c.Ledger.AppendChangeSet(cs, false); err != nil {
                log.Fatal(err)
            }

            c.Ledger.Write()
        },
    })

    gateways.AddCommand(&cobra.Command{
        Use: "add-endpoint",
        Short: "Add an endpoint task to a gateway",
        Run: func (cmd *cobra.Command, args []string) {
            if len(args) != 4 {
                log.Fatal("expected 4 args")
            }

            gatewayId, err := ledger.ParseResourceId(args[0])
            if err != nil {
                log.Fatal(err)
            }

            method := args[1]
            path := args[2]
            task, err := ledger.ParseResourceId(args[3])
            if err != nil {
                log.Fatal(err)
            }

            c := getSyncedLedgerClient()
            key := getKeyPair()

            cs := &ledger.ChangeSet{
                Parent: c.Ledger.Head,
                Actions: []ledger.Action{
                    actions.NewAddGatewayEndpoint(gatewayId, method, path, task),
                },
                Signatures: []ledger.Signature{},
            }

            signature, err := key.SignChangeSet(cs)
            if err != nil {
                log.Fatal(err)
            }

            cs.Signatures = append(cs.Signatures, signature)

            if err := c.PublishChangeSet(cs); err != nil {
                log.Fatal(err)
            }

            if err := c.Ledger.AppendChangeSet(cs, false); err != nil {
                log.Fatal(err)
            }

            c.Ledger.Write()
        },
    })

    root.AddCommand(gateways)

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
            
            var err error
            var id ledger.AssetId = [32]byte{}
            if bs, err := os.ReadFile(args[0]); err == nil {
                // upload the file first
                
                id, err = c.UploadFile(bs)
                if err != nil {
                    log.Fatal(err)
                }
            } else if strings.HasPrefix(args[0], "asset") {
                id, err = ledger.ParseAssetId(args[0])
                if err != nil {
                    log.Fatal(err)
                }
            } else {
                log.Fatal("invalid asset " + args[0])
            }

            key := getKeyPair()

            cs := &ledger.ChangeSet{
                Parent: c.Ledger.Head,
                Actions: []ledger.Action{
                    actions.NewAddTask("nodejs", id),
                },
                Signatures: []ledger.Signature{},
            }

            signature, err := key.SignChangeSet(cs)
            if err != nil {
                log.Fatal(err)
            }

            cs.Signatures = append(cs.Signatures, signature)

            if err := c.PublishChangeSet(cs); err != nil {
                log.Fatal(err)
            }

            if err := c.Ledger.AppendChangeSet(cs, false); err != nil {
                log.Fatal(err)
            }

            c.Ledger.Write()
        },
    })

    root.AddCommand(tasks)

    return root
}

func getSyncedLedgerClient() *sync.LedgerClient {
    l, err := ledger.ReadLedger(false)
    if err != nil {
        log.Fatal(err)
    }

    c := sync.NewLedgerClient(l)

    if err := c.Sync(); err != nil {
        log.Fatal(err)
    }

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