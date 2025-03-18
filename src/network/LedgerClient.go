package network

import (
	"log"

	"ows/ledger"
)

type LedgerClient struct {
	Ledger *ledger.Ledger
}

func NewLedgerClient(l *ledger.Ledger) *LedgerClient {
	return &LedgerClient{l}
}

// Pick any node
// Request the head of that node's ledger
// If the head is the same exit
// If the head isn't the same, fetch all tx Ids
// Find the intersection (last common point)
// Download everything after the intersection
func (c *LedgerClient) Sync() error {
	node := c.PickNode()

	head, err := node.Head()
	if err != nil {
		return err
	}

	if c.Ledger.Head() == head {
		return nil
	}

	remoteChangeSetIDs, err := node.ChangeSetIDChain()
	if err != nil {
		return err
	}

	thisChangeSetIDs := c.Ledger.IDChain()

	p, err := thisChangeSetIDs.Intersect(remoteChangeSetIDs)
	if err != nil {
		return err
	}

	// remove [p+1:] from local ledger
	c.Ledger.Keep(p)

	// download [p+1:] from remote ledger
	if p+1 < len(remoteChangeSetIDs.IDs) {
		for i := int(p) + 1; i < len(remoteChangeSetIDs.IDs); i++ {
			h := remoteChangeSetIDs.IDs[i]

			cs, err := node.ChangeSet(h)
			if err != nil {
				return err
			}

			if err := c.Ledger.Append(cs); err != nil {
				return err
			}
		}
	}

	c.Ledger.Write("")

	return nil
}

// returns the node address
func (c *LedgerClient) PickNode() *NodeSyncClient {
	m := c.Ledger.Snapshot.Nodes

	for _, conf := range m {
		return NewNodeSyncClient(conf.Address, conf.APIPort)
	}

	log.Fatal("no nodes found")

	return nil
}

func (c *LedgerClient) ChangeSetIDChain() (*ledger.ChangeSetIDChain, error) {
	node := c.PickNode()

	return node.ChangeSetIDChain()
}

func (c *LedgerClient) PublishChangeSet(cs *ledger.ChangeSet) error {
	node := c.PickNode()

	return node.PublishChangeSet(cs)
}

func (c *LedgerClient) UploadFile(bs []byte) (string, error) {
	node := c.PickNode()

	return node.UploadFile(bs)
}

func (c *LedgerClient) GetAssets() ([]string, error) {
	node := c.PickNode()

	return node.GetAssets()
}
