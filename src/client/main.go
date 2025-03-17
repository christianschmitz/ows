package main

import (
	"errors"
	"fmt"
	//"log"
	"os"
	//"strconv"
	"strings"

	"github.com/spf13/cobra"

	"ows/ledger"
)

var state = &clientState{}
var gossipPort uint16
var syncPort uint16

func main() {
	cli := makeCLI()

	cli.Execute()
}

func makeCLI() *cobra.Command {
	cli := &cobra.Command{
		Use:   "ows",
		Short: "Open Web Services CLI",
	}

	cli.AddCommand(makeAssetsCLI())
	cli.AddCommand(makeGatewaysCLI())
	//cli.AddCommand(makeInitCommand())
	cli.AddCommand(makeKeyCLI())
	cli.AddCommand(makeLedgerCLI())
	//cli.AddCommand(makeNodesCLI())
	cli.AddCommand(makeTasksCLI())

	cli.PersistentFlags().StringVar(&(state.projectName), "project-name", "default", "project name (defaults to 'default')")
	cli.PersistentFlags().StringVar(&(state.testDir), "test-dir", "", "test directory")

	return cli
}

func makeAssetsCLI() *cobra.Command {
	assetsCLI := &cobra.Command{
		Use:   "assets",
		Short: "List assets",
		//Run:   handleListAssets,
	}

	assetsCLI.AddCommand(&cobra.Command{
		Use:   "upload",
		Short: "Upload file (fails for directories)",
		//Run:   handleUploadAssets,
	})

	return assetsCLI
}

func makeGatewaysCLI() *cobra.Command {
	gatewaysCLI := &cobra.Command{
		Use:   "gateways",
		Short: "List and manage gateways",
		//Run:   handleListGateways,
	}

	//gatewaysCLI.AddCommand(&cobra.Command{
	//	Use:   "add",
	//	Short: "Create a new gateway",
	//	Run:   handleAddGateway,
	//})

	//gatewaysCLI.AddCommand(&cobra.Command{
	//	Use:   "remove",
	//	Short: "Remove a gateway",
	//	Run:   handleRemoveGateway,
	//})

	//gatewaysCLI.AddCommand(&cobra.Command{
	//	Use:   "add-endpoint",
	//	Short: "Add an endpoint task to a gateway",
	//	Run:   handleAddGatewayEndpoint,
	//})

	return gatewaysCLI
}

//func makeInitCommand() *cobra.Command {
//	initCmd := &cobra.Command{
//		Use:   "init <node-public-key> <node-address> [--gossip-port <port>] [--sync-port <port>]",
//		Short: "Initialize an OWS project with a single bootstrap node",
//		Long: `Initialize an OWS project
//If the gossip port isn't specified, a random port is selected.
//If the sync port isn't specified, a random port is selected`,
//		RunE: handleInitProject,
//	}
//
//	initCmd.Flags().Uint16Var(&gossipPort, "gossip-port", 0, "defaults to a random port")
//	initCmd.Flags().Uint16Var(&syncPort, "sync-port", 0, "defaults to a random port")
//
//	return initCmd
//}

func makeKeyCLI() *cobra.Command {
	keyCLI := &cobra.Command{
		Use: "key",
		Short: "Manage client key",
	}
	
	keyCLI.AddCommand(&cobra.Command{
		Use: "gen",
		Short: fmt.Sprintf("Generate a new random key (saved to %s)", state.keyPairPath()),
		RunE: handleGenKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use: "phrase",
		Short: "Show 24-word key phrase for backup",
		RunE: handleShowKeyPhrase,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use: "restore",
		Short: fmt.Sprintf("Restore a key using 24-word phrase (saved to %s)", state.keyPairPath()),
		RunE: handleRestoreKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use: "set",
		Short: "Set client key using hex or base64 encoded ed25519 private key",
		RunE: handleSetKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use: "show",
		Short: "Show base64 encoded client private key",
		RunE: handleShowKey,
	})

	return keyCLI
}

