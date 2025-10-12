package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	dir = "./misc/corpus/clean"
)

var db *gorm.DB

func initDB() {
	var err error
	db, err = gorm.Open(sqlite.Open("./data.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// AutoMigrate do seu modelo
	err = db.AutoMigrate(
		&models.Document{},
		&models.Word{},
		&models.InverseNGram{},
	)
	if err != nil {
		panic("failed to migrate models")
	}
}

func main() {
	initDB()

	var n int64
	if err := db.Model(&models.Document{}).Count(&n).Error; err != nil {
		log.Fatal(err)
	}
	if n <= 0 {
		if err := InsertAll(); err != nil {
			log.Fatal(err)
		}
	}

	content, err := os.ReadFile("./misc/corpus/clean/doc_0001_clean.txt")
	if err != nil {
		log.Fatal(err)
	}
	text := strings.Fields(string(content))

	ngrams := utils.GetGramsLim(text, 3, 2)
	fmt.Println("ngrams 1:", len(ngrams))

	content, err = os.ReadFile("./misc/corpus/clean/doc_0719_clean.txt")
	if err != nil {
		log.Fatal(err)
	}
	text = strings.Fields(string(content))

	ngrams = utils.GetGramsLim(text, 3, 2)
	fmt.Println("ngrams 719:", len(ngrams))

	fmt.Println("Finished...")
}

func InsertAll() error {

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	return db.Transaction(func(tx *gorm.DB) error {

		var vec []*models.Document
		wordSet := myGoUtils.NewSet[string]()
		for _, f := range files {
			ext := filepath.Ext(f.Name())[1:]
			if f.IsDir() || ext != "txt" {
				continue
			}

			info, e := f.Info()
			if e != nil {
				return e
			}

			doc := models.Document{
				Name: info.Name(),
				Size: info.Size(),
				Kind: models.ParseDocKind(ext),
			}
			vec = append(vec, &doc)

			content, e := os.ReadFile(fmt.Sprintf("%s/%s", dir, f.Name()))
			if e != nil {
				return e
			}
			wordSet.Add(strings.Fields(string(content))...)
		}

		if vec == nil || len(vec) == 0 {
			return nil
		}
		err = tx.Create(vec).Error
		if err != nil {
			return err
		}

		words := myGoUtils.VecMap(wordSet.AsArray(), func(t string) *models.Word {
			return &models.Word{Value: t}
		})
		if words == nil || len(words) == 0 {
			return nil
		}
		return tx.Create(words).Error
	})
}
