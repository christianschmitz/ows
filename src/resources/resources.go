package resources

import (
	"ows/ledger"
)

type Manager struct {
	Assets    *AssetManager
	Functions *FunctionManager
	Gateways  *GatewaysManager
	Nodes     *NodeManager
}

func NewManager(current *ledger.KeyPair, assetsDir string) *Manager {
	nodes := newNodeManager(current)
	assets := newAssetManager(assetsDir, nodes)
	functions := newFunctionManager(assets)
	gateways := newGatewaysManager(functions)

	return &Manager{
		Assets:    assets,
		Functions: functions,
		Gateways:  gateways,
		Nodes:     nodes,
	}
}

func (m *Manager) Sync(snapshot *ledger.Snapshot) error {
	if err := m.Functions.Sync(snapshot.Functions); err != nil {
		return err
	}

	if err := m.Gateways.Sync(snapshot.Gateways); err != nil {
		return err
	}

	if err := m.Nodes.Sync(snapshot.Nodes); err != nil {
		return err
	}

	return nil
}
