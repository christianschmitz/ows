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

	// Get own node config
	kp := state.keyPair()
	id := kp.Public.NodeID()
	l := state.ledger()

	// TODO: try to sync from other nodes

	conf, ok := l.Snapshot.Nodes[id]
	if !ok {
		panic("own id not found in ledger (TODO: sync from other nodes first)")
	}

	state.resources = resources.NewManager(state.assetsPath())
	state.resources.Sync(l.Snapshot)

	fmt.Printf("Starting OWS node for %s\n", l.ProjectID())

	go network.ServeAPI(conf.APIPort, kp, state)
	fmt.Printf("Hosting node API at https://%s:%d\n", conf.Address, conf.APIPort)

	go network.ServeGossip(conf.GossipPort, state)
	fmt.Printf("Hosting gossip service at https://%s:%d\n", conf.Address, conf.GossipPort)

	return nil
}
