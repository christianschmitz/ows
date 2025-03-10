package ledger

type ResourceManager interface {
	AddCompute(id ResourceId, addr string)
	AddTask(id ResourceId, handler string)
}