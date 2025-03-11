package resources

import (
	"errors"
	"ows/ledger"
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

func (m *ComputeManager) Add(id ledger.ResourceId, addr string) error {
	sId := ledger.StringifyResourceId(id)

	if _, ok := m.Instances[sId]; ok {
		return errors.New("resource added before")
	}

	m.Instances[sId] = ComputeConfig{addr}

	return nil
}