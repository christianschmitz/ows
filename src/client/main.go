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
                actions.NewAddNode(
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
            m := actions.ListNodes(l)

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
        Use: "assets",
        Short: "List assets",
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) != 0 {
                log.Fatal("expected 0 args")
            }

            c := getSyncedLedgerClient()

            assets, err := c.GetAssets()
            if err != nil {
                log.Fatal(err)
            }

            for _, a := range assets {
                fmt.Println(a)
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

                fmt.Printf("%s: %s\n", arg, id)
            }
        },
    })

    gateways := &cobra.Command{
        Use: "gateways",
        Short: "List and manage gateways",
        Run: func (cmd *cobra.Command, args []string) {
            if len(args) != 0 {
                log.Fatal("unexpected args")
            }

            c := getSyncedLedgerClient()
            gateways := actions.ListGateways(c.Ledger)

            for id, config := range gateways {
                fmt.Printf("%s %d\n", id, config.Port)
            }
        },
    }

    gateways.AddCommand(&cobra.Command{
        Use: "add",
        Short: "Create a new gateway",
        Run: func (cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                log.Fatal("expected 1 arg")
            }

            port, err := strconv.Atoi(args[0])
            if err != nil {
                log.Fatal(err)
            }

            createChangeSet(actions.NewAddGateway(port))
        },
    })

    gateways.AddCommand(&cobra.Command{
        Use: "remove",
        Short: "Remove a gateway",
        Run: func (cmd *cobra.Command, args []string) {
            if len(args) != 1 {
                log.Fatal("expected 1 arg")
            }

            gatewayId := args[0]
            if err := ledger.ValidateResourceId(gatewayId, "gateway"); err != nil {
                log.Fatal(err)
            }

            createChangeSet(actions.NewRemoveGateway(gatewayId))
        },
    })

    gateways.AddCommand(&cobra.Command{
        Use: "add-endpoint",
        Short: "Add an endpoint task to a gateway",
        Run: func (cmd *cobra.Command, args []string) {
            if len(args) != 4 {
                log.Fatal("expected 4 args")
            }

            gatewayId := args[0]
            if err := ledger.ValidateResourceId(gatewayId, "gateway"); err != nil {
                log.Fatal(err)
            }

            method := args[1]
            if (method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE") {
                log.Fatal("invalid method " + method)
            }

            path := args[2]

            taskId := args[3]
            if err := ledger.ValidateResourceId(taskId, "task"); err != nil {
                log.Fatal(err)
            }

            createChangeSet(actions.NewAddGatewayEndpoint(gatewayId, method, path, taskId))            
        },
    })

    root.AddCommand(gateways)

    tasks := &cobra.Command{
        Use: "tasks",
        Short: "List and manage tasks",
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) != 0 {
                log.Fatal("unexpected args")
            }

            c := getSyncedLedgerClient()

            tasks := actions.ListTasks(c.Ledger)

            for id, config := range tasks {
                fmt.Printf("%s %s %s\n", id, config.Runtime, config.Handler)
            }
        },
    }

    tasks.AddCommand(&cobra.Command{
        Use: "add",
        Short: "Create a new task",
        Run: func(cmd *cobra.Command, args []string) {
            if len(args) != 2 {
                log.Fatal("expected 2 args")
            }

            c := getSyncedLedgerClient()

            runtime := args[0]

            if runtime != "nodejs" {
                log.Fatal("only nodejs runtime is currently supported")
            }

            handler := args[1]
            
            id := ""
            if bs, err := os.ReadFile(handler); err == nil {
                // upload the file first
                
                id, err = c.UploadFile(bs)
                if err != nil {
                    log.Fatal(err)
                }
            } else if strings.HasPrefix(handler, "asset") {
                if err := ledger.ValidateResourceId(handler, "asset"); err != nil {
                    log.Fatal(err)
                }

                id = handler
            } else {
                log.Fatal("invalid handler asset " + handler)
            }

            cs := c.Ledger.NewChangeSet(actions.NewAddTask("nodejs", id))

            if err := signAndSubmitChangeSet(c, cs); err != nil {
                log.Fatal(err)
            }
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

func signAndSubmitChangeSet(client *sync.LedgerClient, cs *ledger.ChangeSet) error {
    key := getKeyPair()

    signature, err := key.SignChangeSet(cs)
    if err != nil {
        return err
    }

    cs.Signatures = append(cs.Signatures, signature)

    if err := client.PublishChangeSet(cs); err != nil {
        return err
    }

    if err := client.Ledger.AppendChangeSet(cs, false); err != nil {
        return err
    }

    client.Ledger.Write()

    return err
}

// creates change set, signs it, then submits it
func createChangeSet(actions ...ledger.Action) {
    c := getSyncedLedgerClient()
    cs := c.Ledger.NewChangeSet(actions...)
    if err := signAndSubmitChangeSet(c, cs); err != nil {
        log.Fatal(err)
    }
}