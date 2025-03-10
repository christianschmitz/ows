package ledger

type ResourceId = [32]byte

type ResourceIdGenerator = func () ResourceId

func StringifyResourceId(id ResourceId) string {
	return StringifyBytes(id[:])
}