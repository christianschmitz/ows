package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"ows/ledger"
	"ows/network"
	"ows/resources"
)

const keyPath = "/etc/ows/key"

var (
	Version        = "dev" // set externally
	state          = &nodeState{}
	testPortOffset = 0
)

func main() {
	cli := &cobra.Command{
		Use:   "ows-node",
		Short: "Open Web Services node",
		RunE:  handleStartNode,
	}

	cli.Flags().StringVar(&(state.testDir), "test-dir", "", "test directory")
	cli.Flags().IntVar(&testPortOffset, "test-port-offset", 0, "port offsets (for testing locally)")

	cli.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE:  handleShowVersion,
	})

	cli.Execute()
}

func handleShowVersion(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return nil
	}

	fmt.Println(Version)

	return nil
}

func handleStartNode(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(0)(cmd, args); err != nil {
		return err
	}

	// Get own node config
	kp := state.keyPair()
	l := state.ledger()
	id := kp.Public.NodeID()

	// Set resource object
	log.Printf("starting OWS node for %s\n", l.ProjectID())
	state.resources = resources.NewManager(kp, state.assetsPath(), testPortOffset)
	state.resources.Sync(l.Snapshot)

	// Sync from other nodes (if other nodes are available)
	otherNodes := make([]ledger.NodeID, 0)

	for otherID, _ := range l.Snapshot.Nodes {
		if otherID != id {
			otherNodes = append(otherNodes, otherID)
		}
	}

	if len(otherNodes) >= 1 {
		c := network.NewAPIClient(kp, state)
		if err := c.Sync(); err != nil {
			panic(fmt.Sprintf("failed to sync upon startup (%v)", err))
		}
	}

	conf, ok := l.Snapshot.Nodes[id]
	if !ok {
		panic(fmt.Sprintf("own node id %s not found in synced ledger", id))
	}

	go network.ServeAPI(conf.APIPort, kp, state)
	log.Printf("hosting node API at https://%s:%d\n", conf.Address, conf.APIPort)

	go network.ServeGossip(conf.GossipPort, kp, state)
	log.Printf("hosting gossip service at https://%s:%d\n", conf.Address, conf.GossipPort)

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	return nil
}
