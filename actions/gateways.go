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
	id := gen()

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

type AddGatewayEndpoint struct {
	BaseAction
	GatewayId ledger.ResourceId `cbor:"0,keyasint"`
	Method string `cbor:"1,keyasint"`
	Path string `cbor:"2,keyasint"`
	Task ledger.ResourceId `cbor:"3,keyasint"`
}

func NewAddGatewayEndpoint(gatewayId ledger.ResourceId, method string, path string, task ledger.ResourceId) *AddGatewayEndpoint {
	return &AddGatewayEndpoint{
		BaseAction{},
		gatewayId,
		method,
		path,
		task,
	}
}

func (a *AddGatewayEndpoint) Apply(m ledger.ResourceManager, _ ledger.ResourceIdGenerator) error {
	return m.AddGatewayEndpoint(a.GatewayId, a.Method, a.Path, a.Task)
}

func (a *AddGatewayEndpoint) GetName() string {
	return "AddGatewayEndpoint"
}

func (a *AddGatewayEndpoint) GetCategory() string {
	return "gateway"
}

func (a *AddGatewayEndpoint) GetResources() []ledger.ResourceId {
	return []ledger.ResourceId{a.GatewayId}
}

var _AddGatewayEndpointRegistered = ledger.RegisterAction("gateway", "AddGatewayEndpoint", func (attr []byte) (ledger.Action, error) {
	var a AddGatewayEndpoint
	err := cbor.Unmarshal(attr, &a)
	return &a, err
})