package utils

import (
	"fmt"
	"log"
	"math"
	"strings"
	"sync"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/models/interfaces"
	"github.com/tcc2-davi-arthur/repository"
	"gorm.io/gorm"
)

// ComputeDocPreIndexedBM25 calcula o peso BM25 de cada trigram presente em um documento,
// usando informações pré-indexadas do corpus.
//
// Parâmetros:
//   - trigramList: lista de trigrams pertencentes a um único documento.
//   - totalDocs: número total de documentos no corpus.
//   - totalGrams: quantidade total de trigrams no corpus (para cálculo do avgDL).
//   - cacheN: índice global de trigrams do corpus (key → trigram), usado para obter DF.
//   - normalizeJumps: define se as chaves dos trigrams devem ser normalizadas.
//   - parallel: executa o cálculo de DF de forma concorrente (limite de 25 goroutines).
//
// Retorno:
//   - map[string]*float64: pontuação BM25 de cada trigram do documento.
func ComputeDocPreIndexedBM25(
	trigramList []interfaces.IGram,
	totalDocs, totalGrams int,
	cacheN map[string]map[uint16]interfaces.IGram,
	normalizeJumps, parallel bool,
) (map[string]*float64, error) {

	if len(trigramList) == 0 {
		return nil, fmt.Errorf("trigram list is empty")
	}
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	docID := trigramList[0].GetDocId()
	bm25 := make(map[string]*float64)
	tf := make(map[string]int)
	totalTrigrams := len(trigramList)

	for _, ngram := range trigramList {
		if ngram.GetDocId() != docID {
			return nil, fmt.Errorf("mismatched DocId: expected %d, got %d", docID, ngram.GetDocId())
		}
		key := ngram.GetCacheKey(normalizeJumps, false)
		tf[key]++
	}

	type dfResult struct {
		key string
		df  int
	}
	dfChan := make(chan dfResult, len(tf))

	if parallel {
		sem := make(chan struct{}, 25)
		var wg sync.WaitGroup

		for key := range tf {
			wg.Add(1)
			sem <- struct{}{}
			go func(k string) {
				defer wg.Done()
				defer func() { <-sem }()
				df := len(cacheN[k]) // conta quantos documentos contêm o termo
				dfChan <- dfResult{key: k, df: df}
			}(key)
		}

		go func() {
			wg.Wait()
			close(dfChan)
		}()
	} else {
		for key := range tf {
			df := len(cacheN[key])
			dfChan <- dfResult{key: key, df: df}
		}
		close(dfChan)
	}

	k := 1.5
	b := 0.75
	docLen := float64(totalTrigrams)
	avgDL := float64(totalGrams) / float64(totalDocs)

	for res := range dfChan {
		tfTerm := float64(tf[res.key])
		df := float64(res.df)
		if df == 0 {
			continue
		}

		idf := math.Log((float64(totalDocs)-df+0.5)/(df+0.5) + 1)
		score := idf * (tfTerm * (k + 1)) / (tfTerm + k*(1-b+b*(docLen/avgDL)))
		bm25[res.key] = mgu.Ptr(score)
	}

	return bm25, nil
}

// ComputeDocPosIndexedBM25 calcula os pesos BM25 dos n-grams (1, 2 ou 3) de um documento
// específico, acessando diretamente os dados do banco via GORM.
//
// Parâmetros:
//   - docID: identificador do documento a ser processado.
//   - gramSize: tamanho do n-gram (1 = unigram, 2 = bigram, 3 = trigram).
//   - totalDocs: número total de documentos no corpus.
//   - db: instância ativa do *gorm.DB* para consultas SQL.
//   - normalizeJumps: indica se os campos de salto (jump) devem ser normalizados.
//   - parallel: executa o cálculo de DF de forma concorrente (até 25 goroutines).
//
// Retorno:
//   - map[string]*float64: pontuação BM25 calculada para cada n-gram do documento.
func ComputeDocPosIndexedBM25(docID uint16, gramSize, totalDocs int, db *gorm.DB, normalizeJumps, parallel bool) (map[string]*float64, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	bm25 := make(map[string]*float64)
	tf := make(map[string]int)
	totalTrigrams := 0

	// TF: contagem de trigramas do documento
	tfResults, err := repository.NewGramRepository(db).FindByDocAndSize(docID, gramSize)
	if err != nil {
		return nil, err
	}

	for _, r := range tfResults {
		key := r.GetCacheKey(normalizeJumps, true)
		tf[key] = r.GetCount()
		totalTrigrams += r.GetCount()
	}

	// Calcula DF
	type dfResult struct {
		key string
		df  int
	}
	dfChan := make(chan dfResult, len(tfResults))

	if parallel {
		// Concorrência com até 25 goroutines
		sem := make(chan struct{}, 25) // Semaforo para limitar goroutines
		var wg sync.WaitGroup

		for _, r := range tfResults {
			wg.Add(1)
			sem <- struct{}{} // Adquire semaforo
			go func(r interfaces.IGram) {
				defer wg.Done()
				defer func() { <-sem }() // Libera semaforo
				key := r.GetCacheKey(normalizeJumps, true)
				var docCount int64
				query := r.ApplyWordWheres(db)
				if !normalizeJumps {
					query = r.ApplyJumpWheres(query)
				}
				query.Distinct("docId").Count(&docCount)
				dfChan <- dfResult{key: key, df: int(docCount)}
			}(r)
		}

		go func() {
			wg.Wait()
			close(dfChan)
		}()
	} else {
		// Sequencial
		for _, r := range tfResults {
			key := r.GetCacheKey(normalizeJumps, true)
			var docCount int64
			query := r.ApplyWordWheres(db)
			if !normalizeJumps {
				query = r.ApplyJumpWheres(query)
			}
			query.Distinct("docId").Count(&docCount)
			dfChan <- dfResult{key: key, df: int(docCount)}
		}
		close(dfChan)
	}

	var totalTrigramsAllDocs int64
	err = db.Table("WORD_DOC").Select("SUM(count)").Scan(&totalTrigramsAllDocs).Error
	if err != nil {
		log.Fatal(err)
	}

	k := 1.5
	b := 0.75
	docLen := float64(totalTrigrams) // ou o tamanho do documento
	avgDL := float64(totalTrigramsAllDocs) / float64(totalDocs)

	for res := range dfChan {
		tfTerm := float64(tf[res.key])
		idf := math.Log((float64(totalDocs)-float64(res.df)+0.5)/(float64(res.df)+0.5) + 1)
		bm25Score := idf * (tfTerm * (k + 1)) / (tfTerm + k*(1-b+b*(docLen/avgDL)))
		bm25[res.key] = mgu.Ptr(bm25Score)
	}

	return bm25, nil
}

