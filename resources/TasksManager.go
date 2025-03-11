package resources

import (
	"errors"
	"ows/ledger"
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

func (m *TasksManager) Add(id ledger.ResourceId, handler string) error {
	if _, ok := m.Tasks[id]; ok {
		return errors.New("task added before")
	}

	m.Tasks[id] = TaskConfig{"nodejs", handler}

	return nil
}