package resources

import (
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

func (m *ResourceManager) AddNode(id string, addr string) error {
	return m.Compute.Add(id, addr)
}

func (m *ResourceManager) AddGateway(id string, port int) error {
	return m.Gateways.Add(id, port)
}

func (m *ResourceManager) RemoveGateway(id string) error {
	return m.Gateways.Remove(id)
}

func (m *ResourceManager) AddGatewayEndpoint(gatewayId string, method string, relPath string, taskId string) error {
	return m.Gateways.AddEndpoint(gatewayId, method, relPath, taskId)
}

func (m *ResourceManager) AddTask(id string, handler string) error {
	return m.Tasks.Add(id, handler)
}

func (m *ResourceManager) RemoveTask(id string) error {
	return m.Tasks.Remove(id)
}