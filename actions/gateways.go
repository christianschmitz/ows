package actions

import (
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

type AddGateway struct {
	BaseAction
	Port int `cbor:"0,keyasint"`
}

func NewAddGateway(port int) *AddGateway {
	return &AddGateway{BaseAction{}, port}
}

func (a *AddGateway) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	id := gen("gateway")

	return m.AddGateway(id, a.Port)
}

func (a *AddGateway) GetName() string {
	return "AddGateway"
}

func (a *AddGateway) GetCategory() string {
	return "gateway"
}

var _AddGatewayRegistered = ledger.RegisterAction("gateway", "AddGateway", func (attr []byte) (ledger.Action, error) {
	var a AddGateway
	err := cbor.Unmarshal(attr, &a)
	return &a, err
})

type RemoveGateway struct {
	BaseAction
	Id string `cbor:"0,keyasint"`
}

func NewRemoveGateway(id string) *RemoveGateway {
	return &RemoveGateway{BaseAction{}, id}
}

func (a *RemoveGateway) Apply(m ledger.ResourceManager, _ ledger.ResourceIdGenerator) error {
	return m.RemoveGateway(a.Id)
}

func (a *RemoveGateway) GetName() string {
	return "RemoveGateway"
}

func (a *RemoveGateway) GetCategory() string {
	return "gateway"
}

func (a *RemoveGateway) GetResources() []string {
	return []string{a.Id}
}

var _RemoveGatewayRegistered = ledger.RegisterAction("gateway", "RemoveGateway", func (attr []byte) (ledger.Action, error) {
	var a RemoveGateway
	err := cbor.Unmarshal(attr, &a)
	return &a, err
})

type AddGatewayEndpoint struct {
	BaseAction
	GatewayId string `cbor:"0,keyasint"`
	Method string `cbor:"1,keyasint"`
	Path string `cbor:"2,keyasint"`
	TaskId string `cbor:"3,keyasint"`
}

func NewAddGatewayEndpoint(gatewayId string, method string, path string, taskId string) *AddGatewayEndpoint {
	return &AddGatewayEndpoint{
		BaseAction{},
		gatewayId,
		method,
		path,
		taskId,
	}
}

func (a *AddGatewayEndpoint) Apply(m ledger.ResourceManager, _ ledger.ResourceIdGenerator) error {
	return m.AddGatewayEndpoint(a.GatewayId, a.Method, a.Path, a.TaskId)
}

func (a *AddGatewayEndpoint) GetName() string {
	return "AddGatewayEndpoint"
}

func (a *AddGatewayEndpoint) GetCategory() string {
	return "gateway"
}

func (a *AddGatewayEndpoint) GetResources() []string {
	return []string{a.GatewayId}
}

var _AddGatewayEndpointRegistered = ledger.RegisterAction("gateway", "AddGatewayEndpoint", func (attr []byte) (ledger.Action, error) {
	var a AddGatewayEndpoint
	err := cbor.Unmarshal(attr, &a)
	return &a, err
})

type GatewayConfig struct {
	Port int
}

func ListGateways(l *ledger.Ledger) map[string]GatewayConfig {
	m := map[string]GatewayConfig{}

	for _, cs := range l.Changes {
		for i, action := range cs.Actions {
			if ag, ok := action.(*AddGateway); ok {
				id := ledger.GenerateResourceId("gateway", cs.Parent[:], i)
				m[id] = GatewayConfig{ag.Port}
			} else if rg, ok := action.(*RemoveGateway); ok {
				delete(m, rg.Id)
			}
		}
	}
	
	return m
}
