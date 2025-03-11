package ledger

import (
	"bytes"
	"errors"
)

type computeValidationConfig = bool

type userValidationConfig struct {
	isRoot bool

	// each policy is indepedent and doesn't impact other policies in this list
	// so
	policies []ResourceId 
}

type taskValidationConfig = bool

type policyValidationConfig struct {
	statements []ResourceId
}

// implements ResourceManager interface
// more strict than resources.ResourceManager
type ValidationContext struct {
	compute	map[ResourceId]computeValidationConfig
	users map[PubKey]userValidationConfig
	tasks map[ResourceId]taskValidationConfig
	policyStatements map[ResourceId]PolicyStatement
	policies map[ResourceId]policyValidationConfig
}

func newValidationContext(rootUsers []PubKey) *ValidationContext {
	users := map[ResourceId]userValidationConfig{}

	for _, ru := range rootUsers {
		users[ru] = userValidationConfig{
			isRoot: true,
			policies: []ResourceId{},
		}
	}

	return &ValidationContext{
		compute: map[ResourceId]computeValidationConfig{},
		users: users,
		tasks: map[ResourceId]taskValidationConfig{},
		policyStatements: map[ResourceId]PolicyStatement{},
		policies: map[ResourceId]policyValidationConfig{},
	}
}

func (l *Ledger) ValidateAll() error {
	rootSignatures := l.Changes[0].Signatures[:]

	genesisBytes, err := l.Changes[0].Encode(true)
	if err != nil {
		return err
	}

	for _, s := range rootSignatures {
		if !s.Verify(genesisBytes) {
			return errors.New("invalid root signature")
		}
	}

	rootUsers := l.Changes[0].CollectSignatories()

	// create validation context
	context := newValidationContext(rootUsers)

	// replay all the changes

	head := []byte{}

	for i, c := range l.Changes {
		// check that the Parent corresponds
		if !bytes.Equal(c.Parent, head) {
			return errors.New("Invalid change set head")
		}

		// first validate that the signatures correspond
		signatories := []PubKey{}

		if i == 0 {
			signatories = rootUsers
		} else {
			for _, s := range c.Signatures {
				cbs, err := c.Encode(true)
				if err != nil {
					return err
				}

				if !s.Verify(cbs) {
					return errors.New("invalid change set signatures")
				}
			}

			signatories = c.CollectSignatories()
		}

		// check that all the actions can actually be taken by the signatories
		userPolicies, err := context.getSignatoryPermissions(signatories)
		if err != nil {
			return err
		}

		for _, a := range c.Actions {
			allowed := false
			for _, policy := range userPolicies {
				if policy.AllowsAll(a.GetResources(), a.GetCategory(), a.GetName()) {
					allowed = true
				}
			}

			if !allowed {
				return errors.New("merged policy of all signatories doesn't allow " + a.GetCategory() + ":" + a.GetName())
			}
		}

		if err := c.Apply(context); err != nil {
			return err
		}

		head = c.Hash()
	}

	return nil
}

func (c *ValidationContext) getPolicy(id ResourceId) (*Policy, error) {
	if conf, ok := c.policies[id]; ok {
		statements := []PolicyStatement{}

		for _, statementId := range conf.statements {
			if statement, ok := c.policyStatements[statementId]; ok {
				statements = append(statements, statement)
			} else {
				return nil, errors.New("policy statement not found")
			}
		}

		return &Policy{statements}, nil
	} else {
		return nil, errors.New("policy not found")
	}
}

func (c *ValidationContext) getSignatoryPermissions(signatories []PubKey) ([]*Policy, error) {
	policies := []*Policy{}

	for _, pk := range signatories {
		if conf, ok := c.users[pk]; ok {
			if conf.isRoot {
				// root users can never be locked out by Deny statements, so we immediately return root policy
				policies = append(policies, GenerateRootPolicy())
			} else {
				for _, policyId := range conf.policies {
					policy, err := c.getPolicy(policyId)
					if err != nil {
						return nil, err
					}

					policies = append(policies, policy)
				}
			}
		}
	}

	return policies, nil
}

func (c *ValidationContext) AddCompute(id ResourceId, _addr string) error {
	if _, ok := c.compute[id]; ok {
		return errors.New("compute resource already exists")
	}

	c.compute[id] = true

	return nil
}

func (c *ValidationContext) AddTask(id ResourceId, _handler string) error {
	if _, ok := c.tasks[id]; ok {
		return errors.New("task resource already exists")
	}

	c.tasks[id] = true
	
	return nil
}