package main

import (
	"errors"
	"fmt"
	//"log"
	"os"
	"path"
	//"strconv"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"ows/ledger"
	"ows/resources"
)

const (
	DefaultProjectName = "default"
)

var (
	state      = &clientState{}
	gossipPort uint16 // can't be of type ledger.Port, because cobra flags doesn't accept that
	isOffline  bool
	apiPort    uint16 // can't be of type ledger.Port, because cobra flags doesn't accept that
)

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
	cli.AddCommand(makeFunctionsCLI())
	cli.AddCommand(makeGatewaysCLI())
	//cli.AddCommand(makeInitCommand())
	cli.AddCommand(makeKeyCLI())
	cli.AddCommand(makeLedgerCLI())
	cli.AddCommand(makeNodesCLI())
	cli.AddCommand(makeProjectsCLI())

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

	return withProjectFlags(assetsCLI)
}

func makeFunctionsCLI() *cobra.Command {
	functionsCLI := &cobra.Command{
		Use:   "tasks",
		Short: "List and manage functions",
		//Run:   handleListFunctions,
	}

	//tasksCLI.AddCommand(&cobra.Command{
	//	Use:   "add",
	//	Short: "Create a new task",
	//	Run:   handleAddTask,
	//})

	return withProjectFlags(functionsCLI)
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

	return withProjectFlags(gatewaysCLI)
}

//func makeInitCommand() *cobra.Command {
//	initCmd := &cobra.Command{
//		Use:   "init <node-public-key> <node-address> [--gossip-port <port>] [--api-port <port>]",
//		Short: "Initialize an OWS project with a single bootstrap node",
//		Long: `Initialize an OWS project
//If the gossip port isn't specified, a random port is selected.
//If the sync port isn't specified, a random port is selected`,
//		RunE: handleInitProject,
//	}
//
//	initCmd.Flags().Uint16Var(&gossipPort, "gossip-port", 0, "defaults to a random port")
//	initCmd.Flags().Uint16Var(&apiPort, "api-port", 0, "defaults to a random port")
//
//	return initCmd
//}

func makeKeyCLI() *cobra.Command {
	keyCLI := &cobra.Command{
		Use:   "key",
		Short: "Manage client key",
	}

	keyCLI.AddCommand(&cobra.Command{
		Use:   "gen",
		Short: "Generates a new random key",
		RunE:  handleGenKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use:   "init",
		Short: fmt.Sprintf("Generate a new random key and saves it to %s", state.keyPairPath()),
		RunE:  handleInitClientKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use:   "phrase",
		Short: "Show 24-word key phrase for backup",
		RunE:  handleShowKeyPhrase,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use:   "restore",
		Short: fmt.Sprintf("Restore a key using 24-word phrase (saved to %s)", state.keyPairPath()),
		RunE:  handleRestoreKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use:   "set",
		Short: "Set client key using hex or base64 encoded ed25519 private key",
		RunE:  handleSetKey,
	})

	keyCLI.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show base64 encoded client private key",
		RunE:  handleShowKey,
	})

	return keyCLI
}

func makeLedgerCLI() *cobra.Command {
	ledgerCLI := &cobra.Command{
		Use:   "ledger",
		Short: "Query project ledger",
	}

	ledgerCLI.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List project ledger change set IDs",
		RunE:  handleListLedgerChangeSets,
	})

	return withProjectFlags(ledgerCLI)
}

func makeProjectsCLI() *cobra.Command {
	projectsCLI := &cobra.Command{
		Use:   "projects",
		Short: "Manage projects",
	}

	projectsCLI.AddCommand(&cobra.Command{
		Use:   "default",
		Short: "Set default project",
		RunE:  handleSetDefaultProject,
	})

	projectsCLI.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE:  handleListProjects,
	})

	newProjectCmd := &cobra.Command{
		Use:   "new <project-name> <bootstrap-node-public-key> <bootstrap-node-address>",
		Short: "Create a new project",
		RunE:  handleCreateNewProject,
	}

	newProjectCmd.Flags().Uint16Var(&gossipPort, "gossip-port", 0, "0 results in a random port")
	newProjectCmd.Flags().Uint16Var(&apiPort, "api-port", 0, "0 results in a random port")

	projectsCLI.AddCommand(newProjectCmd)

	projectsCLI.AddCommand(&cobra.Command{
		Use:   "remove <project-name>",
		Short: "Removes a project",
		RunE:  handleRemoveProject,
	})

	return projectsCLI
}

func makeNodesCLI() *cobra.Command {
	nodesCLI := &cobra.Command{
		Use:   "nodes",
		Short: "Manage project nodes",
	}

	nodesCLI.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List project nodes",
		RunE:  handleListNodes,
	})

	return withProjectFlags(nodesCLI)
}

