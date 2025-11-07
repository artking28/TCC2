package interfaces

import "gorm.io/gorm"

type IGram interface {
	GetCacheKey(jumps, doc bool) string
	GetDocId() uint16
	Increment()
	GetCount() int
	ApplyWordWheres(db *gorm.DB) *gorm.DB
	ApplyJumpWheres(db *gorm.DB) *gorm.DB
}
