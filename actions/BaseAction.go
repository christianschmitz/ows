package actions

type BaseAction struct {
}

func (a *BaseAction) GetAddedNodes() []string {
	return []string{}
}

func (a *BaseAction) GetRemovedNodes() []string {
	return []string{}
}
