package resources

import (
	"ows/ledger"
)

type ResourceManager struct {
	Compute *ComputeManager
	Gateways *GatewaysManager
	Tasks *TasksManager
}

func NewResourceManager() *ResourceManager {
	tasks := NewTasksManager()
	return &ResourceManager{
		NewComputeManager(),
		NewGatewaysManager(tasks),
		tasks,
	}
}

func (m *ResourceManager) AddCompute(id ledger.ResourceId, addr string) error {
	return m.Compute.Add(id, addr)
}

func (m *ResourceManager) AddGateway(id ledger.ResourceId, port int) error {
	return m.Gateways.Add(id, port)
}

func (m *ResourceManager) AddGatewayEndpoint(id ledger.ResourceId, method string, relPath string, task ledger.ResourceId) error {
	return m.Gateways.AddEndpoint(id, method, relPath, task)
}

func (m *ResourceManager) AddTask(id ledger.ResourceId, handler ledger.AssetId) error {
	return m.Tasks.Add(id, handler)
}