package ledger

type ResourceManager interface {
	AddCompute(id ResourceId, addr string) error
	AddTask(id ResourceId, handler string) error
}
