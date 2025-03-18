package resources

import (
	"fmt"

	"ows/ledger"
)

type Node struct {
	Config ledger.NodeConfig
}

type NodeManager struct {
	Nodes map[ledger.NodeID]*Node
}

func newNodeManager() *NodeManager {
	return &NodeManager{
		map[ledger.NodeID]*Node{},
	}
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
