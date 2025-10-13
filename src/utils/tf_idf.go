package utils

import (
	"fmt"
	"math"
	"sort"

	"github.com/tcc2-davi-arthur/models"
	"gorm.io/gorm"
)

type TFIDFEntry struct {
	Key   string
	Value float64
}

func ComputePreIndexedTFIDF(trigramList []*models.InverseNGram, totalDocs int, cacheN map[string]*models.InverseNGram, smooth bool) ([]TFIDFEntry, error) {
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
		key := ngram.GetCacheKey(smooth, true)
		tf[key]++
	}

	// Calcula TF-IDF usando CacheN para DF
	for key, count := range tf {
		docSet := make(map[uint16]bool)
		for k, ngram := range cacheN {
			if k == key {
				docSet[ngram.DocId] = true
			}
		}
		df := len(docSet)

		tfVal := float64(count) / float64(totalTrigrams)
		idfVal := math.Log(float64(totalDocs) / (1 + float64(df)))
		tfidf[key] = tfVal * idfVal
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

func ComputePosIndexedTFIDF(docID uint16, totalDocs int, db *gorm.DB, smooth bool) ([]TFIDFEntry, error) {
	if totalDocs <= 0 {
		return nil, fmt.Errorf("totalDocs must be positive")
	}

	tfidf := make(map[string]float64)
	tf := make(map[string]int)
	df := make(map[string]int)

	// TF: contagem de trigramas do documento
	var tfResults []struct {
		Wd0Id, Wd1Id, Wd2Id uint16
		Jump0, Jump1        int8
		CountInDoc          int64
	}
	db.Model(&models.InverseNGram{}).
		Select("wd0_id, wd1_id, wd2_id, jump0, jump1, COUNT(doc_id) AS count_in_doc").
		Where("doc_id = ?", docID).
		Group("wd0_id, wd1_id, wd2_id, jump0, jump1").
		Find(&tfResults)

	totalTrigrams := 0
	for _, r := range tfResults {
		ngram := &models.InverseNGram{
			Wd0Id: r.Wd0Id,
			Wd1Id: r.Wd1Id,
			Wd2Id: r.Wd2Id,
			Jump0: r.Jump0,
			Jump1: r.Jump1,
			DocId: docID,
		}
		key := ngram.GetCacheKey(smooth, true)
		tf[key] = int(r.CountInDoc)
		totalTrigrams += int(r.CountInDoc)
	}

	var trigramKeys []struct {
		Wd0Id, Wd1Id, Wd2Id uint16
		Jump0, Jump1        int8
	}
	for _, r := range tfResults {
		trigramKeys = append(trigramKeys, struct {
			Wd0Id, Wd1Id, Wd2Id uint16
			Jump0, Jump1        int8
		}{r.Wd0Id, r.Wd1Id, r.Wd2Id, r.Jump0, r.Jump1})
	}

	// Calcula DF
	for _, t := range trigramKeys {
		var docCount int64
		ngram := &models.InverseNGram{
			Wd0Id: t.Wd0Id,
			Wd1Id: t.Wd1Id,
			Wd2Id: t.Wd2Id,
			Jump0: t.Jump0,
			Jump1: t.Jump1,
		}
		key := ngram.GetCacheKey(smooth, true)
		query := db.Model(&models.InverseNGram{}).
			Where("wd0_id = ? AND wd1_id = ? AND wd2_id = ?", t.Wd0Id, t.Wd1Id, t.Wd2Id)
		if !smooth {
			query = query.Where("jump0 = ? AND jump1 = ?", t.Jump0, t.Jump1)
		}
		query.Distinct("doc_id").Count(&docCount)
		df[key] = int(docCount)
	}

	// Calcula TF-IDF
	for key, count := range tf {
		tfVal := float64(count) / float64(totalTrigrams)
		idfVal := math.Log(float64(totalDocs) / (1 + float64(df[key])))
		tfidf[key] = tfVal * idfVal
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
