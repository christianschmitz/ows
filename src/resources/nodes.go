package resources

import (
	"errors"
	"fmt"
	"log"

	"ows/ledger"
	"ows/network"
)

type Node struct {
	Config ledger.NodeConfig
}

type NodeManager struct {
	Current *ledger.KeyPair
	Nodes   map[ledger.NodeID]*Node
}

func newNodeManager(current *ledger.KeyPair) *NodeManager {
	return &NodeManager{
		Current: current,
		Nodes:   map[ledger.NodeID]*Node{},
	}
}

func (m *NodeManager) CurrentNodeID() ledger.NodeID {
	return m.Current.Public.NodeID()
}

func (m *NodeManager) OtherNodeIDs() []ledger.NodeID {
	nodeIDs := make([]ledger.NodeID, 0)
	currentID := m.CurrentNodeID()

	for id, _ := range m.Nodes {
		if id != currentID {
			nodeIDs = append(nodeIDs, id)
		}
	}

	return nodeIDs
}

func (m *NodeManager) NewClient(id ledger.NodeID) (*network.NodeAPIClient, error) {
	if id == m.CurrentNodeID() {
		return nil, errors.New("can't connect to self")
	}

	n, ok := m.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}

	return network.NewNodeAPIClient(m.Current, n.Config.Address, n.Config.APIPort, m.Nodes), nil
}

func (m *NodeManager) Sync(nodes map[ledger.NodeID]ledger.NodeConfig) error {
	for id, conf := range nodes {
		if _, ok := m.Nodes[id]; ok {
			if err := m.update(id, conf); err != nil {
				return fmt.Errorf("failed to update node %s (%v)", id, err)
			}
		} else {
			if err := m.add(id, conf); err != nil {
				return fmt.Errorf("failed to add node %s (%v)", id, err)
			}
		}
	}

	for id, _ := range m.Nodes {
		if _, ok := nodes[id]; !ok {
			if err := m.remove(id); err != nil {
				return fmt.Errorf("failed to remove node %s (%v)", id, err)
			}
		}
	}

	return nil
}

func (m *NodeManager) add(id ledger.NodeID, config ledger.NodeConfig) error {
	log.Printf("added node %s\n", id)

	m.Nodes[id] = &Node{
		Config: config,
	}

	return nil
}

func (m *NodeManager) remove(id ledger.NodeID) error {
	delete(m.Nodes, id)

	return nil
}

func (m *NodeManager) update(id ledger.NodeID, config ledger.NodeConfig) error {
	m.Nodes[id].Config = config

	return nil
}
