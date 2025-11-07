package utils

import (
	"fmt"
	"log"
	"math"
	"sync"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models/interfaces"
	"gorm.io/gorm"
)

func ComputePreIndexedTFIDF(trigramList []interfaces.IGram, totalDocs int, cacheGrams map[string]interfaces.IGram, smoothJumps, parallel bool) (map[string]*float64, error) {
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
		key := ngram.GetCacheKey(smoothJumps, true)
		tf[key]++
	}

	// Calcula DF
	type dfResult struct {
		key string
		df  int
	}
	dfChan := make(chan dfResult, len(tf))

	if parallel {
		// Concorrência com até 25 goroutines
		sem := make(chan struct{}, 25) // Semaforo para limitar goroutines
		var wg sync.WaitGroup

		for key := range tf {
			wg.Add(1)
			sem <- struct{}{} // Adquire semaforo
			go func(key string) {
				defer wg.Done()
				defer func() { <-sem }() // Libera semaforo
				docSet := make(map[uint16]bool)
				for k, ngram := range cacheGrams {
					if k == key {
						docSet[ngram.GetDocId()] = true
					}
				}
				dfChan <- dfResult{key: key, df: len(docSet)}
			}(key)
		}

		go func() {
			wg.Wait()
			close(dfChan)
		}()
	} else {
		// Sequencial
		for key := range tf {
			docSet := make(map[uint16]bool)
			for k, ngram := range cacheGrams {
				if k == key {
					docSet[ngram.GetDocId()] = true
				}
			}
			dfChan <- dfResult{key: key, df: len(docSet)}
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

func ComputePosIndexedTFIDF(docID uint16, gramSize, totalDocs int, db *gorm.DB, normalizeJumps, parallel bool) (map[string]*float64, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	tfidf := make(map[string]*float64)
	tf := make(map[string]int)
	totalTrigrams := 0

	var label string
	switch gramSize {
	case 1:
		label = "wd0Id"
		break
	case 2:
		label = "wd0Id, wd1Id, jump0"
		break
	case 3:
		label = "wd0Id, wd1Id, wd2Id, jump0, jump1"
		break
	default:
		log.Fatalf("invalid gramSize: %d", gramSize)
	}

	// TF: contagem de trigramas do documento
	var tfResults []interfaces.IGram
	err := db.Model("WORD_DOC").
		Select(fmt.Sprintf("%s, COUNT(docId) AS count", label)).
		Where("docId = ?", docID).
		Group(label).
		Find(&tfResults).Error
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
