package ledger

type ResourceManager interface {
	AddNode(id string, addr string) error
	AddGateway(id string, port int) error
	AddGatewayEndpoint(id string, method string, path string, task string) error
	AddTask(id string, handler string) error
	RemoveTask(id string) error
	RemoveGateway(id string) error
}
