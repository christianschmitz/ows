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

func NewManager(assetsDir string) *Manager {
	assets := newAssetManager(assetsDir)
	functions := newFunctionManager(assets)
	gateways := newGatewaysManager(functions)
	nodes := newNodeManager()

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

//func (m *Manager) AddNode(id string, addr string) error {
//	return m.Compute.Add(id, addr)
//}
//
//func (m *Manager) AddGateway(id string, port int) error {
//	return m.Gateways.Add(id, port)
//}
//
//func (m *Manager) RemoveGateway(id string) error {
//	return m.Gateways.Remove(id)
//}
//
//func (m *Manager) AddGatewayEndpoint(gatewayId string, method string, relPath string, taskId string) error {
//	return m.Gateways.AddEndpoint(gatewayId, method, relPath, taskId)
//}
//
//func (m *Manager) AddTask(id string, handler string) error {
//	return m.Tasks.Add(id, handler)
//}
//
//func (m *Manager) RemoveTask(id string) error {
//	return m.Tasks.Remove(id)
//}
//
