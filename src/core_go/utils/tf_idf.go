package utils

import (
	"fmt"
	"math"
	"strings"
	"sync"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/models/interfaces"
	"github.com/tcc2-davi-arthur/repository"
	"gorm.io/gorm"
)

// ComputeDocPreIndexedTFIDF calcula o TF-IDF de um documento previamente indexado
// usando um cache de n-gramas no formato map[string]map[uint16]interfaces.IGram.
// - trigramList: lista dos n-gramas do documento alvo
// - totalDocs: número total de documentos do corpus
// - cacheN: cache global de n-gramas agrupado por chave e docID
// - normalizeJumps: define se jumps são normalizados
// - parallel: ativa processamento concorrente
func ComputeDocPreIndexedTFIDF(
	trigramList []interfaces.IGram,
	totalDocs int,
	cacheN map[string]map[uint16]interfaces.IGram,
	normalizeJumps, parallel bool,
) (map[string]*float64, error) {

	if len(trigramList) == 0 {
		return nil, fmt.Errorf("trigram list is empty")
	}
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	expectedDocID := trigramList[0].GetDocId()
	tfidf := make(map[string]*float64)
	tf := make(map[string]int)
	totalTrigrams := len(trigramList)

	// Calcula TF
	for _, ngram := range trigramList {
		if ngram.GetDocId() != expectedDocID {
			return nil, fmt.Errorf("mismatched DocId: expected %d, got %d", expectedDocID, ngram.GetDocId())
		}
		key := ngram.GetCacheKey(normalizeJumps, true)
		tf[key]++
	}

	// Calcula DF
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
			go func(key string) {
				defer wg.Done()
				defer func() { <-sem }()
				df := len(cacheN[key]) // conta quantos docs têm esse termo
				dfChan <- dfResult{key: key, df: df}
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

	// Coleta resultados do DF e calcula TF-IDF
	for res := range dfChan {
		tfVal := float64(tf[res.key]) / float64(totalTrigrams)
		idfVal := math.Log(float64(totalDocs) / (1 + float64(res.df)))
		tfidf[res.key] = mgu.Ptr(tfVal * idfVal)
	}

	return tfidf, nil
}

// ComputeDocPosIndexedTFIDF calcula o TF-IDF de um documento diretamente a partir do banco de dados,
// sem depender de um cache pré-carregado. Ele obtém as frequências de termos (TF) e a frequência
// de documentos (DF) usando consultas SQL, podendo operar de forma concorrente.
//
// Parâmetros:
// - docID: identificador do documento alvo
// - gramSize: tamanho do n-grama (1, 2 ou 3)
// - totalDocs: número total de documentos do corpus
// - db: conexão GORM com o banco de dados
// - normalizeJumps: define se os jumps devem ser normalizados
// - parallel: ativa execução concorrente das consultas DF
//
// Retorna:
// - map[string]*float64: mapa de chaves de n-grama para seus valores TF-IDF
func ComputeDocPosIndexedTFIDF(docID uint16, gramSize, totalDocs int, db *gorm.DB, normalizeJumps, parallel bool) (map[string]*float64, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	tfidf := make(map[string]*float64)
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

	// Coleta resultados do DF
	for res := range dfChan {
		tfVal := float64(tf[res.key]) / float64(totalTrigrams)
		idfVal := math.Log(float64(totalDocs) / (1 + float64(res.df)))
		tfidf[res.key] = mgu.Ptr(tfVal * idfVal)
	}

	return tfidf, nil
}

// ComputeStringTFIDF calcula o TF-IDF de uma frase simples, convertendo-a internamente
// para n-grams compatíveis com o cache global. Usa docId=0 pois a frase não pertence
// a nenhum documento real.
//
// Parâmetros:
// - str: texto alvo
// - gramsSize: tamanho do n-gram (1, 2 ou 3)
// - jumpSize: distância máxima entre termos (para gerar jumps)
// - totalDocs: número total de documentos do corpus
// - cacheN: cache global no formato map[string]map[uint16]interfaces.IGram
// - CacheWords: mapa de palavras para seus objetos indexados
// - smoothJumps: normaliza jumps na geração das chaves
// - parallel: ativa concorrência no cálculo DF
//
// Retorna:
// - map[string]*float64: valores TF-IDF por chave de n-gram
// - error: erro em caso de falha
func ComputeStringTFIDF(
	str string,
	gramsSize, jumpSize, totalDocs int,
	cacheN map[string]map[uint16]interfaces.IGram,
	CacheWords map[string]models.Word,
	smoothJumps, parallel bool,
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
		return nil, nil
	}

	var grams []interfaces.IGram

	// Gera os n-grams da frase (docId = 0)
	for i, word := range result {
		var ngram interfaces.IGram
		switch gramsSize {
		case 1:
			ngram = models.NewInverseUnigram(0, 0, CacheWords[word[0]].ID)
		case 2:
			ngram = models.NewInverseBigram(0, 0, CacheWords[word[0]].ID, CacheWords[word[1]].ID, jumps[i][0])
		case 3:
			ngram = models.NewInverseTrigram(0, 0, CacheWords[word[0]].ID, CacheWords[word[1]].ID,
				CacheWords[word[2]].ID, jumps[i][0], jumps[i][1])
		default:
			return nil, fmt.Errorf("invalid gramsSize: %d", gramsSize)
		}
		ngram.Increment()
		grams = append(grams, ngram)
	}

	tf := make(map[string]int, len(grams))
	tfidf := make(map[string]*float64, len(grams))
	totalGrams := len(grams)

	for _, ngram := range grams {
		key := ngram.GetCacheKey(smoothJumps, true)
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
			go func(key string) {
				defer wg.Done()
				defer func() { <-sem }()
				df := len(cacheN[key])
				dfChan <- dfResult{key: key, df: df}
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

	// Calcula TF-IDF
	for res := range dfChan {
		tfVal := float64(tf[res.key]) / float64(totalGrams)
		idfVal := math.Log(float64(totalDocs) / (1 + float64(res.df)))
		tfidf[res.key] = mgu.Ptr(tfVal * idfVal)
	}

	return tfidf, nil
}
