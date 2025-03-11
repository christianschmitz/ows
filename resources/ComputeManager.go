package resources

import (
	"errors"
	"ows/ledger"
)

type ComputeConfig struct {
	Address string
}

type ComputeManager struct {
	Instances map[ledger.ResourceId]ComputeConfig
}

func NewComputeManager() *ComputeManager {
	return &ComputeManager{
		map[ledger.ResourceId]ComputeConfig{},
	}
}

func (m *ComputeManager) Add(id ledger.ResourceId, addr string) error {
	if _, ok := m.Instances[id]; ok {
		return errors.New("resource added before")
	}

	m.Instances[id] = ComputeConfig{addr}

	return nil
}