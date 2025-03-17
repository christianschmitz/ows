package ledger

import ()

const (
	FunctionsCategory = "functions"
	AddFunctionName = "Add"
	RemoveFunctionName = "Remove"
)

// Currently, the only supported function runtime is "nodejs".
type AddFunction struct {
	Runtime   string `cbor:"0,keyasint"`
	HandlerID AssetID `cbor:"1,keyasint"`
}

func (a AddFunction) Category() string {
	return FunctionsCategory
}

func (a AddFunction) Name() string {
	return AddFunctionName
}

func (a AddFunction) Resources() []ResourceID {
	return []ResourceID{GlobalResourceID}
}

func (a AddFunction) Apply(s *Snapshot, genID ResourceIDGenerator) error {
	id := genID(FunctionIDPrefix)

	return s.AddFunction(id, FunctionConfig{
		Runtime: a.Runtime,
		HandlerID: a.HandlerID,
	})
}

type RemoveFunction struct {
	ID ResourceID `cbor:"0,keyasint"`
}

func (a RemoveFunction) Category() string {
	return FunctionsCategory
}

func (a RemoveFunction) Name() string {
	return RemoveFunctionName
}

func (a RemoveFunction) Resources() []ResourceID {
	return []ResourceID{a.ID}
}

func (a RemoveFunction) Apply(s *Snapshot, _ ResourceIDGenerator) error {
	return s.RemoveFunction(a.ID)
}

const (
	GatewaysCategory = "gateways"
	AddGatewayName = "Add"
	AddGatewayEndpointName = "AddEndpoint"
	RemoveGatewayName = "Remove"
)

type AddGateway struct {
	Port Port `cbor:"0,keyasint"`
}

func (a AddGateway) Category() string {
	return GatewaysCategory
}

func (a AddGateway) Name() string {
	return AddGatewayName
}

func (a AddGateway) Resources() []ResourceID {
	return []ResourceID{GlobalResourceID}
}

func (a AddGateway) Apply(s *Snapshot, genID ResourceIDGenerator) error {
	id := genID(GatewayIDPrefix)

	return s.AddGateway(id, GatewayConfig{
		Port: a.Port,
		Endpoints: []GatewayEndpointConfig{},
	})
}

// Valid methods are "GET", "POST", "PUT", "PATCH", or "DELETE".
// FunctionID refers to the handler that will be invoked when the endpoint is
// requested.
type AddGatewayEndpoint struct {
	GatewayID ResourceID `cbor:"0,keyasint"`
	Method    string `cbor:"1,keyasint"`
	Path      string `cbor:"2,keyasint"`
	FunctionID ResourceID `cbor:"3,keyasint"`
}

func (a AddGatewayEndpoint) Category() string {
	return GatewaysCategory
}

func (a AddGatewayEndpoint) Name() string {
	return AddGatewayEndpointName
}

func (a AddGatewayEndpoint) Resources() []ResourceID {
	// Permissions related to the invocation of a.FunctionID are attached to
	// the resource id of the endpoint, not to the user who created the 
	// endpoint.
	// TODO: when invoking the function without authorization, should a 
	// permission error be thrown during runtime? or is it better the thrown an
	// error during change set validation?
	return []ResourceID{a.GatewayID}
}

func (a AddGatewayEndpoint) Apply(s *Snapshot, _ ResourceIDGenerator) error {
	return s.AddGatewayEndpoint(a.GatewayID, GatewayEndpointConfig{
		Method: a.Method,
		Path: a.Path,
		FunctionID: a.FunctionID,
	})
}

type RemoveGateway struct {
	ID GatewayID `cbor:"0,keyasint"`
}

func (a RemoveGateway) Category() string {
	return GatewaysCategory
}

func (a RemoveGateway) Name() string {
	return RemoveGatewayName
}

func (a RemoveGateway) Resources() []ResourceID {
	return []ResourceID{a.ID}
}

func (a RemoveGateway) Apply(s *Snapshot, _ ResourceIDGenerator) error {
	return s.RemoveGateway(a.ID)
}

const (
	NodesCategory = "nodes"
	AddNodeName = "Add"
	RemoveNodeName = "Remove"
)

// When applied, creates a node with all the properties in NodeConfig.
// This doesn't wrap NodeConfig though so that codec format changes can be
// hidden behind `AddNodeV1`, `AddNodeV2` etc.
//
// The node ResourceID is derived from its public key.
type AddNode struct {
	Key        PublicKey `cbor:"0,keyasint"`
	Address    string    `cbor:"1,keyasint"`
	GossipPort Port      `cbor:"2,keyasint"`
	SyncPort   Port      `cbor:"3,keyasint"`
}

func (a AddNode) Category() string {
	return NodesCategory
}

func (a AddNode) Name() string {
	return AddNodeName
}

func (a AddNode) Resources() []ResourceID {
	return []ResourceID{GlobalResourceID}
}

func (a AddNode) Apply(s *Snapshot, _ ResourceIDGenerator) error {
	id := a.Key.NodeID()

	return s.AddNode(id, NodeConfig{
		Key: a.Key,
		Address: a.Address,
		GossipPort: a.GossipPort,
		SyncPort: a.SyncPort,
	})
}

type RemoveNode struct {
	ID ResourceID `cbor:"0,keyasint"`
}

func (a RemoveNode) Category() string {
	return NodesCategory
}

func (a RemoveNode) Name() string {
	return RemoveNodeName
}

func (a RemoveNode) Resources() []ResourceID {
	return []ResourceID{a.ID}
}

func (a RemoveNode) Apply(s *Snapshot, _ ResourceIDGenerator) error {
	return s.RemoveNode(a.ID)
}

const (
	PermissionsCategory = "permissions"
	AddUserName = "AddUser"
)

type AddUser struct {
	Key PublicKey `cbor:"0,keyasint"`
}

func (a AddUser) Category() string {
	return PermissionsCategory
}

func (a AddUser) Name() string {
	return AddUserName
}

func (a AddUser) Resources() []ResourceID {
	return []ResourceID{GlobalResourceID}
}

func (a AddUser) Apply(s *Snapshot, _ ResourceIDGenerator) error {
	id := a.Key.UserID()

	return s.AddUser(id, UserConfig{
		Key: a.Key,
		Policies: []ResourceID{},
	})
}