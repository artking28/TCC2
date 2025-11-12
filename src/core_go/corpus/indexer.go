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
	DbFile = "./../data/data.db"
	Dir    = "./../../misc/corpus/clean"
)

// Variáveis globais
var (
	CountAllNGrams int
	CacheWords     map[string]models.Word                 // Cache em memória de palavras para evitar consultas repetidas
	CacheDocs      map[string]*models.Document            // CacheD em memória de n-gramas
	CacheGrams     map[string]map[uint16]interfaces.IGram // CacheN em memória de n-gramas
	Docs           map[uint16][]interfaces.IGram
)

func CreateDatabaseCaches(id int64, fromScratch bool, gramsSize int, jumpSize int) (string, *gorm.DB) {

	targetFile, db := utils.InitDB(id, max(1, gramsSize%4), DbFile, fromScratch)
	log.Printf("[INFO] Banco inicializado: %s", targetFile)

	var n int64
	err := db.Model(&models.Document{}).Count(&n).Error
	if err != nil {
		log.Fatalf("[ERRO] Falha ao contar documentos: %v", err)
	}
	log.Printf("[INFO] Documentos existentes no banco: %d", n)

	if n <= 0 {
		if err = RegisterDocs(db); err != nil {
			log.Fatalf("[ERRO] Falha ao registrar documentos: %v", err)
		}
		log.Println("[INFO] Documentos registrados com sucesso.")
	}

	DefineCaches(db)
	log.Println("[INFO] Caches definidos.")

	inserted, err := IndexDocsGrams(db, gramsSize, jumpSize)
	if err != nil {
		log.Fatalf("[ERRO] Falha ao indexar documentos: %v", err)
	}
	log.Printf("[INFO] Indexação concluída. Inseridos %d registros.", inserted)

	if inserted <= 0 {
		log.Println("[INFO] Finalizado. Banco vazio.")
		os.Exit(0)
	}

	log.Println("[INFO] Tarefas de banco finalizadas.")
	return targetFile, db
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
	if Docs == nil {
		Docs = make(map[uint16][]interfaces.IGram)
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
				ngram = models.NewInverseUnigram(0, CacheDocs[f.Name()].ID, CacheWords[word[0]].ID)
			case 2:
				ngram = models.NewInverseBigram(0, CacheDocs[f.Name()].ID, CacheWords[word[0]].ID,
					CacheWords[word[1]].ID, jumps[i][0])
			case 3:
				ngram = models.NewInverseTrigram(0, CacheDocs[f.Name()].ID, CacheWords[word[0]].ID,
					CacheWords[word[1]].ID, CacheWords[word[2]].ID, jumps[i][0], jumps[i][1])
			}
			ngram.Increment()

			key := ngram.GetCacheKey(true, false)
			if CacheGrams[key] == nil {
				CacheGrams[key] = make(map[uint16]interfaces.IGram)
			}
			if CacheGrams[key][ngram.GetDocId()] == nil {
				CacheGrams[key][ngram.GetDocId()] = ngram
				Docs[ngram.GetDocId()] = append(Docs[ngram.GetDocId()], ngram)
			}
			CacheGrams[key][ngram.GetDocId()].Increment()
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

	vec0 := mgu.VecMap(mgu.MapValues(CacheGrams), func(t map[uint16]interfaces.IGram) []interfaces.IGram {
		return mgu.MapValues(t)
	})
	vec1, _ := mgu.VecReduce(vec0, func(grams []interfaces.IGram, grams2 []interfaces.IGram) []interfaces.IGram {
		return append(grams, grams2...)
	})

	switch gramsSize {
	case 1:
		all := mgu.VecMap(vec1, func(t interfaces.IGram) *models.InverseUnigram {
			return t.(*models.InverseUnigram)
		})
		return len(all), db.CreateInBatches(all, 1000).Error

	case 2:
		all := mgu.VecMap(vec1, func(t interfaces.IGram) *models.InverseBigram {
			return t.(*models.InverseBigram)
		})
		return len(all), db.CreateInBatches(all, 1000).Error

	case 3:
		all := mgu.VecMap(vec1, func(t interfaces.IGram) *models.InverseTrigram {
			return t.(*models.InverseTrigram)
		})
		return len(all), db.CreateInBatches(all, 1000).Error
	}

	return 0, nil
}
