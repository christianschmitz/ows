package actions

import (
	"errors"
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

type RemoveNode struct {
	BaseAction
	Id string `cbor:"0,keyasint"`
}

func NewRemoveNode(id string) *RemoveNode {
	return &RemoveNode{BaseAction{}, id}
}

func (c *RemoveNode) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	return errors.New("RemoveNode.Apply() not yet implemented")
}

func (c *RemoveNode) GetName() string {
	return "RemoveNode"
}

func (c *RemoveNode) GetCategory() string {
	return "compute"
}

func (c *RemoveNode) GetResources() []string {
	return []string{c.Id}
}

var _RemoveNodeRegistered = ledger.RegisterAction("compute", "RemoveNode", func (attr []byte) (ledger.Action, error) {
	var c RemoveNode
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})