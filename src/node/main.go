package main

import (
	"fmt"
	//"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	//"ows/ledger"
	"ows/network"
	"ows/resources"
)

const keyPath = "/etc/ows/key"

var (
	state = &nodeState{}
)

func main() {
	cli := &cobra.Command{
		Use:   "ows-node",
		Short: "Open Web Services node",
		RunE:  handleStartNode,
	}

	cli.Flags().StringVar(&(state.testDir), "test-dir", "", "test directory")

	cli.Execute()

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	//l, err := ledger.ReadLedger(true)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//
	//
	//
	//// TODO: sync from the snapshot instead
	//l.ApplyAll(rm)
	//
	//go ledger.ListenAndServeLedger(l, rm)
	//
	//select {}
}

func handleStartNode(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	// get own node config
	kp := state.keyPair()
	id := kp.Public.NodeID()
	l := state.ledger()

	// TODO: try to sync from other nodes

	conf, ok := l.Snapshot.Nodes[id]
	if !ok {
		panic("own id not found in ledger (TODO: sync from other nodes first)")
	}

	// Start the manager
	state.resources = resources.NewManager(state.assetsPath())
	state.resources.Sync(l.Snapshot)
	// TODO: sync resources

	go network.ServeAPI(conf.APIPort, state)
	fmt.Printf("Hosting node API at https://%s:%d\n", conf.Address, conf.APIPort)

	// TODO: start the gossip service

	fmt.Printf("Started OWS node for %s\n", l.ProjectID())

	return nil
}

// Preference order:
//   1. `/etc/ows/key` or
//   2.
//func readKeyPair() *KeyPair {
//	if testDir == "" {
//		_, err := os.Stat(keyPath)
//		if err == nil {
//			k, err := ledger.ReadKeyPair(keyPath)
//			if err != nil {
//				panic(fmt.Sprintf9"unable to read keypair at %s (%v)", keyPath, err)
//			}
//
//			return k
//		} else if !errors.Is(err, os.ErrNotExist) {
//			panic(fmt.Sprintf("failed to stat /etc/ows/key (%v)", err))
//		}
//	}
//
//	str, exists := os.LookupEnv(GENESIS_ENV_VAR_NAME)
//
//
//}

//func initializeHomeDir() {
//	path, exists := os.LookupEnv("HOME")
//
//	if exists {
//		path = path + "/.ows/node"
//	} else {
//		// assume that if HOME isn't set the node has root user rights
//		path = "/ows"
//	}
//
//	ledger.SetHomeDir(path)
//
//	fmt.Println("Home dir: " + path)
//}
