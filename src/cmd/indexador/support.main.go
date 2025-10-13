package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/utils"
)

func IndexDocs() error {

	if CacheN == nil {
		CacheN = make(map[string]*models.InverseTrigram)
	}

	files, err := os.ReadDir(Dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		ext := filepath.Ext(f.Name())[1:]
		if f.IsDir() || ext != "txt" {
			continue
		}

		content, err := os.ReadFile(fmt.Sprintf("%s/%s", Dir, f.Name()))
		if err != nil {
			return err
		}
		text := strings.Fields(string(content))

		result, jumps := utils.GetGramsLim(text, 3, 2)
		for i, word := range result {
			ngram := models.InverseTrigram{
				DocId: CacheD[f.Name()].ID,
				Wd0Id: Cache[word[0]].ID,
				Wd1Id: Cache[word[1]].ID,
				Wd2Id: Cache[word[2]].ID,
				Jump0: jumps[i][0],
				Jump1: jumps[i][1],
				Count: 1,
			}
			if CacheN[ngram.GetCacheKey(true, true)] != nil {
				CacheN[ngram.GetCacheKey(true, true)].Count++
				CountAllNGrams++
				continue
			}
			CacheN[ngram.GetCacheKey(true, true)] = &ngram
			CountAllNGrams++
		}
	}

	// Verifica se existem documentos no banco
	var n int64
	err = Db.Model(&models.InverseTrigram{}).Count(&n).Error
	if err != nil {
		log.Fatal(err)
	}

	// Se não houver documentos, insere todos os arquivos do diretório e suas palavras
	if n > 0 {
		return nil
	}

	all := mgu.MapValues(CacheN)
	return Db.CreateInBatches(all, 1000).Error
}
