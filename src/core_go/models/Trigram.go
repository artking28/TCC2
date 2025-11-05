package models

import (
	"fmt"

	"gorm.io/gorm"
)

type InverseTrigram struct {
	Wd0Id uint16 `gorm:"column:wd0Id;uniqueIndex:compositeindex;notnull"`
	Wd1Id uint16 `gorm:"column:wd1Id;uniqueIndex:compositeindex;"`
	Wd2Id uint16 `gorm:"column:wd2Id;uniqueIndex:compositeindex;"`
	DocId uint16 `gorm:"column:docId;uniqueIndex:compositeindex;notnull"`
	Jump0 int8   `gorm:"column:jump0;uniqueIndex:compositeindex;"`
	Jump1 int8   `gorm:"column:jump1;uniqueIndex:compositeindex;"`

	Count uint16 `gorm:"column:count;notnull"`

	Document *Document `gorm:"foreignKey:DocId;references:ID"`
	Wd0      *Word     `gorm:"foreignKey:Wd0Id;references:ID"`
	Wd1      *Word     `gorm:"foreignKey:Wd1Id;references:ID"`
	Wd2      *Word     `gorm:"foreignKey:Wd2Id;references:ID"`
}

func NewInverseNGram() *InverseTrigram {
	return &InverseTrigram{
		Wd0Id: 0,
		Wd1Id: 0,
		Wd2Id: 0,
		DocId: 0,
		Jump0: -1,
		Jump1: -1,
	}
}

func (this *InverseTrigram) GetCacheKey(jumps, doc bool) string {
	j0 := fmt.Sprintf("%d", this.Jump0)
	if this.Jump0 == -1 {
		j0 = "n"
	}
	j1 := fmt.Sprintf("%d", this.Jump1)
	if this.Jump1 == -1 {
		j1 = "n"
	}
	ret := fmt.Sprintf("%05d-%05d-%05d", this.Wd0Id, this.Wd1Id, this.Wd2Id)
	if jumps {
		ret = fmt.Sprintf("%s-%s%s", ret, j0, j1)
	}
	if doc {
		ret = fmt.Sprintf("%s-%04d", ret, this.DocId)
	}
	return ret
}

func (this *InverseTrigram) ToString() string {
	id := fmt.Sprintf("%s-%s-%s", this.Wd0.Value, this.Wd1.Value, this.Wd2.Value)
	return fmt.Sprintf("{ id: %s; count: %d; docId: %d }", id, this.Count, this.DocId)
}

func (this *InverseTrigram) TableName() string {
	return "WORD_DOC"
}

func (this *InverseTrigram) GetId() uint64 {
	return 0
}

func (this *InverseTrigram) BeforeCreate(_ *gorm.DB) error {
	return nil
}
