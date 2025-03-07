package ledger

import ()

type AddCompute struct {
	addr string `cbor:"0,keyasint"`
}

func NewAddCompute(addr string) *AddCompute {
	return &AddCompute{addr}
}

func (c *AddCompute) GetName() string {
	return "AddCompute"
}

func (c *AddCompute) GetCategory() string {
	return "compute"
}

type RemoveCompute struct {
	addr string `cbor:"0,keyasint"`
}

func (c *RemoveCompute) GetName() string {
	return "RemoveCompute"
}

func (c *RemoveCompute) GetCategory() string {
	return "compute"
}

type AddUser struct {
	// public Ed25519 32 byte key
	// TODO: like to policy
	key PubKey `cbor:"0,keyasint"`
}

func (c *AddUser) GetName() string {
	return "AddUser"
}

func (c *AddUser) GetCategory() string {
	return "permissions"
}