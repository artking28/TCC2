package interfaces

type IGram interface {
	GetCacheKey(jumps, doc bool) string
	GetDocId() uint16
	Increment()
}
