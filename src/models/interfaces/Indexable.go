package interfaces

import (
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/schema"

	"github.com/google/uuid"
)

type (
	UniqueID interface {
		uint16 | uuid.UUID
	}

	Indexable[ID UniqueID] interface {
		GetId() ID
		ToString() string

		schema.Tabler
		callbacks.BeforeCreateInterface
	}
)
