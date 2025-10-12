package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	mgu "github.com/artking28/myGoUtils" // Biblioteca customizada para utilitários, como Set
	"github.com/tcc2-davi-arthur/models" // Models do projeto: Document, Word, InverseNGram
	"gorm.io/driver/sqlite"              // Driver SQLite do GORM
	"gorm.io/gorm"                       // ORM GORM
)

// Diretório onde os arquivos de texto serão lidos
const dir = "./misc/corpus/clean"

// Variáveis globais
var (
	cache  map[string]models.Word         // Cache em memória de palavras para evitar consultas repetidas
	cacheN map[string]models.InverseNGram // Cache em memória de n-gramas
	db     *gorm.DB                       // Conexão com o banco de dados
)

func main() {
	initDB() // Inicializa o banco

	// Verifica se existem documentos no banco
	var n int64
	if err := db.Model(&models.Document{}).Count(&n).Error; err != nil {
		log.Fatal(err)
	}

	// Se não houver documentos, insere todos os arquivos do diretório
	if n <= 0 {
		if err := InsertAll(); err != nil {
			log.Fatal(err)
		}
	}

	// Inicializa cache em memória com todas as palavras
	if len(cache) == 0 {
		if cache == nil {
			cache = make(map[string]models.Word)
		}

		var vec []*models.Word
		if err := db.Model(&models.Word{}).Find(&vec).Error; err != nil {
			log.Fatal(err)
		}

		// Preenche o cache
		for _, word := range vec {
			cache[word.Value] = *word
		}
	}

	fmt.Println("Finished...")
}

// Inicializa o banco de dados e faz a migração automática dos modelos
func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// AutoMigrate cria as tabelas para os modelos, se não existirem
	err = db.AutoMigrate(
		&models.Document{},
		&models.Word{},
		&models.InverseNGram{},
	)
	if err != nil {
		panic("failed to migrate models")
	}
}

// InsertAll Lê todos os arquivos de texto do diretório, cria documentos e palavras, e insere no banco
func InsertAll() error {

	files, err := os.ReadDir(dir) // Lista os arquivos no diretório
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
				Size: info.Size(),
				Kind: models.ParseDocKind(ext),
			}
			vec = append(vec, &doc)

			// Lê o conteúdo do arquivo e adiciona palavras ao conjunto
			content, e := os.ReadFile(fmt.Sprintf("%s/%s", dir, f.Name()))
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
