package ledger

import (
	"log"
)

type LedgerClient struct {
	Ledger *Ledger
}

func NewLedgerClient(l *Ledger) *LedgerClient {
	return &LedgerClient{l}
}

// Pick any node
// Request the head of that node's ledger
// If the head is the same exit
// If the head isn't the same, fetch all tx Ids
// Find the intersection (last common point)
// Download everything after the intersection
func (c *LedgerClient) Sync() {
	node := c.pickNode()

	head := node.GetHead()

	if (IsSameChangeSetHash(c.Ledger.Head, head)) {
		return
	}

	remoteChangeSetHashes := node.GetChangeSetHashes()
	thisChangeSetHashes := c.Ledger.GetChangeSetHashes()

	p, err := thisChangeSetHashes.FindIntersection(remoteChangeSetHashes)

	if err != nil {
		log.Fatal(err)
	}

	// remove p+1: from this ledger
	c.Ledger.KeepChangeSets(p)

	// download q+1: from remote
	if p+1 < len(remoteChangeSetHashes.Hashes) {
		for i := p+1; i < len(remoteChangeSetHashes.Hashes); i++ {
			h := remoteChangeSetHashes.Hashes[i]

			cs, err := node.GetChangeSet(h)

			if err != nil {
				log.Fatal(err)
			}

			c.Ledger.AppendChangeSet(cs)
		}
	}

	c.Ledger.Persist()
}

// returns the node address
func (c *LedgerClient) pickNode() *NodeSyncClient {
	m := c.Ledger.GetNodeAddresses()

	for _, a := range m {
		return NewNodeSyncClient(a)
	}

	log.Fatal("no nodes found")

	return nil
}

func (c *LedgerClient) GetChangeSetHashes() *ChangeSetHashes {
	node := c.pickNode()

	return node.GetChangeSetHashes()
}

func (c *LedgerClient) PublishChangeSet(cs *ChangeSet) {
	node := c.pickNode()

	node.PublishChangeSet(cs)
}