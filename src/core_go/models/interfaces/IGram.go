package interfaces

import (
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type IGram interface {
	GetCacheKey(jumps, doc bool) string
	GetDocId() uint16
	Increment()
	GetCount() int
	ApplyWordWheres(db *gorm.DB) *gorm.DB
	ApplyJumpWheres(db *gorm.DB) *gorm.DB

	schema.Tabler
}
