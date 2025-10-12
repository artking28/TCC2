package models

import (
	"fmt"

	"gorm.io/gorm"
)

type InverseNGram struct {
	Wd0Id uint64 `gorm:"column:wd0Id;uniqueIndex:compositeindex;notnull"`
	Wd1Id uint64 `gorm:"column:wd1Id;uniqueIndex:compositeindex;"`
	Wd2Id uint64 `gorm:"column:wd2Id;uniqueIndex:compositeindex;"`
	DocId uint64 `gorm:"column:docId;uniqueIndex:compositeindex;notnull"`
	Count uint64 `gorm:"column:count;notnull"`
	Jump0 int8   `gorm:"column:jump0;"`
	Jump1 int8   `gorm:"column:jump1;"`

	Document *Document `gorm:"foreignKey:DocId;references:ID"`
	Wd0      *Word     `gorm:"foreignKey:Wd0Id;references:ID"`
	Wd1      *Word     `gorm:"foreignKey:Wd1Id;references:ID"`
	Wd2      *Word     `gorm:"foreignKey:Wd2Id;references:ID"`
}

func NewInverseNGram() *InverseNGram {
	return &InverseNGram{
		Wd0Id: 0,
		Wd1Id: 0,
		Wd2Id: 0,
		DocId: 0,
		Jump0: -1,
		Jump1: -1,
	}
}

func (this *InverseNGram) GetCacheKey() string {
	j0 := fmt.Sprintf("%d", this.Jump0)
	if this.Jump0 == -1 {
		j0 = "n"
	}
	j1 := fmt.Sprintf("%d", this.Jump1)
	if this.Jump1 == -1 {
		j1 = "n"
	}
	return fmt.Sprintf("%05d-%05d-%05d-%s%s-%04d", this.Wd0Id, this.Wd1Id, this.Wd2Id, j0, j1, this.DocId)
}

func (this *InverseNGram) ToString() string {
	id := fmt.Sprintf("%s-%s-%s", this.Wd0.Value, this.Wd1.Value, this.Wd2.Value)
	return fmt.Sprintf("{ id: %s; count: %d; docId: %d }", id, this.Count, this.DocId)
}

func (this *InverseNGram) TableName() string {
	return "WORD_DOC"
}

func (this *InverseNGram) GetId() uint64 {
	return 0
}

func (this *InverseNGram) BeforeCreate(_ *gorm.DB) error {
	return nil
}
