package actions

import (
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

// adds a node
type AddNode struct {
	BaseAction
	Key ledger.PubKey `cbor:"0,keyasint"`
	Address string `cbor:"1,keyasint"`
}

func NewAddNode(addr string) *AddNode {
	return &AddNode{BaseAction{}, [32]byte{},addr}
}

func GenerateNodeId(key ledger.PubKey) string {
	nodeIdBytes := ledger.DigestCompact(key[:])

	return ledger.StringifyHumanReadableBytes("node", nodeIdBytes)
}

func (a *AddNode) Apply(m ledger.ResourceManager, _ ledger.ResourceIdGenerator) error {
	id := GenerateNodeId(a.Key)

	return m.AddNode(id, a.Address)
}

func (a *AddNode) GetName() string {
	return "AddNode"
}

func (a *AddNode) GetCategory() string {
	return "compute"
}

var _AddNodeRegistered = ledger.RegisterAction("compute", "AddNode", func (attr []byte) (ledger.Action, error) {
	var a AddNode
	err := cbor.Unmarshal(attr, &a)
	return &a, err
})
