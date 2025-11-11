package corpus

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/models/support"
	"github.com/tcc2-davi-arthur/utils"
	"gorm.io/gorm"
)

func ApplyLegalInputsDir(db *gorm.DB, legalInputs string, algo support.Algo, preIndexed, normalizeJumps, parallel bool, size, jumps int) (*models.TestConfigResult, error) {
	inputs, err := os.ReadFile(legalInputs)
	if err != nil {
		return nil, fmt.Errorf("error reading legal entries file: %v", err)
	}

	data := new(support.InteractionPackage)
	if err = json.Unmarshal(inputs, &data); err != nil {
		return nil, fmt.Errorf("error unmarshalling legal entries: %v", err)
	}

	var all []models.Document
	if err = db.Model(&models.Document{}).Select(&all).Error; err != nil {
		return nil, fmt.Errorf("error reading documents: %v", err)
	}

	ret := models.NewTestConfigResult()
	docCache := make(map[uint16]map[string]*float64) // cache dos documentos
	phraseCache := make(map[string]map[string]*float64)

	processPhrase := func(phrase support.Interaction, pushFunc func(float64, int64)) error {
		// vetor TF-IDF da frase
		var phraseVec map[string]*float64
		var err error
		elapsedPhrase := utils.Stopwatch(func() {
			phraseVec, err = utils.ComputeStringTFIDF(phrase.Input, size, jumps, len(all), CacheGrams, CacheWords, normalizeJumps, parallel)
			phraseCache[phrase.Input] = phraseVec
		}).Milliseconds()
		if err != nil {
			return err
		}

		// acumula similaridade por documento para a frase
		docSimAccum := make(map[uint16]float64)

		for _, doc := range all {
			docVec, ok := docCache[doc.ID]
			if !ok {
				elapsed := utils.Stopwatch(func() {
					if algo == support.TdIdf {
						if preIndexed {
							docVec, err = utils.ComputeDocPreIndexedTFIDF(nil, len(CacheDocs), CacheGrams, normalizeJumps, parallel)
						} else {
							docVec, err = utils.ComputeDocPosIndexedTFIDF(doc.ID, size, len(all), db, normalizeJumps, parallel)
						}
					} else {
						if preIndexed {
							docVec, err = utils.ComputeDocPreIndexedBM25(nil, len(CacheDocs), CountAllNGrams, CacheGrams, normalizeJumps, parallel)
						} else {
							docVec, err = utils.ComputeDocPosIndexedBM25(doc.ID, size, len(all), db, normalizeJumps, parallel)
						}
					}
				})
				if err != nil {
					return err
				}
				docCache[doc.ID] = docVec
				ret.DocCalcBytesPerSecondAvgTime += float64(doc.Size) / elapsed.Seconds()
			}

			sim := utils.CosineSimMaps(phraseVec, docVec)
			docSimAccum[doc.ID] = sim
		}

		// ordena top documentos
		pairs := make([]mgu.Pair[uint16, float64], 0, len(docSimAccum))
		for id, sim := range docSimAccum {
			pairs = append(pairs, mgu.NewPair(id, sim))
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].Right > pairs[j].Right })

		// calcula Spearman médio
		list := mgu.VecMap(pairs, func(t mgu.Pair[uint16, float64]) uint16 { return t.Left })
		spearmanSim, err := utils.Spearman(phrase.Bert, list)
		if err != nil {
			return err
		}

		// atualiza stats usando Push
		pushFunc(spearmanSim, elapsedPhrase)
		return nil
	}

	for _, phrase := range data.Words10 {
		if e := processPhrase(phrase, ret.Push10); e != nil {
			return nil, e
		}
	}
	for _, phrase := range data.Words20 {
		if e := processPhrase(phrase, ret.Push20); e != nil {
			return nil, e
		}
	}
	for _, phrase := range data.Words40 {
		if e := processPhrase(phrase, ret.Push40); e != nil {
			return nil, e
		}
	}

	return &ret, nil
}
