package corpus

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mgu "github.com/artking28/myGoUtils" // Biblioteca customizada para utilitários, como Set
	"github.com/tcc2-davi-arthur/models" // Models do projeto: Document, Word, InverseTrigram
	"github.com/tcc2-davi-arthur/models/interfaces"
	"github.com/tcc2-davi-arthur/utils"
	"gorm.io/gorm" // ORM GORM
)

// Diretório onde os arquivos de texto serão lidos
const (
	DbFile = "./data.db"
	Dir    = "./misc/corpus/clean"
)

// Variáveis globais
var (
	CountAllNGrams uint32
	CacheWords     map[string]models.Word                 // Cache em memória de palavras para evitar consultas repetidas
	CacheDocs      map[string]*models.Document            // CacheD em memória de n-gramas
	CacheGrams     map[string]map[uint16]interfaces.IGram // CacheN em memória de n-gramas
)

func CreateDatabaseCaches(id int64, fromScratch bool, gramsSize int, jumpSize int) string {

	targetFile, db := utils.InitDB(id, max(1, gramsSize%4), DbFile, fromScratch) // Inicializa o banco

	// Verifica se existem documentos no banco
	var n int64
	err := db.Model(&models.Document{}).Count(&n).Error
	if err != nil {
		log.Fatal(err)
	}

	// Se não houver documentos, insere todos os arquivos do diretório e suas palavras
	if n <= 0 {
		if err = RegisterDocs(db); err != nil {
			log.Fatal(err)
		}
	}

	DefineCaches(db)

	inserted, err := IndexDocsGrams(db, gramsSize, jumpSize)
	if err != nil {
		log.Fatal(err)
	}

	if inserted <= 0 {
		fmt.Println("Finished... [empty database]")
		os.Exit(0)
	}

	//if algo == support.TdIdf {
	//	if preIndexed {
	//		_, err = utils.ComputePreIndexedTFIDF(1, len(CacheDocs), db, normalizeJumps, pa)
	//	} else {
	//		_, err = utils.ComputePosIndexedTFIDF(1, len(CacheDocs), db, normalizeJumps, false)
	//	}
	//} else {
	//	if preIndexed {
	//		_, err = utils.ComputePreIndexedBM25(1, len(CacheDocs), db, normalizeJumps, false)
	//	} else {
	//		_, err = utils.ComputePosIndexedBM25(1, len(CacheDocs), db, normalizeJumps, false)
	//	}
	//}
	//if err != nil {
	//	log.Fatal(err)
	//}

	return targetFile
}

// RegisterDocs Lê todos os arquivos de texto do diretório, cria documentos e palavras, e insere no banco
func RegisterDocs(db *gorm.DB) error {

	files, err := os.ReadDir(Dir) // Lista os arquivos no diretório
	if err != nil {
		log.Fatal(err)
	}

	// Transaction garante que tudo seja inserido de forma atômica
	return db.Transaction(func(tx *gorm.DB) error {

		var vec []*models.Document
		wordSet := mgu.NewSet[string]() // Conjunto para armazenar palavras únicas

		for _, f := range files {
			ext := filepath.Ext(f.Name())[1:] // Pega a extensão do arquivo sem o ponto
			if f.IsDir() || ext != "txt" {    // Pula diretórios e arquivos não-txt
				continue
			}

			info, e := f.Info()
			if e != nil {
				return e
			}

			// Cria um documento com nome, tamanho e tipo
			doc := models.Document{
				Name: info.Name(),
				Size: uint16(info.Size()),
				Kind: models.ParseDocKind(ext),
			}
			vec = append(vec, &doc)

			// Lê o conteúdo do arquivo e adiciona palavras ao conjunto
			content, e := os.ReadFile(fmt.Sprintf("%s/%s", Dir, f.Name()))
			if e != nil {
				return e
			}
			wordSet.Add(strings.Fields(string(content))...)
		}

		if vec == nil || len(vec) == 0 {
			return nil // Não há documentos para inserir
		}

		// Insere os documentos no banco
		err = tx.Create(vec).Error
		if err != nil {
			return err
		}

		// Converte o conjunto de palavras em slice de modelos Word
		words := mgu.VecMap(wordSet.AsArray(), func(t string) *models.Word {
			return &models.Word{Value: t}
		})
		if words == nil || len(words) == 0 {
			return nil
		}

		// Insere as palavras no banco
		return tx.Create(words).Error
	})
}

func DefineCaches(db *gorm.DB) {
	// Inicializa Cache em memória com todas as palavras
	if len(CacheWords) == 0 {
		if CacheWords == nil {
			CacheWords = make(map[string]models.Word)
		}

		var vec []*models.Word
		if err := db.Model(&models.Word{}).Find(&vec).Error; err != nil {
			log.Fatal(err)
		}

		// Preenche o cache
		for _, word := range vec {
			CacheWords[word.Value] = *word
		}
	}

	// Inicializa cacheD em memória com todos os documentos
	if len(CacheDocs) == 0 {
		if CacheDocs == nil {
			CacheDocs = make(map[string]*models.Document)
		}

		var vec []*models.Document
		if err := db.Model(&models.Document{}).Find(&vec).Error; err != nil {
			log.Fatal(err)
		}

		// Preenche o cache
		for _, doc := range vec {
			CacheDocs[doc.Name] = doc
		}
	}

}

func IndexDocsGrams(db *gorm.DB, gramsSize, jumpSize int) (int, error) {
	gramsSize = max(1, gramsSize%4)

	if CacheGrams == nil {
		CacheGrams = make(map[string]map[uint16]interfaces.IGram)
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
			return 0, err
		}
		text := strings.Fields(string(content))

		result, jumps := utils.GetGramsLim(text, gramsSize, jumpSize)
		for i, word := range result {
			var ngram interfaces.IGram

			switch gramsSize {
			case 1:
				ngram = &models.InverseUnigram{
					DocId: CacheDocs[f.Name()].ID,
					Wd0Id: CacheWords[word[0]].ID,
				}
			case 2:
				ngram = &models.InverseBigram{
					DocId: CacheDocs[f.Name()].ID,
					Wd0Id: CacheWords[word[0]].ID,
					Wd1Id: CacheWords[word[1]].ID,
					Jump0: jumps[i][0],
				}
			case 3:
				ngram = &models.InverseTrigram{
					DocId: CacheDocs[f.Name()].ID,
					Wd0Id: CacheWords[word[0]].ID,
					Wd1Id: CacheWords[word[1]].ID,
					Wd2Id: CacheWords[word[2]].ID,
					Jump0: jumps[i][0],
					Jump1: jumps[i][1],
				}
			}
			ngram.Increment()

			key := ngram.GetCacheKey(true, false)
			if CacheGrams[key] != nil && CacheGrams[key][ngram.GetDocId()] != nil {
				CacheGrams[key][ngram.GetDocId()].Increment()
				CountAllNGrams++
				continue
			}
			CacheGrams[key] = make(map[uint16]interfaces.IGram)
			CacheGrams[key][ngram.GetDocId()] = ngram
			CountAllNGrams++
		}
	}

	// Verifica se existem documentos no banco
	var n int64
	err = db.Table("WORD_DOC").Count(&n).Error
	if err != nil {
		log.Fatal(err)
	}

	// Se não houver documentos, insere todos os arquivos do diretório e suas palavras
	if n > 0 {
		return 0, nil
	}

	all := mgu.MapValues(CacheGrams)
	return len(all), db.CreateInBatches(all, 1000).Error
}
