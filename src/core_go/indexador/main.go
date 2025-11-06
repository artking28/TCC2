package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	mgu "github.com/artking28/myGoUtils" // Biblioteca customizada para utilitários, como Set
	"github.com/tcc2-davi-arthur/models" // Models do projeto: Document, Word, InverseTrigram
	"github.com/tcc2-davi-arthur/utils"
	"gorm.io/driver/sqlite" // Driver SQLite do GORM
	"gorm.io/gorm"          // ORM GORM
)

// Diretório onde os arquivos de texto serão lidos
const (
	DbFile = "./data.db"
	Dir    = "./misc/corpus/clean"
)

// Variáveis globais
var (
	CountAllNGrams uint32
	Cache          map[string]models.Word                       // Cache em memória de palavras para evitar consultas repetidas
	CacheN         map[string]map[uint16]*models.InverseTrigram // CacheN em memória de n-gramas
	CacheD         map[string]*models.Document                  // CacheD em memória de n-gramas
	Db             *gorm.DB                                     // Conexão com o banco de dados
)

func main() {

	initDB() // Inicializa o banco

	// Verifica se existem documentos no banco
	var n int64
	err := Db.Model(&models.Document{}).Count(&n).Error
	if err != nil {
		log.Fatal(err)
	}

	// Se não houver documentos, insere todos os arquivos do diretório e suas palavras
	if n <= 0 {
		if err = InsertAll(); err != nil {
			log.Fatal(err)
		}
	}

	// Sempre q roda cria um backup
	//if err = utils.DuplicateFile(DbFileBackup, DbFile); err != nil {
	//	log.Fatal(err)
	//}
	//Db, err = gorm.Open(sqlite.Open(DbFile), &gorm.Config{})
	//if err != nil {
	//	panic("failed to connect database")
	//}

	// Inicializa Cache em memória com todas as palavras
	if len(Cache) == 0 {
		if Cache == nil {
			Cache = make(map[string]models.Word)
		}

		var vec []*models.Word
		if err = Db.Model(&models.Word{}).Find(&vec).Error; err != nil {
			log.Fatal(err)
		}

		// Preenche o cache
		for _, word := range vec {
			Cache[word.Value] = *word
		}
	}

	// Inicializa cacheD em memória com todos os documentos
	if len(CacheD) == 0 {
		if CacheD == nil {
			CacheD = make(map[string]*models.Document)
		}

		var vec []*models.Document
		if err = Db.Model(&models.Document{}).Find(&vec).Error; err != nil {
			log.Fatal(err)
		}

		// Preenche o cache
		for _, doc := range vec {
			CacheD[doc.Name] = doc
		}
	}

	if err = IndexDocs(); err != nil {
		log.Fatal(err)
	}

	err = Db.Model(&models.InverseTrigram{}).Count(&n).Error
	if err != nil {
		log.Fatal(err)
	}

	if n <= 0 {
		fmt.Println("Finished... [empty database]")
		os.Exit(0)
	}

	start := time.Now()
	vec, err := utils.ComputePosIndexedTFIDF(1, len(CacheD), Db, true, false)
	if err != nil {
		log.Fatal(err)
	}
	println(vec)
	fmt.Println("PreIndexed elapsed:", time.Since(start))

	fmt.Println("Finished...")
}

// Inicializa o banco de dados e faz a migração automática dos modelos
func initDB() {
	var err error
	Db, err = gorm.Open(sqlite.Open(DbFile), &gorm.Config{})
	if err != nil {
		println("failed to connect database")
		log.Fatal(err)
	}

	// AutoMigrate cria as tabelas para os modelos, se não existirem
	err = Db.AutoMigrate(
		&models.Document{},
		&models.Word{},
		&models.InverseTrigram{},
	)
	if err != nil {
		println("failed to migrate models")
		log.Fatal(err)
	}

	err = Db.Exec(`CREATE INDEX IF NOT EXISTS idx_worddoc_wdids ON WORD_DOC (wd0_id, wd1_id, wd2_id);`).Error
	if err != nil {
		println("failed to create index")
		log.Fatal(err)
	}
}

// InsertAll Lê todos os arquivos de texto do diretório, cria documentos e palavras, e insere no banco
func InsertAll() error {

	files, err := os.ReadDir(Dir) // Lista os arquivos no diretório
	if err != nil {
		log.Fatal(err)
	}

	// Transaction garante que tudo seja inserido de forma atômica
	return Db.Transaction(func(tx *gorm.DB) error {

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
