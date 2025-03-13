package resources

import (
	"errors"
)

type ComputeConfig struct {
	Address string
}

type ComputeManager struct {
	Instances map[string]ComputeConfig
}

func NewComputeManager() *ComputeManager {
	return &ComputeManager{
		map[string]ComputeConfig{},
	}
}

func (m *ComputeManager) Add(id string, addr string) error {
	if _, ok := m.Instances[id]; ok {
		return errors.New("resource added before")
	}

	m.Instances[id] = ComputeConfig{addr}

	return nil
}
