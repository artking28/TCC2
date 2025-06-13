package models

import (
	"fmt"
	"gorm.io/gorm"
)

type WordDoc struct {
	ID        uint64 `json:"id"        gorm:"column:id;primary_key;auto_increment;notnull"`
	DocID     uint64 `json:"docId"     gorm:"column:docId;notnull;uniqueIndex:compositeindex"`
	WordID    uint64 `json:"wordId"    gorm:"column:wordId;notnull;uniqueIndex:compositeindex"`
	Frequency uint   `json:"frequency" gorm:"column:frequency;notnull"`

	// Many-to-one
	Word Word `json:"word" gorm:"foreignKey:WordID;references:ID"`
	Doc  Doc  `json:"doc"  gorm:"foreignKey:DocID;references:ID"`
}

func (this* WordDoc) ToString() string {
	return fmt.Sprintf("{ id: %d; word: %s; wordId: %d; docId: %d }", this.ID, this.Word.Value, this.WordID, this.DocID)
}

func (this* WordDoc) TableName() string {
	return "WORD_DOC"
}

func (this *WordDoc) GetId() uint64 {
	return this.ID
}

func (this *WordDoc) BeforeCreate(_ gorm.DB) error {
	return nil
}

func GetInverseIndex(docs []WordDoc) (ret map[string]map[uint64]uint) {
    ret = make(map[string]map[uint64]uint)
    for _, doc := range docs {
        wordKey := doc.Word.Value
        if _, exists := ret[wordKey]; !exists {
            ret[wordKey] = make(map[uint64]uint)
        }
        ret[wordKey][doc.DocID] = doc.Frequency
    }
    return ret
}