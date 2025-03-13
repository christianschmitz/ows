package ledger

import (
	"strings"
)

type PolicyStatement struct {
	Resources []string // "*" or "resource..."
	Actions []string // "*" or "<category>:*" or "<category>:<action-name>"
	Effect string // "Allow" or "Deny"
}

func GenerateRootPolicyStatement() PolicyStatement {
	return PolicyStatement{
		Resources: []string{"*"},
		Actions: []string{"*"},
		Effect: "Allow",
	}
}

func (s *PolicyStatement) Allows(resource string, category string, action string) bool {
	if (s.Effect == "Allow") {
		return s.matches(resource, category, action)
	} else {
		return false
	}
}

func (s *PolicyStatement) Denies(resource string, category string, action string) bool {
	if (s.Effect == "Deny") {
		return s.matches(resource, category, action)
	} else {
		return false
	}
}

func (s *PolicyStatement) matches(resource string, category string, action string) bool {
	return s.matchesResource(resource) && s.matchesCategory(category) && s.matchesAction(action)
}

func (s *PolicyStatement) matchesResource(resourceId string) bool {
	for _, r := range s.Resources {
		if (r == "*") {
			return true
		} else if r == resourceId {
			return true
		}
	}

	return false
}

func (s *PolicyStatement) matchesCategory(category string) bool {
	for _, a := range s.Actions{
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
	for _, a := range s.Actions{
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