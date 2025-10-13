package models

import (
	"time"
)

const (
	TF_IDF Algo = "tf_idf"
	BM_25  Algo = "bm_25"
)

type (
	Algo string

	StatEntry struct {
		CreatedAt   time.Time         `json:"createdAt,omitempty"`
		PreIndexed  bool              `json:"preIndexed,omitempty"`
		DurationMs  int64             `json:"durationMs,omitempty"`
		Algorithm   Algo              `json:"algorithm,omitempty"`
		Docs        uint16            `json:"docs,omitempty"`
		Words       uint32            `json:"words,omitempty"`
		TotalNgrams uint32            `json:"totalNgrams,omitempty"`
		NSize       uint8             `json:"nSize,omitempty"`
		Jump0       uint8             `json:"jump0,omitempty"`
		Jump1       uint8             `json:"jump1,omitempty"`
		Metadata    map[string]string `json:"metadata,omitempty"`
	}
)
