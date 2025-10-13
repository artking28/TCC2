package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
)

const (
	DatabasePath  = "./data.db"
	ResultsOutput = "./misc/results.json"
)
const (
	Unigram = 1
	Bigram  = 2
	Trigram = 3
)

const (
	MaxTrigramJumps = 2
	MaxBigramJumps  = 4
)

type TestParams struct {
	Algo           models.Algo       `json:"algo"`
	PreIndexed     bool              `json:"preIndexed"`
	NormalizeJumps bool              `json:"normalizeJumps"`
	Size           int               `json:"size"`
	Jumps          int               `json:"jumps"`
	DurationMS     int64             `json:"durationMs"`
	Results        map[string]string `json:"results"`
}

func NewTestParams(algo models.Algo, preIndexed bool, normalizeJumps bool, size int, jumps int) TestParams {
	return TestParams{
		Algo:           algo,
		PreIndexed:     preIndexed,
		NormalizeJumps: normalizeJumps,
		Size:           size,
		Jumps:          jumps,
		Results:        make(map[string]string),
	}
}

func TestAll(t *testing.T) {

	var results []TestParams
	var id int64 = 1

	sizes := []mgu.Pair[int, int]{
		mgu.NewPair(Unigram, 0),
		mgu.NewPair(Bigram, MaxBigramJumps),
		mgu.NewPair(Trigram, MaxTrigramJumps),
	}

	// Loop principal: percorre cada tipo de n-gram configurado acima
	for _, s := range sizes {
		size, maxJumps := s.Left, s.Right

		// Para cada tamanho de n-gram, testa todos os níveis de salto possíveis
		for jump := 0; jump <= maxJumps; jump++ {

			// Testa as duas variações: com e sem pré-indexação (cache vs banco)
			for _, preIndexed := range []bool{false, true} {

				// Para cada modo de pré-indexação, testa normalização de jumps ligada e desligada
				for _, normalize := range []bool{false, true} {

					// Monta os parâmetros do teste atual com TF-IDF
					p0 := NewTestParams(models.TdIdf, preIndexed, normalize, size, jump)

					// Executa o teste com essa combinação de parâmetros
					BaseTest(t, id, p0, &results)
					id++

					// Monta os parâmetros do teste atual com BM25
					p1 := NewTestParams(models.Bm25, preIndexed, normalize, size, jump)

					// Executa o teste com essa combinação de parâmetros
					BaseTest(t, id, p1, &results)
					id++
				}
			}
		}
	}

	fmt.Println()
	content, err := json.MarshalIndent(results, "", "   ")
	if err != nil {
		t.Fatalf("erro ao serializar resultados: %v", err)
	}

	err = os.WriteFile(ResultsOutput, content, 0644)
	if err != nil {
		t.Fatalf("erro ao salvar results.json: %v", err)
	}
}

// Test executa um ciclo completo de benchmark e validação
// algo: modelo de ponderação (ex: BM25 ou TF-IDF)
// preIndexed: se true, usa cache em memória; se false, consulta o banco
// normalizeJumps: unifica n-grams com mesmos termos e jumps diferentes
// size: tamanho do n-gram
// jumps: número máximo de saltos entre palavras
func BaseTest(t *testing.T, testId int64, params TestParams, results *[]TestParams) {
	start := time.Now()
	fmt.Printf(
		"\n[TEST %02d] → algo=%s | preIndexed=%v | normalizeJumps=%v | size=%d | jumps=%d\n",
		testId, params.Algo, params.PreIndexed, params.NormalizeJumps, params.Size, params.Jumps,
	)

	elapsed := time.Since(start)
	params.DurationMS = elapsed.Milliseconds()

	fmt.Printf("[OK %02d] Finalizado em %s (%d n-grams processados)\n", testId, elapsed.String(), 1)
	*results = append(*results, params)
}
