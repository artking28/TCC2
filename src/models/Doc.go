package models

import (
	"fmt"
	"regexp"
	"gorm.io/gorm"
)

var docNameRgx = regexp.MustCompile(`[A-Z]{2,3} \d{2}/\d{4}`)

const (
	DocKindText DocKind = "txt"
	DocKindPDF  DocKind = "pdf"
)

type (
	DocKind string

	Doc struct {
		ID      uint64  `json:"id"      gorm:"column:id;primary_key;auto_increment;notnull"`
		Name    string  `json:"name"    gorm:"column:name;type:varchar(20);notnull"`
		Size    uint    `json:"size"    gorm:"column:size;notnull"`
		Kind    DocKind `json:"kind"    gorm:"column:kind;type:varchar(5);notnull"`
		Content []byte  `json:"content" gorm:"-"`
	}
)

func NewDoc(name string, kind DocKind, content []byte) *Doc {
	
	name = docNameRgx.FindString(name)
	return &Doc{
		Name:    name,
		Size:    uint(len(content)),
		Kind:    kind,
		Content: content,
	}
}

func (this* Doc) ToString() string {
	return fmt.Sprintf("{ id: %d; name: %s; size: %d }", this.ID, this.Name, this.Size)
}

func (this* Doc) TableName() string {
	return "DOC"
}

func (this *Doc) GetId() uint64 {
	return this.ID
}

func (this *Doc) BeforeCreate(_ gorm.DB) error {
	return nil
}