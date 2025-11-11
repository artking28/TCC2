package utils

import (
	"log"
	"math"
)

func CosineSimVecs(a, b []float64) float64 {
	if len(a) != len(b) {
		log.Fatal("Vectors must have the same length")
	}

	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}

	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}

func CosineSimMaps(a, b map[string]*float64) float64 {

	var dot, magA, magB float64
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
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}
