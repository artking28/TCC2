package repository

import (
	"fmt"

	"github.com/tcc2-davi-arthur/models"
	"github.com/tcc2-davi-arthur/models/interfaces"
	"gorm.io/gorm"
)

type GramRepository struct {
	db *gorm.DB
}

func NewGramRepository(db *gorm.DB) *GramRepository {
	return &GramRepository{db: db}
}

func (r *GramRepository) FindByDocAndSize(docID uint16, gramSize int) ([]interfaces.IGram, error) {

	var label string
	var data any
	switch gramSize {
	case 1:
		label = "wd0Id"
		data = []models.InverseUnigram{}
		break
	case 2:
		label = "wd0Id, wd1Id, jump0"
		data = []models.InverseBigram{}
		break
	case 3:
		label = "wd0Id, wd1Id, wd2Id, jump0, jump1"
		data = []models.InverseTrigram{}
		break
	default:
		return nil, fmt.Errorf("unsupported gram size: %d", gramSize)
	}

	err := r.db.Table("WORD_DOC").
		Select(fmt.Sprintf("%s AS wd0Id, COUNT(docId) AS count", label)).
		Where("docId = ?", docID).
		Group(label).
		Find(&data).Error
	if err != nil {
		return nil, err
	}

	var ret []interfaces.IGram
	switch gramSize {
	case 1:
		for _, one := range data.([]models.InverseUnigram) {
			ret = append(ret, &one)
		}
		break
	case 2:
		for _, one := range data.([]models.InverseBigram) {
			ret = append(ret, &one)
		}
		break
	case 3:
		for _, one := range data.([]models.InverseTrigram) {
			ret = append(ret, &one)
		}
		break
	}
	return ret, nil
}