func withProjectFlags(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().BoolVar(&(state.isOffline), "offline", false, "don't sync")
	cmd.PersistentFlags().StringVar(&(state.projectName), "project-name", DefaultProjectName, "project name")

	return cmd
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

func handleCreateNewProject(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(3)(cmd, args); err != nil {
		return err
	}

	projectName := args[0]
	if projectName == DefaultProjectName {
		return fmt.Errorf("project name %q forbidden (hint: use another name and then call `ows projects default <name>`)", projectName)
	}

	d := state.projectsConfigPath()

	isFirst := true
	if _, err := os.Stat(d); err == nil {
		fs, err := os.ReadDir(d)
		if err != nil {
			panic(err)
		}

		isFirst = len(fs) == 0
	}

	mappingPath := path.Join(d, projectName)

	if _, err := os.Stat(mappingPath); err == nil {
		return fmt.Errorf("project %s already exists (at %s)", projectName, mappingPath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("project %s already exists, but failed to read it (%v)", err)
	}

	nodePubKey, err := ledger.ParsePublicKey(args[1])
	if err != nil {
		return fmt.Errorf("invalid node public key %s (%v)", args[1], err)
	}

	// TODO: parse internet address
	address := args[2]

	if gossipPort != 0 && gossipPort == apiPort {
		return fmt.Errorf("gossip port can't be equal to sync port")
	}

	if gossipPort == 0 {
		gp, err := resources.RandomPort()
		if err != nil {
			return fmt.Errorf("failed to generate random gossip port (%v)", err)
		}

		gossipPort = uint16(gp)
	}

	for apiPort == 0 || apiPort == gossipPort {
		sp, err := resources.RandomPort()
		if err != nil {
			return fmt.Errorf("failed to generate random sync port (%v)", err)
		}

		apiPort = uint16(sp)
	}

	initialVersion := ledger.LatestLedgerVersion

	action := ledger.AddNode{
		Key:        nodePubKey,
		Address:    address,
		GossipPort: ledger.Port(gossipPort),
		APIPort:    ledger.Port(apiPort),
	}

	kp := state.keyPair()

	cs := &ledger.ChangeSet{
		Prev:    "",
		Actions: []ledger.Action{action},
	}

	s, err := kp.SignChangeSet(cs)
	if err != nil {
		return fmt.Errorf("failed to sign initial config (%v)", err)
	}

	cs.Signatures = []ledger.Signature{s}

	l, err := ledger.NewLedger(initialVersion, cs)
	if err != nil {
		return fmt.Errorf("failed to create ledger for project %s (%v)", projectName, err)
	}

	projectID := l.ProjectID()

	// write name to ID mapping
	if err := ledger.OverwriteSafe(mappingPath, []byte(projectID)); err != nil {
		return fmt.Errorf("failed to write mapping file for %s to %s (%v)", projectName, mappingPath, err)
	}

	// if first, also write the default file
	if isFirst {
		defaultMappingPath := path.Join(d, DefaultProjectName)
		if err := ledger.OverwriteSafe(defaultMappingPath, []byte(projectID)); err != nil {
			return fmt.Errorf("failed to write default mapping file for %s to %s (%v)", projectName, defaultMappingPath, err)
		}
	}

	lp := path.Join(state.projectsDataPath(), string(projectID), LedgerFileName)

	// write ledger
	if err := l.Write(lp); err != nil {
		return fmt.Errorf("failed to write initial ledger file for %s to %s (%v)", projectName, lp, err)
	}

	fmt.Printf("ProjectID: %s\n", projectID)
	fmt.Printf("InitialConfig: %s\n", l.String())

	return nil
}

func handleGenKey(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	kp, err := ledger.RandomKeyPair()
	if err != nil {
		return fmt.Errorf("unable to generate random key (%v)", err)
	}

	fmt.Println("PrivateKey: ", kp.Private.String())
	fmt.Println("PublicKey: ", kp.Public.String())

	return nil
}

func handleInitClientKey(cmd *cobra.Command, args []string) error {
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

//func handleListFunctions(cmd *cobra.Command, args []string) {
//	if len(args) != 0 {
//		log.Fatal("unexpected args")
//	}
//
//	c := getSyncedLedgerClient()
//
//	functions := ledger.ListFunctions(c.Ledger)
//
//	for id, config := range functions {
//		fmt.Printf("%s %s %s\n", id, config.Runtime, config.Handler)
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

func handleListLedgerChangeSets(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	l := state.ledger()
	chain := l.IDChain()

	for _, id := range chain.IDs {
		fmt.Println(id)
	}

	return nil
}

func handleListNodes(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	l := state.ledger()
	s := l.Snapshot

	for id, conf := range s.Nodes {
		fmt.Printf("%s %s %d %d\n", id, conf.Address, conf.GossipPort, conf.APIPort)
	}

	return nil
}

type projectsTable struct {
	rows []projectsTableRow
}

type projectsTableRow struct {
	name      string
	projectID string
	isDefault bool
}

func (t *projectsTable) Len() int {
	return len(t.rows)
}

func (t *projectsTable) Less(i, j int) bool {
	a := t.rows[i]
	b := t.rows[j]

	return a.name < b.name
}

func (t *projectsTable) Swap(i, j int) {
	a := t.rows[i]
	b := t.rows[j]

	t.rows[i] = b
	t.rows[j] = a
}

func handleListProjects(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	d := state.projectsConfigPath()

	if _, err := os.Stat(d); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		} else {
			return err
		}
	}

	projects, err := os.ReadDir(d)
	if err != nil {
		return err
	}

	var defaultProjectID string

	for _, f := range projects {
		name := f.Name()
		if name == DefaultProjectName {
			p := path.Join(d, name)

			bs, err := os.ReadFile(p)
			if err != nil {
				return err
			}

			defaultProjectID = string(bs)
			if err := ledger.ValidateID(defaultProjectID, ledger.ProjectIDPrefix); err != nil {
				return fmt.Errorf("invalid project id in %s (%v)", p, err)
			}
		}
	}

	defaultFound := false
	rows := make([]projectsTableRow, 0)

	for _, f := range projects {
		name := f.Name()

		if name != DefaultProjectName {
			p := path.Join(d, name)

			bs, err := os.ReadFile(p)
			if err != nil {
				return err
			}

			projectID := string(bs)
			if err := ledger.ValidateID(projectID, ledger.ProjectIDPrefix); err != nil {
				return fmt.Errorf("invalid project id in %s (%v)", p, err)
			}

			if projectID == defaultProjectID {
				defaultFound = true
			}

			rows = append(rows, projectsTableRow{name, projectID, projectID == defaultProjectID})
		}
	}

	if defaultProjectID != "" && !defaultFound {
		rows = append(rows, projectsTableRow{DefaultProjectName, defaultProjectID, true})
	}

	table := &projectsTable{rows}

	sort.Sort(table)

	// sort the table
	for _, row := range table.rows {
		defaultStar := ""
		if row.isDefault {
			defaultStar = "*"
		}

		fmt.Printf("%s %s %s\n", row.name, row.projectID, defaultStar)
	}

	return nil
}

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

func handleRemoveProject(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	projectName := args[0]
	if projectName == DefaultProjectName {
		return fmt.Errorf("invalid project name %q", projectName)
	}

	d := state.projectsConfigPath()

	mappingPath := path.Join(d, projectName)

	bs, err := os.ReadFile(mappingPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("project %s not found in %s", projectName, d)
		} else {
			return fmt.Errorf("error reading project %s at %s (%v)", projectName, mappingPath, err)
		}
	}

	projectID := string(bs)
	if err := ledger.ValidateID(projectID, ledger.ProjectIDPrefix); err != nil {
		return fmt.Errorf("invalid project id at %s (%v)", mappingPath, err)
	}

	// check if default matches this
	defaultMappingPath := path.Join(d, DefaultProjectName)
	bsDef, err := os.ReadFile(defaultMappingPath)
	if err == nil {
		if string(bsDef) == projectID {
			if err := os.Remove(defaultMappingPath); err != nil {
				return err
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	projectDataPath := path.Join(state.projectsDataPath(), projectID)
	if err := os.RemoveAll(projectDataPath); err != nil {
		return err
	}

	if err := os.Remove(mappingPath); err != nil {
		return err
	}

	fmt.Printf("Removed project %s\n", projectID)

	return nil
}

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

func handleSetDefaultProject(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(1)(cmd, args); err != nil {
		return err
	}

	projectName := args[0]
	if projectName == DefaultProjectName {
		return fmt.Errorf("invalid project name %q", projectName)
	}

	d := state.projectsConfigPath()

	mappingPath := path.Join(d, projectName)

	bs, err := os.ReadFile(mappingPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("project %s not found in %s", projectName, d)
		} else {
			return fmt.Errorf("error reading project %s at %s (%v)", projectName, mappingPath, err)
		}
	}

	projectID := string(bs)
	if err := ledger.ValidateID(projectID, ledger.ProjectIDPrefix); err != nil {
		return fmt.Errorf("invalid project id at %s (%v)", mappingPath, err)
	}

	defaultPath := path.Join(d, DefaultProjectName)

	if err := ledger.OverwriteSafe(defaultPath, []byte(projectID)); err != nil {
		return err
	}

	fmt.Printf("The default project is now %s (a.k.a. %s)\n", projectID, projectName)

	return nil
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
