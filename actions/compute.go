package actions

import (
	"log"
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

// adds a node
type AddCompute struct {
	Address string `cbor:"0,keyasint"`
}

func NewAddCompute(addr string) *AddCompute {
	return &AddCompute{addr}
}

func (c *AddCompute) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) {
	id := gen()

	m.AddCompute(id, c.Address)
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
	Id ledger.ResourceId `cbor:"0,keyasint"`
	Address string `cbor:"1,keyasint"`
}

func (c *RemoveCompute) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) {
	log.Fatal("not yet implemented")
}

func (c *RemoveCompute) GetName() string {
	return "RemoveCompute"
}

func (c *RemoveCompute) GetCategory() string {
	return "compute"
}

var _RemoveComputeRegistered = ledger.RegisterAction("compute", "RemoveCompute", func (attr []byte) (ledger.Action, error) {
	var c RemoveCompute
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})

func GetNodeAddresses(l *ledger.Ledger) map[ledger.ResourceId]string {
	m := map[ledger.ResourceId]string{}

	for _, cs := range l.Changes {
		for i, a := range cs.Actions {
			if ac, ok := a.(*AddCompute); ok {
				id := ledger.GenerateResourceId(cs.Parent[:], i)
				m[id] = ac.Address
			} else if rc, ok := a.(*RemoveCompute); ok {
				delete(m, rc.Id)
			}
		}
	}

	return m
}