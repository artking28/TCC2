package corpus

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"sync"

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

	var all []*models.Document
	if err = db.Model(&models.Document{}).Find(&all).Error; err != nil {
		return nil, fmt.Errorf("error reading documents: %v", err)
	}

	ret := models.NewTestConfigResult(len(all))
	docCache := make(map[uint16]map[string]*float64)
	phraseCache := make(map[string]map[string]*float64)

	processPhrase := func(phrase support.Interaction, pushFunc func(float64, int64)) error {
		var phraseVec map[string]*float64
		var err error

		elapsedPhrase := utils.Stopwatch(func() {
			phraseVec, err = utils.ComputeStringTFIDF(phrase.Input, size, jumps, len(all), CacheGrams, CacheWords, normalizeJumps, parallel)
			phraseCache[phrase.Input] = phraseVec
		}).Microseconds()
		if err != nil {
			return err
		}

		docSimAccum := make(map[uint16]float64)
		var mu sync.Mutex
		wg := sync.WaitGroup{}
		sem := make(chan struct{}, runtime.NumCPU())
		progress := make(chan int, len(all))

		ret.TotalTime += utils.Stopwatch(func() {
			for _, doc := range all {
				wg.Add(1)
				sem <- struct{}{}
				go func(doc *models.Document) {
					defer wg.Done()
					defer func() { <-sem }()
					var err error

					mu.Lock()
					docVec, ok := docCache[doc.ID]
					mu.Unlock()
					if !ok {
						if algo == support.TdIdf {
							docVec, err = utils.ComputeDocPreIndexedTFIDF(Docs[doc.ID], len(CacheDocs), CacheGrams, normalizeJumps, parallel)
						} else {
							docVec, err = utils.ComputeDocPreIndexedBM25(Docs[doc.ID], len(CacheDocs), CountAllNGrams, CacheGrams, normalizeJumps, parallel)
						}
						if err != nil {
							fmt.Printf("error computing doc %d: %v\n", doc.ID, err)
							return
						}
						mu.Lock()
						docCache[doc.ID] = docVec
						mu.Unlock()
					}

					sim := utils.CosineSimMaps(phraseVec, docVec)
					mu.Lock()
					docSimAccum[doc.ID] = sim
					mu.Unlock()

					progress <- 1
				}(doc)
			}
			wg.Wait()
			close(progress)
		}).Milliseconds()

		// ordena top documentos
		pairs := make([]mgu.Pair[uint16, float64], 0, len(docSimAccum))
		for id, sim := range docSimAccum {
			pairs = append(pairs, mgu.NewPair(id, sim))
		}
		sort.Slice(pairs, func(i, j int) bool { return pairs[i].Right > pairs[j].Right })

		// calcula Spearman médio
		list := mgu.VecMap(pairs, func(t mgu.Pair[uint16, float64]) uint16 { return t.Left })
		spearmanSim, err := utils.Spearman(phrase.Glove, list)
		if err != nil {
			return err
		}

		pushFunc(spearmanSim, elapsedPhrase)
		return nil
	}

	done := 0
	for _, phrase := range data.Words10 {
		if e := processPhrase(phrase, ret.Push10); e != nil {
			return nil, e
		}
		done++
	}
	for _, phrase := range data.Words20 {
		if e := processPhrase(phrase, ret.Push20); e != nil {
			return nil, e
		}
		done++
	}
	for _, phrase := range data.Words40 {
		if e := processPhrase(phrase, ret.Push40); e != nil {
			return nil, e
		}
		done++
	}

	return &ret, nil
}
