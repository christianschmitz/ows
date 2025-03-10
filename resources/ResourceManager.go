package resources

import (
	"cws/ledger"
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

func (m *ResourceManager) AddCompute(id ledger.ResourceId, addr string) {
	m.Compute.Add(id, addr)
}

func (m *ResourceManager) AddTask(id ledger.ResourceId, handler string) {
	m.Tasks.Add(id, handler)
}