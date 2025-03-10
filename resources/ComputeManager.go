package resources

import (
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

func (m *ComputeManager) Add(id ledger.ResourceId, addr string) {
	m.Instances[id] = ComputeConfig{addr}
}