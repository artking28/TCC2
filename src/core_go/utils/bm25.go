package utils

import (
	"fmt"
	"log"
	"math"
	"sync"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/models/interfaces"
	"gorm.io/gorm"
)

func ComputePreIndexedBM25(trigramList []interfaces.IGram, totalDocs, totalGrams int, cacheN map[string]interfaces.IGram, smoothJumps, parallel bool) (map[string]*float64, error) {
	if len(trigramList) == 0 {
		return nil, fmt.Errorf("trigram list is empty")
	}
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	expectedDocID := trigramList[0].GetDocId()
	bm25 := make(map[string]*float64)
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
				for k, ngram := range cacheN {
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
			for k, ngram := range cacheN {
				if k == key {
					docSet[ngram.GetDocId()] = true
				}
			}
			dfChan <- dfResult{key: key, df: len(docSet)}
		}
		close(dfChan)
	}

	k := 1.5
	b := 0.75
	docLen := float64(totalTrigrams) // ou o tamanho do documento
	avgDL := float64(totalGrams) / float64(totalDocs)

	for res := range dfChan {
		tfTerm := float64(tf[res.key])
		idf := math.Log((float64(totalDocs)-float64(res.df)+0.5)/(float64(res.df)+0.5) + 1)
		bm25Score := idf * (tfTerm * (k + 1)) / (tfTerm + k*(1-b+b*(docLen/avgDL)))
		bm25[res.key] = mgu.Ptr(bm25Score)
	}

	return bm25, nil
}

func ComputePosIndexedBM25(docID uint16, gramSize, totalDocs int, db *gorm.DB, normalizeJumps, parallel bool) (map[string]*float64, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	bm25 := make(map[string]*float64)
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
		totalTrigrams += int(r.GetCount())
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
	err = db.Model(&models.InverseTrigram{}).Select("SUM(count)").Scan(&totalTrigramsAllDocs).Error
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
