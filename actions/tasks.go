package actions

import (
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

type AddTask struct {
	BaseAction
	Runtime string `cbor:"0,keyasint"`
	Handler ledger.AssetId `cbor:"1,keyasint"`
}

func NewAddTask(runtime string, handler ledger.AssetId) *AddTask {
	return &AddTask{BaseAction{}, runtime, handler}
}

func (c *AddTask) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	id := gen()

	return m.AddTask(id, c.Handler)
}

func (c *AddTask) GetName() string {
	return "AddTask"
}

func (c *AddTask) GetCategory() string {
	return "tasks"
}

var _AddTaskRegistered = ledger.RegisterAction("tasks", "AddTask", func (attr []byte) (ledger.Action, error) {
	var c AddTask
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})