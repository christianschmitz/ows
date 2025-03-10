package resources

import (
	"cws/ledger"
)

type TaskConfig struct {
	Runtime string
	Handler string
}

type TasksManager struct {
	Tasks map[ledger.ResourceId]TaskConfig
}

func NewTasksManager() *TasksManager {
	return &TasksManager{
		map[ledger.ResourceId]TaskConfig{},
	}
}

func (m *TasksManager) Add(id ledger.ResourceId, handler string) {
	m.Tasks[id] = TaskConfig{"nodejs", handler}
}