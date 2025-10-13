package utils

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/tcc2-davi-arthur/models"
	"gorm.io/gorm"
)

type TFIDFEntry struct {
	Key   string
	Value float64
}

func ComputePreIndexedTFIDF(trigramList []*models.InverseNGram, totalDocs int, cacheN map[string]*models.InverseNGram, smoothJumps, parallel bool) ([]TFIDFEntry, error) {
	if len(trigramList) == 0 {
		return nil, fmt.Errorf("trigram list is empty")
	}
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	expectedDocID := trigramList[0].DocId
	tfidf := make(map[string]float64)
	tf := make(map[string]int)
	totalTrigrams := len(trigramList)

	// Calcula TF
	for _, ngram := range trigramList {
		if ngram.DocId != expectedDocID {
			return nil, fmt.Errorf("mismatched DocId: expected %d, got %d", expectedDocID, ngram.DocId)
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
				for k, ngram := range cacheN {
					if k == key {
						docSet[ngram.DocId] = true
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
			for k, ngram := range cacheN {
				if k == key {
					docSet[ngram.DocId] = true
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
		tfidf[res.key] = tfVal * idfVal
	}

	// Converte mapa para slice e ordena por chave
	result := make([]TFIDFEntry, 0, len(tfidf))
	for key, value := range tfidf {
		result = append(result, TFIDFEntry{Key: key, Value: value})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result, nil
}

func ComputePosIndexedTFIDF(docID uint16, totalDocs int, db *gorm.DB, smoothJumps, parallel bool) ([]TFIDFEntry, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	tfidf := make(map[string]float64)
	tf := make(map[string]int)
	totalTrigrams := 0

	// TF: contagem de trigramas do documento
	var tfResults []*models.InverseNGram
	err := db.Model(&models.InverseNGram{}).
		Select("wd0Id, wd1Id, wd2Id, jump0, jump1, COUNT(docId) AS count").
		Where("docId = ?", docID).
		Group("wd0Id, wd1Id, wd2Id, jump0, jump1").
		Find(&tfResults).Error
	if err != nil {
		return nil, err
	}

	for _, r := range tfResults {
		ngram := &models.InverseNGram{
			Wd0Id: r.Wd0Id,
			Wd1Id: r.Wd1Id,
			Wd2Id: r.Wd2Id,
			Jump0: r.Jump0,
			Jump1: r.Jump1,
			DocId: docID,
		}
		key := ngram.GetCacheKey(smoothJumps, true)
		tf[key] = int(r.Count)
		totalTrigrams += int(r.Count)
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
			go func(r *models.InverseNGram) {
				defer wg.Done()
				defer func() { <-sem }() // Libera semaforo
				ngram := &models.InverseNGram{
					Wd0Id: r.Wd0Id,
					Wd1Id: r.Wd1Id,
					Wd2Id: r.Wd2Id,
					Jump0: r.Jump0,
					Jump1: r.Jump1,
				}
				key := ngram.GetCacheKey(smoothJumps, true)
				var docCount int64
				query := db.Model(&models.InverseNGram{}).
					Where("wd0Id = ? AND wd1Id = ? AND wd2Id = ?", r.Wd0Id, r.Wd1Id, r.Wd2Id)
				if !smoothJumps {
					query = query.Where("jump0 = ? AND jump1 = ?", r.Jump0, r.Jump1)
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
			ngram := &models.InverseNGram{
				Wd0Id: r.Wd0Id,
				Wd1Id: r.Wd1Id,
				Wd2Id: r.Wd2Id,
				Jump0: r.Jump0,
				Jump1: r.Jump1,
				DocId: docID,
			}
			key := ngram.GetCacheKey(smoothJumps, true)
			var docCount int64
			query := db.Model(&models.InverseNGram{}).
				Where("wd0Id = ? AND wd1Id = ? AND wd2Id = ?", r.Wd0Id, r.Wd1Id, r.Wd2Id)
			if !smoothJumps {
				query = query.Where("jump0 = ? AND jump1 = ?", r.Jump0, r.Jump1)
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
		tfidf[res.key] = tfVal * idfVal
	}

	// Converte mapa para slice e ordena por chave
	result := make([]TFIDFEntry, 0, len(tfidf))
	for key, value := range tfidf {
		result = append(result, TFIDFEntry{Key: key, Value: value})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result, nil
}
