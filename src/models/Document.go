package models

import (
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

var docNameRgx = regexp.MustCompile(`[A-Z]{2,3} \d{2}/\d{4}`)

const (
	DocKindNone DocKind = "none"
	DocKindText DocKind = "txt"
	DocKindPDF  DocKind = "pdf"
)

func ParseDocKind(s string) DocKind {
	switch s {
	case "txt":
		return DocKindText
	case "pdf":
		return DocKindPDF
	default:
		return DocKindNone
	}
}

type (
	DocKind string

	Document struct {
		ID      uint16  `json:"id"      gorm:"column:id;primary_key;auto_increment;notnull"`
		Name    string  `json:"name"    gorm:"column:name;type:varchar(20);notnull"`
		Size    uint16  `json:"size"    gorm:"column:size;notnull"`
		Kind    DocKind `json:"kind"    gorm:"column:kind;type:varchar(5);notnull"`
		Content []byte  `json:"content" gorm:"-"`
	}
)

func NewDoc(name string, kind DocKind, content []byte) *Document {

	name = docNameRgx.FindString(name)
	return &Document{
		Name:    name,
		Size:    uint16(len(content)),
		Kind:    kind,
		Content: content,
	}
}

func (this *Document) ToString() string {
	return fmt.Sprintf("{ id: %d; name: %s; size: %d }", this.ID, this.Name, this.Size)
}

func (this *Document) TableName() string {
	return "DOCUMENT"
}

func (this *Document) GetId() uint16 {
	return this.ID
}

func (this *Document) BeforeCreate(_ *gorm.DB) error {
	return nil
}
