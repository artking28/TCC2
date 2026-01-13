package utils

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tcc2-davi-arthur/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB inicializa o banco de dados e faz a migração automática dos modelos
func InitDB(id int64, gramSize int, dbFile string, fromScratch bool) (string, *gorm.DB) {
	var targetFile string

	if fromScratch {
		if err := os.Remove(dbFile); err != nil && !os.IsNotExist(err) {
			println("failed to delete database")
			log.Fatal(err)
		}
		targetFile = dbFile
	} else {
		newName := fmt.Sprintf("%s_%d.db", strings.TrimSuffix(dbFile, ".db"), id)
		err := DuplicateFile(dbFile, newName)
		if err != nil {
			println("failed to duplicate database")
			log.Fatal(err)
		}
		targetFile = newName
	}

	ret, err := gorm.Open(sqlite.Open(targetFile), &gorm.Config{})
	if err != nil {
		println("failed to connect database")
		log.Fatal(err)
	}

	var gramModel any
	var index string

	switch gramSize {
	case 1:
		gramModel = &models.InverseUnigram{}
		index = "(wd0Id)"
		break
	case 2:
		gramModel = &models.InverseBigram{}
		index = "(wd0Id, wd1Id)"
		break
	case 3:
		gramModel = &models.InverseTrigram{}
		index = "(wd0Id, wd1Id, wd2Id)"
		break
	}

	// AutoMigrate cria as tabelas para os modelos, se não existirem
	err = ret.AutoMigrate(
		&models.Document{},
		&models.Word{},
		&gramModel,
	)
	if err != nil {
		println("failed to migrate models")
		log.Fatal(err)
	}

	query := fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_worddoc_wdids ON WORD_DOC %s;", index)
	err = ret.Exec(query).Error
	if err != nil {
		println("failed to create index")
		log.Fatal(err)
	}

	return targetFile, ret
}
