package resources

import (
	"errors"
	"fmt"
	"log"

	"ows/ledger"
	"ows/network"
)

func (m *Manager) CurrentNodeID() ledger.NodeID {
	return m.Current.Public.NodeID()
}

func (m *Manager) OtherNodeIDs() []ledger.NodeID {
	nodeIDs := make([]ledger.NodeID, 0)
	currentID := m.CurrentNodeID()

	for id, _ := range m.Nodes {
		if id != currentID {
			nodeIDs = append(nodeIDs, id)
		}
	}

	return nodeIDs
}

func (m *Manager) NewNodeAPIClient(id ledger.NodeID) (*network.NodeAPIClient, error) {
	if id == m.CurrentNodeID() {
		return nil, errors.New("can't connect to self")
	}

	n, ok := m.Nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}

	return network.NewNodeAPIClient(m.Current, n.Config.Address, n.Config.APIPort, m.Nodes), nil
}

func (m *Manager) SyncNodes(nodes map[ledger.NodeID]ledger.NodeConfig) error {
	for id, conf := range nodes {
		if _, ok := m.Nodes[id]; ok {
			if err := m.updateNode(id, conf); err != nil {
				return fmt.Errorf("failed to update node %s (%v)", id, err)
			}
		} else {
			if err := m.addNode(id, conf); err != nil {
				return fmt.Errorf("failed to add node %s (%v)", id, err)
			}
		}
	}

	for id, _ := range m.Nodes {
		if _, ok := nodes[id]; !ok {
			if err := m.removeNode(id); err != nil {
				return fmt.Errorf("failed to remove node %s (%v)", id, err)
			}
		}
	}

	return nil
}

func (m *Manager) addNode(id ledger.NodeID, config ledger.NodeConfig) error {
	log.Printf("added node %s\n", id)

	m.Nodes[id] = &Node{
		Config: config,
	}

	return nil
}

func (m *Manager) removeNode(id ledger.NodeID) error {
	delete(m.Nodes, id)

	return nil
}

func (m *Manager) updateNode(id ledger.NodeID, config ledger.NodeConfig) error {
	m.Nodes[id].Config = config

	return nil
}