func makeLedgerCLI() *cobra.Command {
	return &cobra.Command{
		Use:   "ledger",
		Short: "List config change hashes (including genesis)",
		//Run:   handleListLedgerChangeSets,
	}
}

//func makeNodesCLI() *cobra.Command {
//	return &cobra.Command{
//		Use:   "nodes",
//		Short: "List and manage nodes",
//		Run:   handleListNodes,
//	}
//}

func makeTasksCLI() *cobra.Command {
	tasksCLI := &cobra.Command{
		Use:   "tasks",
		Short: "List and manage tasks",
		//Run:   handleListTasks,
	}

	//tasksCLI.AddCommand(&cobra.Command{
	//	Use:   "add",
	//	Short: "Create a new task",
	//	Run:   handleAddTask,
	//})

	return tasksCLI
}

//func handleAddGateway(cmd *cobra.Command, args []string) {
//	if len(args) != 1 {
//		log.Fatal("expected 1 arg")
//	}
//
//	port, err := strconv.Atoi(args[0])
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	createChangeSet(ledger.NewAddGateway(port))
//}

//func handleAddGatewayEndpoint(cmd *cobra.Command, args []string) {
//	if len(args) != 4 {
//		log.Fatal("expected 4 args")
//	}
//
//	gatewayId := args[0]
//	if err := ledger.ValidateResourceId(gatewayId, "gateway"); err != nil {
//		log.Fatal(err)
//	}
//
//	method := args[1]
//	if method != "GET" && method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
//		log.Fatal("invalid method " + method)
//	}
//
//	path := args[2]
//
//	taskId := args[3]
//	if err := ledger.ValidateResourceId(taskId, "task"); err != nil {
//		log.Fatal(err)
//	}
//
//	createChangeSet(ledger.NewAddGatewayEndpoint(gatewayId, method, path, taskId))
//}

//func handleAddTask(cmd *cobra.Command, args []string) {
//	if len(args) != 2 {
//		log.Fatal("expected 2 args")
//	}
//
//	c := getSyncedLedgerClient()
//
//	runtime := args[0]
//
//	if runtime != "nodejs" {
//		log.Fatal("only nodejs runtime is currently supported")
//	}
//
//	handler := args[1]
//
//	id := ""
//	if bs, err := os.ReadFile(handler); err == nil {
//		// upload the file first
//
//		id, err = c.UploadFile(bs)
//		if err != nil {
//			log.Fatal(err)
//		}
//	} else if strings.HasPrefix(handler, "asset") {
//		if err := ledger.ValidateResourceId(handler, "asset"); err != nil {
//			log.Fatal(err)
//		}
//
//		id = handler
//	} else {
//		log.Fatal("invalid handler asset " + handler)
//	}
//
//	cs := c.Ledger.NewChangeSet(ledger.NewAddTask("nodejs", id))
//
//	if err := signAndSubmitChangeSet(c, cs); err != nil {
//		log.Fatal(err)
//	}
//}

func handleGenKey(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	kp, err := ledger.RandomKeyPair()
	if err != nil {
		return fmt.Errorf("unable to generate random key (%v)", err) 
	}

	return saveKeyPair(kp)
}

//func handleInitProject(cmd *cobra.Command, args []string) error {
//	if err := cobra.ExactArgs(2)(cmd, args); err != nil {
//		return err
//	}
//
//	//nodePubKey := ledger.ParsePubKey(args[0])
//	//rawNodeAddress := args[1]
//
//	keyPair := getKeyPair()
//
//	g := ledger.NewGenesisChangeSet(
//		ledger.NewAddNode(
//			args[0],
//		),
//	)
//
//	// TODO: multi-sig
//	signature, err := keyPair.SignChangeSet(g)
//	if err != nil {
//		return err
//	}
//
//	g.Signatures = append(g.Signatures, signature)
//
//	fmt.Println(g.EncodeToString())
//
//	return nil
//}
//
//func handleListAssets(cmd *cobra.Command, args []string) {
//	if len(args) != 0 {
//		log.Fatal("expected 0 args")
//	}
//
//	c := getSyncedLedgerClient()
//
//	assets, err := c.GetAssets()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, a := range assets {
//		fmt.Println(a)
//	}
//}

