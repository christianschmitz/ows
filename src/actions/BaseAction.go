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

func (a *BaseAction) GetResources() []string {
	return []string{ledger.GenerateGlobalResourceId()}
}
