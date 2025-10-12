package utils

import "fmt"

// Spearman returns the Spearman rank correlation (ρ) between two rankings of the same items.
// Requires: a and b contain exactly the same elements (permutations), no ties.
// ρ ∈ [-1, 1]. Returns error if sizes differ, n < 2, or items mismatch.
func Spearman[C comparable](a, b []C) (float64, error) {
	n := len(a)
	if n*len(b) == 0 {
		return 0, fmt.Errorf("empty slices not supported")
	}
	if n != len(b) {
		return 0, fmt.Errorf("slices must have same length")
	}
	if n < 2 {
		return 1, nil
	}

	// rankA maps item -> rank (1..n)
	rankA := make(map[C]int, n)
	for i, v := range a {
		if _, dup := rankA[v]; dup {
			return 0, fmt.Errorf("ties not supported (duplicate in a)")
		}
		rankA[v] = i + 1
	}

	// compute sum of squared rank differences using positions in b
	var sumD2 float64
	seen := make(map[C]struct{}, n)
	for j, v := range b {
		if _, dup := seen[v]; dup {
			return 0, fmt.Errorf("ties not supported (duplicate in b)")
		}
		seen[v] = struct{}{}

		ra, ok := rankA[v]
		if !ok {
			return 0, fmt.Errorf("item %v in b not found in a", v)
		}
		rb := j + 1
		d := float64(ra - rb)
		sumD2 += d * d
	}

	// ρ = 1 - (6 Σ d_i^2) / (n (n^2 - 1))
	den := float64(n * (n*n - 1))
	rho := 1.0 - (6.0*sumD2)/den
	return rho, nil
}
