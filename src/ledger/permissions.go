package ledger

import (
	"fmt"
	"strings"
)

// Resource addition actions don't operate on existing resources. Policies must
// use wildcard in the resources list to allow such actions. The easiest way to
// enforce this is to specify a global resource id which is equal to the
// wildcard character.
const GlobalResourceID = ResourceID("*")

// Just like AWS policies, OWS policies consist of lists of policy statements.
// An action is allowed if it is explicitly included in the set of all Allowed
// actions minus the set of all Denied actions.
// This means that an action that has been Denied, can't be overridden by 
// Allow.
type Policy struct {
	Statements []PolicyStatement
}

type PolicyStatement struct {
	Actions   []string // "*" or "<category>:*" or "<category>:<action-name>"
	Resources []string // "*" or "resource..."
	Effect    string   // "Allow" or "Deny"
}

// A policy statement that allows all actions
var RootPolicyStatement = &PolicyStatement{
	Actions:   []string{"*"},
	Resources: []string{"*"},
	Effect:    "Allow",
}

// A policy that allows all actions
var RootPolicy = &Policy{
	[]PolicyStatement{
		*RootPolicyStatement,
	},
}

// If multiple resources are specified, all must be allowed.
func (p *Policy) Allows(category string, action string, resources ...ResourceID) bool {
	if len(resources) == 0 {
		panic(fmt.Sprintf("no resources specified"))
	} else if len(resources) == 1 {
		return p.allows(category, action, resources[0])
	} else {
		for _, r := range resources {
			if !p.allows(category, action, r) {
				return false
			}
		}

		return true
	}
}

func (p *Policy) allows(category string, action string, resource ResourceID) bool {
	allowed := false

	for _, s := range p.Statements {
		if s.Allows(category, action, resource) {
			allowed = true
		}

		if s.Denies(category, action, resource) {
			return false
		}
	}

	return allowed
}

func (s *PolicyStatement) Allows(category string, action string, resource ResourceID) bool {
	if s.Effect == "Allow" {
		return s.matches(category, action, resource)
	} else {
		return false
	}
}

func (s *PolicyStatement) Denies(category string, action string, resource ResourceID) bool {
	if s.Effect == "Deny" {
		return s.matches(category, action, resource)
	} else {
		return false
	}
}

func (s *PolicyStatement) matches(category string, action string, resource ResourceID) bool {
	return s.matchesResource(resource) && s.matchesCategory(category) && s.matchesAction(action)
}

func (s *PolicyStatement) matchesCategory(category string) bool {
	for _, a := range s.Actions {
		if a == "*" {
			return true
		}

		fields := strings.Split(a, ":")

		if len(fields) == 2 && (fields[0] == "*" || fields[0] == category) {
			return true
		}
	}

	return false
}

func (s *PolicyStatement) matchesAction(name string) bool {
	for _, a := range s.Actions {
		if a == "*" {
			return true
		}

		fields := strings.Split(a, ":")

		if len(fields) == 2 && (fields[1] == "*" || fields[1] == name) {
			return true
		}
	}

	return false
}

func (s *PolicyStatement) matchesResource(resourceId ResourceID) bool {
	for _, r := range s.Resources {
		if r == "*" {
			return true
		} else if r == string(resourceId) {
			return true
		}
	}

	return false
}

func actionAllowed(action Action, policies ...*Policy) bool {	
	for _, policy := range policies {
		if policy.Allows(action.Category(), action.Name(), action.Resources()...) {
			return true
		}
	}

	return false
}