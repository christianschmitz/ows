package ledger

import (
	"fmt"
)

type Policy struct {
	Statements []PolicyStatement
}

func GenerateRootPolicy() *Policy {
	return &Policy{
		[]PolicyStatement{
			GenerateRootPolicyStatement(),
		},
	}
}

func (p *Policy) Merge(other *Policy) *Policy {
	if len(p.Statements) == 0 {
		return other
	} else if len(other.Statements) == 0 {
		return p
	} else {
		statements := p.Statements[:]

		for _, os := range other.Statements {
			statements = append(statements, os)
		}
		return &Policy{statements}
	}
}

// everything in allow - everything in deny
// this means that something that has been denied can't be overridden with allow
func (p *Policy) Allows(resource string, category string, action string) bool {
	isAllowed := false

	for _, s := range p.Statements {
		if s.Allows(resource, category, action) {
			isAllowed = true
		}

		if s.Denies(resource, category, action) {
			return false
		}
	}

	return isAllowed
}

func (p *Policy) AllowsAll(resources []string, category string, action string) bool {
	if len(resources) == 0 {
		fmt.Println("resources list can't be empty in Policy.AllowsAll(), disallowing")
		return false
	}

	for _, r := range resources {
		if !p.Allows(r, category, action) {
			return false
		}
	}

	return true
}