// ComputeStringBM25 calcula os pesos BM25 de uma frase simples (não associada a um documento real),
// convertendo-a em n-grams e utilizando o cache global como base para obter DF.
//
// Parâmetros:
// - str: texto de entrada
// - gramsSize: tamanho do n-gram (1, 2 ou 3)
// - jumpSize: tamanho máximo dos jumps entre termos
// - totalDocs: número total de documentos no corpus
// - totalGrams: número total de n-grams no corpus (para cálculo de avgDL)
// - cacheN: cache global no formato map[string]map[uint16]interfaces.IGram
// - cacheWords: mapa de palavras indexadas
// - normalizeJumps: indica se os jumps devem ser normalizados
// - parallel: ativa execução concorrente
func ComputeStringBM25(
	str string,
	gramsSize, jumpSize, totalDocs, totalGrams int,
	cacheN map[string]map[uint16]interfaces.IGram,
	cacheWords map[string]models.Word,
	normalizeJumps, parallel bool,
) (map[string]*float64, error) {

	if str == "" {
		return nil, fmt.Errorf("input text is empty")
	}
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	text := strings.Fields(str)
	if len(text) == 0 {
		return nil, fmt.Errorf("no valid tokens found")
	}

	result, jumps := GetGramsLim(text, gramsSize, jumpSize)
	if len(result) == 0 {
		return nil, fmt.Errorf("no n-grams generated")
	}

	var grams []interfaces.IGram

	// Gera os n-grams com docId = 0
	for i, word := range result {
		var ngram interfaces.IGram
		switch gramsSize {
		case 1:
			w0, ok := cacheWords[word[0]]
			if !ok {
				continue
			}
			ngram = models.NewInverseUnigram(0, 0, w0.ID)
			break
		case 2:
			w0, ok0 := cacheWords[word[0]]
			w1, ok1 := cacheWords[word[1]]
			if !ok0 || !ok1 {
				continue
			}
			ngram = models.NewInverseBigram(0, 0, w0.ID, w1.ID, jumps[i][0])
			break
		case 3:
			w0, ok0 := cacheWords[word[0]]
			w1, ok1 := cacheWords[word[1]]
			w2, ok2 := cacheWords[word[2]]
			if !ok0 || !ok1 || !ok2 {
				continue
			}
			ngram = models.NewInverseTrigram(0, 0, w0.ID, w1.ID, w2.ID, jumps[i][0], jumps[i][1])
			break
		default:
			return nil, fmt.Errorf("invalid gramsSize: %d", gramsSize)
		}
		ngram.Increment()
		grams = append(grams, ngram)
	}

	bm25 := make(map[string]*float64, len(grams))
	tf := make(map[string]int, len(grams))
	totalGramsDoc := len(grams)

	for _, ngram := range grams {
		key := ngram.GetCacheKey(normalizeJumps, false)
		tf[key]++
	}

	type dfResult struct {
		key string
		df  int
	}
	dfChan := make(chan dfResult, len(tf))

	if parallel {
		sem := make(chan struct{}, 25)
		var wg sync.WaitGroup
		for key := range tf {
			wg.Add(1)
			sem <- struct{}{}
			go func(k string) {
				defer wg.Done()
				defer func() { <-sem }()
				df := len(cacheN[k])
				dfChan <- dfResult{key: k, df: df}
			}(key)
		}
		go func() {
			wg.Wait()
			close(dfChan)
		}()
	} else {
		for key := range tf {
			df := len(cacheN[key])
			dfChan <- dfResult{key: key, df: df}
		}
		close(dfChan)
	}

	k := 1.5
	b := 0.75
	docLen := float64(totalGramsDoc)
	avgDL := float64(totalGrams) / float64(totalDocs)

	for res := range dfChan {
		tfTerm := float64(tf[res.key])
		df := float64(res.df)
		if df == 0 {
			continue
		}
		idf := math.Log((float64(totalDocs)-df+0.5)/(df+0.5) + 1)
		score := idf * (tfTerm * (k + 1)) / (tfTerm + k*(1-b+b*(docLen/avgDL)))
		bm25[res.key] = mgu.Ptr(score)
	}

	return bm25, nil
}