//func handleListGateways(cmd *cobra.Command, args []string) {
//	if len(args) != 0 {
//		log.Fatal("unexpected args")
//	}
//
//	c := getSyncedLedgerClient()
//	gateways := ledger.ListGateways(c.Ledger)
//
//	for id, config := range gateways {
//		fmt.Printf("%s %d\n", id, config.Port)
//	}
//}

//func handleListLedgerChangeSets(cmd *cobra.Command, _ []string) {
//	c := getSyncedLedgerClient()
//
//	hashes, err := c.GetChangeSetIDs()
//
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for i, h := range hashes.Hashes {
//		if i == 0 {
//			fmt.Printf("%s\n", ledger.StringifyProjectHash(h))
//		} else {
//			fmt.Printf("%s\n", ledger.StringifyChangeSetID(h))
//		}
//	}
//}

//func handleListNodes(cmd *cobra.Command, args []string) {
//	l := readLedger()
//	m := ledger.ListNodes(l)
//
//	for id, addr := range m {
//		fmt.Printf("%s %s\n", id, addr)
//	}
//}

//func handleListTasks(cmd *cobra.Command, args []string) {
//	if len(args) != 0 {
//		log.Fatal("unexpected args")
//	}
//
//	c := getSyncedLedgerClient()
//
//	tasks := ledger.ListTasks(c.Ledger)
//
//	for id, config := range tasks {
//		fmt.Printf("%s %s %s\n", id, config.Runtime, config.Handler)
//	}
//}

//func handleRemoveGateway(cmd *cobra.Command, args []string) {
//	if len(args) != 1 {
//		log.Fatal("expected 1 arg")
//	}
//
//	gatewayId := args[0]
//	if err := ledger.ValidateResourceId(gatewayId, "gateway"); err != nil {
//		log.Fatal(err)
//	}
//
//	createChangeSet(ledger.NewRemoveGateway(gatewayId))
//}

func handleRestoreKey(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(24)(cmd, args); err != nil {
		return err
	}

	kp, err := ledger.RestoreKeyPair(args)
	if err != nil {
		return err
	}

	return saveKeyPair(kp)
}

func handleSetKey(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	privateKeyStr := args[0]

	privateKey, err := ledger.ParsePrivateKey(privateKeyStr)
	if err != nil {
		return err
	}

	kp := privateKey.KeyPair()

	return saveKeyPair(kp)
}

func handleShowKey(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	kp := state.keyPair()

	fmt.Println(kp.Private.String())

	return nil
}

func handleShowKeyPhrase(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	kp := state.keyPair()

	phrase, err := kp.Phrase()
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(phrase, " "))

	return nil
}

//func handleUploadAssets(cmd *cobra.Command, args []string) {
//	if len(args) < 1 {
//		log.Fatal("expected at least 1 arg")
//	}
//
//	c := getSyncedLedgerClient()
//
//	for _, arg := range args {
//		bs, err := os.ReadFile(arg)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		id, err := c.UploadFile(bs)
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Printf("%s: %s\n", arg, id)
//	}
//}
//

//func getLedgerPath(genesis *ChangeSet) string {
//	h := StringifyProjectHash(genesis.Hash())
//
//	projectPath := HomeDir + "/" + h
//
//	if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
//		log.Fatal(err)
//	}
//
//	return makeLedgerPath(projectPath)
//}

func saveKeyPair(kp *ledger.KeyPair) error {
	p := state.keyPairPath()

	// Make sure key doesn't already exist
	if _, err := os.Stat(p); err == nil {
		return fmt.Errorf("key already exists at %s", p)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("an error occured while reading existing key at %s (%v)", p, err)
	}

	if err := kp.Write(p); err != nil {
		return fmt.Errorf("failure while writing key to %s (%v)", p, err)
	}

	return nil
}