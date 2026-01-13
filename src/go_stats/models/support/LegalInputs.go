package support

type InteractionPackage struct {
	Words10 []Interaction `json:"words10"`
	Words20 []Interaction `json:"words20"`
	Words40 []Interaction `json:"words40"`
}

type Interaction struct {
	Input     string   `json:"input"`
	Bert      []uint16 `json:"bert"`
	BertT     int64    `json:"bertT"`
	Word2vec  []uint16 `json:"word2vec"`
	Word2vecT int64    `json:"word2vecT"`
	Glove     []uint16 `json:"glove"`
	GloveT    int64    `json:"gloveT"`
}
