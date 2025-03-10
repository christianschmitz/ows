package actions

import (
	"log"
	"ows/ledger"
	"github.com/fxamacker/cbor/v2"
)

type AddUser struct {
	Key ledger.PubKey `cbor:"0,keyasint"`
}

func (c *AddUser) Apply(m ledger.ResourceManager, gen ledger.ResourceIdGenerator) {
	log.Fatal("not yet implemented")
}

func (c *AddUser) GetName() string {
	return "AddUser"
}

func (c *AddUser) GetCategory() string {
	return "permissions"
}

var _AddUserRegistered = ledger.RegisterAction("permissions", "AddUser", func (attr []byte) (ledger.Action, error) {
	var c AddUser
	err := cbor.Unmarshal(attr, &c)
	return &c, err
})