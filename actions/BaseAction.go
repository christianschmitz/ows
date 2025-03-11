package actions

import (
	"ows/ledger"
)

type BaseAction struct {
}

func (a *BaseAction) GetAddedNodes() []string {
	return []string{}
}

func (a *BaseAction) GetRemovedNodes() []string {
	return []string{}
}

func (a *BaseAction) GetResources() []ledger.ResourceId {
	return []ledger.ResourceId{ledger.GenerateGlobalResourceId()}
}