package ledger

import (
	"bytes"
	"errors"
	"strconv"	
)

type computeValidationConfig = bool

type userValidationConfig struct {
	isRoot bool

	// each policy is indepedent and doesn't impact other policies in this list
	// so
	policies []string 
}

type taskValidationConfig = bool

type policyValidationConfig struct {
	statements []string
}

type gatewayValidationConfig struct {
	port int
}

// implements ResourceManager interface
// more strict than resources.ResourceManager
type ValidationContext struct {
	validateAssets bool
	compute	map[string]computeValidationConfig
	users map[string]userValidationConfig
	tasks map[string]taskValidationConfig
	policyStatements map[string]PolicyStatement
	policies map[string]policyValidationConfig
	gateways map[string]gatewayValidationConfig
}

func newValidationContext(validateAssets bool, rootUsers []PubKey) *ValidationContext {
	users := map[string]userValidationConfig{}

	for _, ru := range rootUsers {
		users[StringifyPubKey(ru)] = userValidationConfig{
			isRoot: true,
			policies: []string{},
		}
	}

	return &ValidationContext{
		validateAssets: validateAssets,
		compute: map[string]computeValidationConfig{},
		users: users,
		tasks: map[string]taskValidationConfig{},
		policyStatements: map[string]PolicyStatement{},
		policies: map[string]policyValidationConfig{},
		gateways: map[string]gatewayValidationConfig{},
	}
}

func (l *Ledger) ValidateAll(validateAssets bool) error {
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
	context := newValidationContext(validateAssets, rootUsers)

	// replay all the changes

	head := []byte{}

	for i, c := range l.Changes {
		// check that the Parent corresponds
		if !bytes.Equal(c.Parent, head) {
			return errors.New("Invalid change set head, expected " + StringifyChangeSetHash(head) + ", got " + StringifyChangeSetHash(c.Parent))
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

func (c *ValidationContext) getPolicy(id string) (*Policy, error) {
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
		if conf, ok := c.users[StringifyPubKey(pk)]; ok {
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
	sId := StringifyResourceId(id)
	if _, ok := c.compute[sId]; ok {
		return errors.New("compute resource already exists")
	}

	c.compute[sId] = true

	return nil
}

func (c *ValidationContext) AddTask(id ResourceId, handler AssetId) error {
	sId := StringifyResourceId(id)

	if _, ok := c.tasks[sId]; ok {
		return errors.New("task resource already exists")
	}

	if (c.validateAssets) {
		if ok := AssetExists(handler); !ok {
			return errors.New("handler asset " + StringifyAssetId(handler) + " not found")
		}
	}

	c.tasks[sId] = true
	
	return nil
}

func (c *ValidationContext) AddGateway(id ResourceId, port int) error {
	sId := StringifyResourceId(id)

	if _, ok := c.gateways[sId]; ok {
		return errors.New("gateway resource already exists")
	}

	if port == SYNC_PORT {
		return errors.New("can't use port " + strconv.Itoa(port))
	}

	for _, other := range c.gateways {
		if (other.port == port) {
			return errors.New("port " + strconv.Itoa(port) + " already used by other gateway")
		}
	}

	c.gateways[sId] = gatewayValidationConfig{port}

	return nil
}

func (c *ValidationContext) AddGatewayEndpoint(id ResourceId, method string, endpoint string, task ResourceId) error {
	sId := StringifyResourceId(id)

	if _, ok := c.gateways[sId]; !ok {
		return errors.New("gateway " + sId + " not found")
	}

	if _, ok := c.tasks[StringifyResourceId(task)]; !ok {
		return errors.New("endpoint task not found")
	}

	// TODO: check that endpoints aren't duplicated

	return nil
}