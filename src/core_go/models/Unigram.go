package models

import (
	"fmt"

	"gorm.io/gorm"
)

type InverseUnigram struct {
	Wd0Id uint16 `gorm:"column:wd0Id;uniqueIndex:compositeindex;notnull"`
	DocId uint16 `gorm:"column:docId;uniqueIndex:compositeindex;notnull"`

	Count uint16 `gorm:"column:count;notnull"`

	Document *Document `gorm:"foreignKey:DocId;references:ID"`
	Wd0      *Word     `gorm:"foreignKey:Wd0Id;references:ID"`
}

func NewInverseUnigram() *InverseUnigram {
	return &InverseUnigram{
		Wd0Id: 0,
		DocId: 0,
	}
}

func (this *InverseUnigram) GetCacheKey(_, doc bool) string {
	ret := fmt.Sprintf("%05d", this.Wd0Id)
	if doc {
		ret = fmt.Sprintf("%s-%04d", ret, this.DocId)
	}
	return ret
}

func (this *InverseUnigram) GetDocId() uint16 {
	return this.DocId
}

func (this *InverseUnigram) Increment() {
	this.Count++
}

func (this *InverseUnigram) ToString() string {
	return fmt.Sprintf("{ id: %s; count: %d; docId: %d }", this.Wd0.Value, this.Count, this.DocId)
}

func (this *InverseUnigram) TableName() string {
	return "WORD_DOC"
}

func (this *InverseUnigram) GetId() uint64 {
	return 0
}

func (this *InverseUnigram) BeforeCreate(_ *gorm.DB) error {
	return nil
}
