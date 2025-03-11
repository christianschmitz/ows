package ledger

type ResourceManager interface {
	AddCompute(id ResourceId, addr string) error
	AddTask(id ResourceId, handler AssetId) error
	AddGateway(id ResourceId, port int) error
	AddGatewayEndpoint(id ResourceId, method string, path string, task ResourceId) error
}
