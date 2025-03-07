package ledger

import (
	"errors"
	"log"
	"github.com/fxamacker/cbor/v2"
)

type Action interface {
	GetName() string

	// valid categories are: compute, permissions
	GetCategory() string
}

type ActionHelper struct {
	action Action
}

type ActionCbor struct {
	Name       string `cbor:"0,keyasint"`
	Category   string `cbor:"1,keyasint"`
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
		h.action.GetName(),
		h.action.GetCategory(),
		attr,
	}
}

func (a ActionCbor) convertToAction() (Action, error) {
	name := a.Name
	category := a.Category
	attrBytes := a.Attributes

	switch category {
	case "compute":
		switch name {
		case "AddCompute":
			var c AddCompute
			err := cbor.Unmarshal(attrBytes, &c)
			return &c, err
		case "RemoveCompute":
			var c RemoveCompute
			err := cbor.Unmarshal(attrBytes, &c)
			return &c, err
		default:
			return nil, errors.New("invalid " + category + " action " + name)
		}
	case "permissions":
		switch name {
		case "AddUser":
			var c AddUser
			err := cbor.Unmarshal(attrBytes, &c)
			return &c, err
		default:
			return nil, errors.New("invalid " + category + " action " + name)
		}
	default:
		return nil, errors.New("invalid category " + category)
	}
}