package utils

import (
	"log"
)

//func main() {
//  size, jump := 3, 2
//  vec := "abcdefghijklmnopqrstuvwxyz"
//  fmt.Println(vec)
//
//  res := getGrams([]byte(vec), size, jump)
//  for _, g := range res {
//     fmt.Println(string(g))
//  }
//
//  res = getGrams([]byte(vec), size, jump)
//  for _, g := range res {
//     fmt.Println(string(g))
//  }
//}

// GetGramsLim Generates n-grams explicitly/limitedly (max 3-grams, max 2 jumps).
// [T any] defines the function as generic, accepting slices of any type T.
func GetGramsLim[T any](vec []T, size, jump int) ([][]T, [][]int8) {
	n := len(vec)
	if size <= 0 || size > 3 {
		log.Fatalln("Error: getGramsLim supports a maximum of 3-grams and minimum of 1.")
	}

	var ret [][]T
	var retJumps [][]int8

	// Base Case: size=1 (Unigrams)
	if size == 1 {
		// Generates all unigrams (each element is a gram of size 1).
		for i := 0; i < n; i++ {
			// vec[i:i+1] creates a slice of a single element.
			ret = append(ret, vec[i:i+1])
		}
		return ret, nil
	}

	// Case size=2 (Bigrams with Jumps)
	if size == 2 {
		for i := 0; i < n-1; i++ {
			for j := 1; j <= jump+1 && i+j < n; j++ {
				// Creates the gram: [element at 'i', element at 'i+j']
				gram := []T{vec[i], vec[i+j]}
				ret = append(ret, gram)
				retJumps = append(retJumps, []int8{int8(j)})
			}
		}
		return ret, retJumps
	}

	// Case size=3 (Trigrams with Jumps)
	// The triple loop generates trigrams [vec[i], vec[i+j], vec[i+j+k]].
	for i := 0; i < n-2; i++ {
		for j := 1; j <= jump+1 && i+j < n-1; j++ {
			for k := 1; k <= jump+1 && i+j+k < n; k++ {
				// Creates the 3-element gram.
				gram := []T{vec[i], vec[i+j], vec[i+j+k]}
				ret = append(ret, gram)
				retJumps = append(retJumps, []int8{int8(j), int8(k)})
			}
		}
	}

	return ret, retJumps
}

// GetGrams Generates all possible n-grams with jumps up to "jump".
// This is a recursive (DFS) and general implementation, without the restrictions of GetGramsLim.
func GetGrams[T any](vec []T, size, jump int) [][]T {
	n := len(vec)
	if size <= 0 || n < size {
		return nil
	}

	var ret [][]T

	// DFS (Depth-First Search) Recursive
	// dfs is an anonymous function that implements the gram generation algorithm.
	// start: The index of the last element added to the gram.
	// depth: The current position in the gram being built (from 0 to size-1).
	// curr: The gram currently being built (temporary slice).
	var dfs func(start int, depth int, curr []T)
	dfs = func(start, depth int, curr []T) {

		// Stop Condition (Complete Gram)
		if depth == size {
			tmp := make([]T, size)
			copy(tmp, curr)
			ret = append(ret, tmp)
			return
		}

		// Tries all possible next positions (j=0 is jump 1, j=jump is jump+1)
		for j := 0; j <= jump; j++ {
			next := start + j + 1
			if next >= n {
				break
			}
			curr[depth] = vec[next]
			dfs(next, depth+1, curr)
		}
	}

	// Each element of the vector can be the 1st element of an n-gram.
	for i := 0; i < n; i++ {
		curr := make([]T, size)
		curr[0] = vec[i]
		dfs(i, 1, curr)
	}

	return ret
}
