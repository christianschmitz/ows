package ledger

import (
	"fmt"
	"log"
)

type AddCompute struct {
	Address string `cbor:"0,keyasint"`
}

func NewAddCompute(addr string) *AddCompute {
	return &AddCompute{addr}
}

func (c *AddCompute) Apply(m ResourceManager, gen ResourceIdGenerator) {
	id := gen()

	fmt.Println("addr: " + c.Address)
	m.AddCompute(id, c.Address)
}

func (c *AddCompute) GetName() string {
	return "AddCompute"
}

func (c *AddCompute) GetCategory() string {
	return "compute"
}

type RemoveCompute struct {
	Id ResourceId `cbor:"0,keyasint"`
	Address string `cbor:"1,keyasint"`
}

func (c *RemoveCompute) Apply(m ResourceManager, gen ResourceIdGenerator) {
	log.Fatal("not yet implemented")
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
	Key PubKey `cbor:"0,keyasint"`
}

func (c *AddUser) Apply(m ResourceManager, gen ResourceIdGenerator) {
	log.Fatal("not yet implemented")
}

func (c *AddUser) GetName() string {
	return "AddUser"
}

func (c *AddUser) GetCategory() string {
	return "permissions"
}

type AddTask struct {
	Runtime string `cbor:"0,keyasint"`
	Handler string `cbor:"1,keyasint"`
}

func (c *AddTask) Apply(m ResourceManager, gen ResourceIdGenerator) {
	m.AddTask(gen(), c.Handler)
}

func (c *AddTask) GetName() string {
	return "AddTask"
}

func (c *AddTask) GetCategory() string {
	return "tasks"
}