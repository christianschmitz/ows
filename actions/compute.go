package actions

import (
	"errors"
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

// adds a node
type AddCompute struct {
	BaseAction
	Address string `cbor:"0,keyasint"`
}

func NewAddCompute(addr string) *AddCompute {
	return &AddCompute{BaseAction{}, addr}
}

func (c *AddCompute) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	id := gen()

	return m.AddCompute(id, c.Address)
}

func (c *AddCompute) GetName() string {
	return "AddCompute"
}

func (c *AddCompute) GetCategory() string {
	return "compute"
}

var _AddComputeRegistered = ledger.RegisterAction("compute", "AddCompute", func (attr []byte) (ledger.Action, error) {
	var c AddCompute
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})

// removes a node
type RemoveCompute struct {
	BaseAction
	Id ledger.ResourceId `cbor:"0,keyasint"`
}

func NewRemoveCompute(id ledger.ResourceId) *RemoveCompute {
	return &RemoveCompute{BaseAction{}, id}
}

func (c *RemoveCompute) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	return errors.New("RemoveCompute.Apply() not yet implemented")
}

func (c *RemoveCompute) GetName() string {
	return "RemoveCompute"
}

func (c *RemoveCompute) GetCategory() string {
	return "compute"
}

func (c *RemoveCompute) GetResources() []ledger.ResourceId {
	return []ledger.ResourceId{c.Id}
}

var _RemoveComputeRegistered = ledger.RegisterAction("compute", "RemoveCompute", func (attr []byte) (ledger.Action, error) {
	var c RemoveCompute
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})

func GetNodeAddresses(l *ledger.Ledger) map[string]string {
	m := map[string]string{}

	for _, cs := range l.Changes {
		for i, a := range cs.Actions {
			if ac, ok := a.(*AddCompute); ok {
				id := ledger.GenerateResourceId(cs.Parent[:], i)
				m[ledger.StringifyResourceId(id)] = ac.Address
			} else if rc, ok := a.(*RemoveCompute); ok {
				delete(m, ledger.StringifyResourceId(rc.Id))
			}
		}
	}

	return m
}