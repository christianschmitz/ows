package resources

import (
	"ows/ledger"
)

type ResourceManager struct {
	Compute *ComputeManager
	Tasks *TasksManager
}

func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		NewComputeManager(),
		NewTasksManager(),
	}
}

func (m *ResourceManager) AddCompute(id ledger.ResourceId, addr string) error {
	return m.Compute.Add(id, addr)
}

func (m *ResourceManager) AddTask(id ledger.ResourceId, handler string) error {
	return m.Tasks.Add(id, handler)
}