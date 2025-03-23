package resources

import (
	"net/http"

	"ows/ledger"
)

type Manager struct {
	Current   *ledger.KeyPair
	AssetsDir string
	Functions map[ledger.FunctionID]*Function
	Gateways  map[ledger.GatewayID]*Gateway
	Nodes     map[ledger.NodeID]*Node

	portOffset        int
	dockerInitialized bool // TODO: this should be moved into an Images or Containers field
}

type Function struct {
	Config ledger.FunctionConfig
}

type Gateway struct {
	Port    ledger.Port
	Handler *GatewayHandler
	Server  *http.Server
}

type GatewayHandler struct {
	Manager *Manager // need access to manager to be able to run functions and fetch assets
	// first key is method: "GET", "POST", "DELETE", "PUT", "PATCH"
	// second key is relative path, including initial slash (eg. "/assets")
	Endpoints map[string]map[string]*GatewayEndpoint
}

type GatewayEndpoint struct {
	Config ledger.GatewayEndpointConfig
}

type Node struct {
	Config ledger.NodeConfig
}

func NewManager(current *ledger.KeyPair, assetsDir string, portOffset int) *Manager {
	return &Manager{
		Current:           current,
		AssetsDir:         assetsDir,
		Functions:         map[ledger.FunctionID]*Function{},
		Gateways:          map[ledger.GatewayID]*Gateway{},
		Nodes:             map[ledger.NodeID]*Node{},
		portOffset:        portOffset,
		dockerInitialized: false,
	}
}

func (m *Manager) Sync(snapshot *ledger.Snapshot) error {
	if err := m.SyncFunctions(snapshot.Functions); err != nil {
		return err
	}

	if err := m.SyncGateways(snapshot.Gateways); err != nil {
		return err
	}

	if err := m.SyncNodes(snapshot.Nodes); err != nil {
		return err
	}

	return nil
}
