package actions

import (
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

type AddTask struct {
	BaseAction
	Runtime string `cbor:"0,keyasint"`
	Handler string `cbor:"1,keyasint"`
}

func NewAddTask(runtime string, handler string) *AddTask {
	return &AddTask{BaseAction{}, runtime, handler}
}

func (a *AddTask) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) error {
	id := gen("task")

	return m.AddTask(id, a.Handler)
}

func (a *AddTask) GetName() string {
	return "AddTask"
}

func (a *AddTask) GetCategory() string {
	return "tasks"
}

var _AddTaskRegistered = ledger.RegisterAction("tasks", "AddTask", func (attr []byte) (ledger.Action, error) {
	var action AddTask
	err := cbor.Unmarshal(attr, &action)
	return &action, err
})

type RemoveTask struct {
	BaseAction
	Id string `cbor:"0,keyasint"`
}

func NewRemoveTask(id string) *RemoveTask {
	return &RemoveTask{BaseAction{}, id}
}

func (a *RemoveTask) Apply(m ledger.ResourceManager, _ ledger.ResourceIdGenerator) error {
	return m.RemoveTask(a.Id)
}

func (a *RemoveTask) GetName() string {
	return "RemoveTask"
}

func (a *RemoveTask) GetCategory() string {
	return "tasks"
}

var _RemoveTaskRegisterd = ledger.RegisterAction("tasks", "RemoveTask", func (attr []byte) (ledger.Action, error) {
	var action RemoveTask
	err := cbor.Unmarshal(attr, &action)
	return &action, err
})

type TaskConfig struct {
	Runtime string
	Handler string
}

func ListTasks(l *ledger.Ledger) map[string]TaskConfig {
	m := map[string]TaskConfig{}

	for _, cs := range l.Changes {
		for i, action := range cs.Actions {
			if at, ok := action.(*AddTask); ok {
				id := ledger.GenerateResourceId("task", cs.Parent[:], i)
				m[id] = TaskConfig{at.Runtime, at.Handler}
			} else if rt, ok := action.(*RemoveTask); ok {
				delete(m, rt.Id)
			}
		}
	}

	return m
}