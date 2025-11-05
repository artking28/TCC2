package models

import (
	"fmt"

	"gorm.io/gorm"
)

type Word struct {
	ID    uint16 `json:"id"    gorm:"column:id;primary_key;auto_increment;notnull"`
	Value string `json:"value" gorm:"value:kind;type:varchar(30);notnull"`
}

func NewWord(value string) *Word {
	return &Word{
		Value: value,
	}
}

func (this *Word) ToString() string {
	return fmt.Sprintf("{ id: %d; value: %s }", this.ID, this.Value)
}

func (this *Word) TableName() string {
	return "WORD"
}

func (this *Word) GetId() uint16 {
	return this.ID
}

func (this *Word) BeforeCreate(_ *gorm.DB) error {
	return nil
}
