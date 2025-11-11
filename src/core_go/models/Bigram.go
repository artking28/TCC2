package models

import (
	"fmt"

	"gorm.io/gorm"
)

type InverseBigram struct {
	Wd0Id uint16 `gorm:"column:wd0Id;uniqueIndex:compositeindex;notnull"`
	Wd1Id uint16 `gorm:"column:wd1Id;uniqueIndex:compositeindex;"`
	DocId uint16 `gorm:"column:docId;uniqueIndex:compositeindex;notnull"`
	Jump0 int8   `gorm:"column:jump0;uniqueIndex:compositeindex;"`

	Count int `gorm:"column:count;notnull"`

	Document *Document `gorm:"foreignKey:DocId;references:ID"`
	Wd0      *Word     `gorm:"foreignKey:Wd0Id;references:ID"`
	Wd1      *Word     `gorm:"foreignKey:Wd1Id;references:ID"`
}

func NewInverseBigram(size int, docID, wdId0, wdId1 uint16, jump0 int8) *InverseBigram {
	return &InverseBigram{
		DocId: docID,
		Wd0Id: wdId0,
		Wd1Id: wdId1,
		Jump0: jump0,
		Count: size,
	}
}

func (this *InverseBigram) GetCacheKey(jumps, doc bool) string {
	j0 := fmt.Sprintf("%d", this.Jump0)
	if this.Jump0 == -1 {
		j0 = "n"
	}
	ret := fmt.Sprintf("%05d-%05d", this.Wd0Id, this.Wd1Id)
	if jumps {
		ret = fmt.Sprintf("%s-%s", ret, j0)
	}
	if doc {
		ret = fmt.Sprintf("%s-%04d", ret, this.DocId)
	}
	return ret
}

func (this *InverseBigram) GetDocId() uint16 {
	return this.DocId
}

func (this *InverseBigram) Increment() {
	this.Count++
}

func (this *InverseBigram) GetCount() int {
	return this.Count
}

func (this *InverseBigram) ApplyWordWheres(db *gorm.DB) *gorm.DB {
	return db.Where("wd0Id = ? AND wd1Id = ?", this.Wd0Id, this.Wd1Id)
}

func (this *InverseBigram) ApplyJumpWheres(db *gorm.DB) *gorm.DB {
	return db.Where("jump0 = ?", this.Jump0)
}

func (this *InverseBigram) ToString() string {
	id := fmt.Sprintf("%s-%s", this.Wd0.Value, this.Wd1.Value)
	return fmt.Sprintf("{ id: %s; count: %d; docId: %d }", id, this.Count, this.DocId)
}

func (this *InverseBigram) TableName() string {
	return "WORD_DOC"
}

func (this *InverseBigram) GetId() uint64 {
	return 0
}

func (this *InverseBigram) BeforeCreate(_ *gorm.DB) error {
	return nil
}
