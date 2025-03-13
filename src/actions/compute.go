package actions

import (
	"ows/ledger"
)

func ListNodes(l *ledger.Ledger) map[string]string {
	m := map[string]string{}

	for _, cs := range l.Changes {
		for i, a := range cs.Actions {
			if ac, ok := a.(*AddNode); ok {
				id := ledger.GenerateResourceId("node", cs.Parent[:], i)
				m[id] = ac.Address
			} else if rc, ok := a.(*RemoveNode); ok {
				delete(m, rc.Id)
			}
		}
	}

	return m
}
