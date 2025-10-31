package main

import (
	"encoding/json"
	"os"
)

type DataSample struct {
	Words10 []Test `json:"words10"`
	Words20 []Test `json:"words20"`
	Words40 []Test `json:"words40"`
}

type Test struct {
	Input string `json:"input"`

	Word2Vec  []uint16 `json:"word2vec"`
	Word2VecT uint64   `json:"word2vecT"`

	LegalBert  []uint16 `json:"legalBert"`
	LegalBertT uint64   `json:"legalBertT"`

	OpenAi  []uint16 `json:"openai"`
	OpenAiT uint64   `json:"openaiT"`

	Bert  []uint16 `json:"bert"`
	BertT uint64   `json:"bertT"`
}

func main() {

	content, err := os.ReadFile("searchLegalInputs.json")
	if err != nil {
		panic(err)
	}

	var obj DataSample
	json.Unmarshal(content, obj)
}
