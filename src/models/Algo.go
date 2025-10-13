package models

type Algo string

const (
	None  Algo = "none"
	TdIdf Algo = "tdIdf"
	Bm25  Algo = "bm25"
)

func NewAlgo(input string) Algo {
	switch input {
	case "tdIdf":
		return TdIdf
	case "bm25":
		return Bm25
	case "none":
	default:
		return None
	}
	return None
}

func (this Algo) ToString() string {
	return string(this)
}
