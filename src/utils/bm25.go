package utils

import (
	"fmt"
	"log"
	"math"
	"sync"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"gorm.io/gorm"
)

func ComputePreIndexedBM25(trigramList []*models.InverseNGram, totalDocs, totalGrams int, cacheN map[string]*models.InverseNGram, smoothJumps, parallel bool) (map[string]*float64, error) {
	if len(trigramList) == 0 {
		return nil, fmt.Errorf("trigram list is empty")
	}
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	expectedDocID := trigramList[0].DocId
	bm25 := make(map[string]*float64)
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

func ComputePosIndexedBM25(docID uint16, totalDocs int, db *gorm.DB, smoothJumps, parallel bool) (map[string]*float64, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	bm25 := make(map[string]*float64)
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

	var totalTrigramsAllDocs int64
	err = db.Model(&models.InverseNGram{}).Select("SUM(count)").Scan(&totalTrigramsAllDocs).Error
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
