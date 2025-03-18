package ledger

import (
	"fmt"
)

// Snapshot is used to validate a ledger.
type Snapshot struct {
	Version   LedgerVersion
	Head      ChangeSetID
	Functions map[FunctionID]FunctionConfig
	Gateways  map[GatewayID]GatewayConfig
	Nodes     map[NodeID]NodeConfig
	Policies  map[PolicyID]Policy
	Users     map[UserID]UserConfig
}

func newSnapshot(v LedgerVersion) *Snapshot {
	return &Snapshot{
		Version:   v,
		Head:      ChangeSetID(""),
		Functions: map[FunctionID]FunctionConfig{},
		Gateways:  map[GatewayID]GatewayConfig{},
		Nodes:     map[NodeID]NodeConfig{},
		Policies:  map[PolicyID]Policy{},
		Users:     map[UserID]UserConfig{},
	}
}

func (s *Snapshot) AddFunction(id FunctionID, config FunctionConfig) error {
	if _, ok := s.Functions[id]; ok {
		return fmt.Errorf("function resource %s already exists", id)
	}

	s.Functions[id] = config

	return nil
}

func (s *Snapshot) RemoveFunction(id FunctionID) error {
	if _, ok := s.Functions[id]; !ok {
		return fmt.Errorf("function %s doesn't exist", id)
	}

	delete(s.Functions, id)

	return nil
}

func (s *Snapshot) AddGateway(id GatewayID, config GatewayConfig) error {
	if _, ok := s.Gateways[id]; ok {
		return fmt.Errorf("gateway resource %s already exists", id)
	}

	if id, ok := s.Ports()[config.Port]; ok {
		return fmt.Errorf("port %d already used by %s", config.Port, id)
	}

	s.Gateways[id] = config

	return nil
}

func (s *Snapshot) AddGatewayEndpoint(id GatewayID, config GatewayEndpointConfig) error {
	gatewayConfig, ok := s.Gateways[id]
	if !ok {
		return fmt.Errorf("gateway %s doesn't exist", id)
	}

	if _, ok := s.Functions[config.FunctionID]; !ok {
		return fmt.Errorf("function %s doesn't exist", config.FunctionID)
	}

	for _, ep := range gatewayConfig.Endpoints {
		if ep.Method == config.Method && ep.Path == config.Path {
			return fmt.Errorf("duplicate endpoint for gateway %s (method=%s, path=%s)", id, config.Method, config.Path)
		}
	}

	gatewayConfig.Endpoints = append(gatewayConfig.Endpoints, config)

	return nil
}

func (s *Snapshot) RemoveGateway(id GatewayID) error {
	if _, ok := s.Gateways[id]; !ok {
		return fmt.Errorf("gateway %s doesn't exist", id)
	}

	delete(s.Gateways, id)

	return nil
}

func (s *Snapshot) RemoveGatewayEndpoint(id GatewayID, method string, path string) error {
	conf, ok := s.Gateways[id]
	if !ok {
		return fmt.Errorf("gateway %s doesn't exist", id)
	}

	endpoints := make([]GatewayEndpointConfig, 0)

	found := false
	for _, ep := range conf.Endpoints {
		if ep.Method == method && ep.Path == path {
			found = true
		} else {
			endpoints = append(endpoints, ep)
		}
	}

	if !found {
		return fmt.Errorf("gateway endpoint %s %s of %s doesn't exist", method, path, id)
	}

	conf.Endpoints = endpoints

	return nil
}

func (s *Snapshot) AddNode(id NodeID, config NodeConfig) error {
	if _, ok := s.Nodes[id]; ok {
		return fmt.Errorf("node %s already exists", id)
	}

	s.Nodes[id] = config

	return nil
}

func (s *Snapshot) RemoveNode(id NodeID) error {
	if _, ok := s.Nodes[id]; !ok {
		return fmt.Errorf("node %s doesn't exist", id)
	}

	delete(s.Nodes, id)

	return nil
}

func (s *Snapshot) AddUser(id UserID, config UserConfig) error {
	if _, ok := s.Users[id]; ok {
		return fmt.Errorf("user %s already exists", id)
	}

	s.Users[id] = config

	return nil
}

func (s *Snapshot) RemoveUser(id UserID) error {
	conf, ok := s.Users[id]
	if !ok {
		return fmt.Errorf("user %s doesn't exist", id)
	}

	if conf.IsRoot {
		return fmt.Errorf("can't remove root user %s", id)
	}

	delete(s.Users, id)

	return nil
}

func (s *Snapshot) addRootUsers(users ...PublicKey) {
	for _, user := range users {
		id := user.UserID()

		if err := s.AddUser(id, UserConfig{
			IsRoot:   true,
			Policies: []PolicyID{},
		}); err != nil {
			panic(fmt.Sprintf("unable to add root user (%v)", err))
		}
	}
}

func (s *Snapshot) Ports() map[Port]ResourceID {
	ports := map[Port]ResourceID{}

	for id, gateway := range s.Gateways {
		ports[gateway.Port] = id
	}

	for id, node := range s.Nodes {
		ports[node.GossipPort] = id
		ports[node.APIPort] = id
	}

	return ports
}

func (s *Snapshot) UserPolicies(users []PublicKey) ([]*Policy, error) {
	policies := []*Policy{}

	for _, key := range users {
		id := key.UserID()

		if conf, ok := s.Users[id]; ok {
			if conf.IsRoot {
				// Root users can never be locked out by Deny statements, so
				// the root policy is immediately returend.
				policies = append(policies, RootPolicy)
			} else {
				for _, policyID := range conf.Policies {
					policy, ok := s.Policies[policyID]
					if !ok {
						return nil, fmt.Errorf("policy %s not found", id)
					}

					policies = append(policies, &policy)
				}
			}
		}
	}

	return policies, nil
}
