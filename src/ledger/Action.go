package ledger

import (
	"errors"
	"log"
	"github.com/fxamacker/cbor/v2"
)

type ActionGenerator = func(attrBytes []byte) (Action, error)

var ACTIONS = map[string]map[string]ActionGenerator{}

type Action interface { 
	// valid categories are: compute, permissions
	GetCategory() string
	GetName() string
	GetResources() []string
	GetAddedNodes() []string
	GetRemovedNodes() []string

	Apply(m ResourceManager, gen ResourceIdGenerator) error
}

type ActionHelper struct {
	action Action
}

type ActionCbor struct {
	Category   string `cbor:"0,keyasint"`
	Name       string `cbor:"1,keyasint"`
	Attributes []byte `cbor:"2,keyasint"`
}

func DecodeAction(bytes []byte) (Action, error) {
	var a ActionCbor

	err := cbor.Unmarshal(bytes, &a)

	if err != nil {
		return nil, err
	}

	return a.convertToAction()
}

func NewActionHelper(a Action) *ActionHelper {
	return &ActionHelper{a}
}

func (h *ActionHelper) Encode() []byte {
	bytes, err := cbor.Marshal(h.convertToActionCbor())

	if err != nil {
		log.Fatal(err)
	}

	return bytes
}

func (h *ActionHelper) convertToActionCbor() ActionCbor {
	attr, err := cbor.Marshal(h.action)

	if err != nil {
		log.Fatal(err)
	}

	return ActionCbor{
		Category: h.action.GetCategory(),
		Name: h.action.GetName(),
		Attributes: attr,
	}
}

func RegisterAction(category string, name string, generator ActionGenerator) bool {
	if prev, ok := ACTIONS[category]; ok {
		prev[name] = generator
	} else {
		ACTIONS[category] = map[string]ActionGenerator{
			name: generator,
		}
	}

	return true
}

func (a ActionCbor) convertToAction() (Action, error) {
	category := a.Category
	name := a.Name
	attrBytes := a.Attributes

	if categoryActions, ok := ACTIONS[category]; ok {
		if action, ok := categoryActions[name]; ok {
			return action(attrBytes)
		} else {
			return nil, errors.New("invalid " + category + " action " + name)
		}
	} else {
		return nil, errors.New("invalid category " + category)
	}
}