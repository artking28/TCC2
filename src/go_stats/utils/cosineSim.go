package utils

import (
	"log"
	"math"
)

type Float interface{ float64 | float32 }

func CosineSimVecs[F Float](a, b []F) F {
	if len(a) != len(b) {
		log.Fatal("Vectors must have the same length")
	}

	var dot, magA, magB F
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	rDot, ra, rb := float64(dot), float64(magA), float64(magB)
	return F(rDot / (math.Sqrt(ra) * math.Sqrt(rb)))
}

func CosineSimMaps[F Float](a, b map[string]*F) F {

	var dot, magA, magB F
	for k, va := range a {
		if va == nil {
			continue
		}
		magA += (*va) * (*va)
		if vb, ok := b[k]; ok && vb != nil {
			dot += (*va) * (*vb)
		}
	}

	for _, vb := range b {
		if vb != nil {
			magB += (*vb) * (*vb)
		}
	}

	if magA == 0 || magB == 0 {
		return 0
	}

	rDot, ra, rb := float64(dot), float64(magA), float64(magB)
	return F(rDot / (math.Sqrt(ra) * math.Sqrt(rb)))
}
