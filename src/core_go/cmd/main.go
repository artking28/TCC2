package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tcc2-davi-arthur/corpus"
	"github.com/tcc2-davi-arthur/models/support"
)

// ResultsOutput represents path to the file where benchmark results will be saved.
const ResultsOutput = "./misc/results.txt"

// N-gram Size Constants
const (
	Unigram = 1 // Represents a 1-word n-gram.
	Bigram  = 2 // Represents a 2-word n-gram (consecutive words or with jumps).
	Trigram = 3 // Represents a 3-word n-gram (consecutive words or with jumps).
)

// Maximum Jump Limit Constants
const (
	MaxTrigramJumps = 2 // Maximum allowed jumps between words for a Trigram.
	MaxBigramJumps  = 4 // Maximum allowed jumps between words for a Bigram.
)

// main iterates over all parameter combinations and saves results to a JSON file.
func main() {

	//corpus.StartScrapping(5000, 25, 500*time.Millisecond)

	//corpus.TextProcessor(25)

	var id int64 = 1 // Unique counter to identify each test.

	//// Defines n-gram sizes to be tested and their respective maximum jump limits.
	//sizes := []mgu.Pair[int, int]{
	//	mgu.NewPair(Unigram, 0),               // Unigram: 0 jumps (always).
	//	mgu.NewPair(Bigram, MaxBigramJumps),   // Bigram: up to 4 jumps.
	//	mgu.NewPair(Trigram, MaxTrigramJumps), // Trigram: up to 2 jumps.
	//}

	strB := strings.Builder{}

	//// Main loop: iterates through each configured n-gram type (Unigram, Bigram, Trigram).
	//for _, s := range sizes {
	//	size, maxJumps := s.Left, s.Right // Unpacks n-gram size and jump limit.
	//
	//	// For each n-gram size, test all possible jump levels (from 0 to maxJumps).
	//	for jump := 0; jump <= maxJumps; jump++ {
	//
	//		// Test both variations: with and without pre-indexing (in-memory cache vs database query).
	//		for _, preIndexed := range []bool{false, true} {
	//
	//			// For each pre-index mode, test jump normalization on and off.
	//			for _, normalize := range []bool{false, true} {
	//
	//				// For each parallel mode, test multithreading.
	//				for _, parallel := range []bool{false, true} {
	//
	//					// Execute the test with this parameter combination
	//					strB.WriteString(BaseTest(id, support.TdIdf, parallel, preIndexed, normalize, size, jump))
	//					id++ // Increment test ID
	//
	//					// Execute the test with this parameter combination
	//					strB.WriteString(BaseTest(id, support.Bm25, parallel, preIndexed, normalize, size, jump))
	//					id++ // Increment test ID
	//				}
	//			}
	//		}
	//	}
	//}

	strB.WriteString(BaseTest(id, support.TdIdf, true, false, false, 1, 2))
	strB.WriteString(BaseTest(id, support.Bm25, true, false, false, 3, 2))

	// --- Saving Results ---

	// Print an empty line for better console formatting.
	fmt.Println()
	fmt.Println(strB.String())

	err := os.WriteFile(ResultsOutput, []byte(strB.String()), 0644) // 0644 are file permissions (read/write for owner, read for others).
	if err != nil {
		// If saving fails, fail the test.
		log.Fatalf("error saving results.json: %v", err)
	}
}

// BaseTest executes a full benchmark and validation cycle.
// Records execution time and adds parameters and duration to the results slice.
func BaseTest(testId int64, algo support.Algo, parallel, preIndexed, normalizeJumps bool, size, jumps int) string {

	// Marks the start of the test execution.
	label := fmt.Sprintf("[TEST %02d] algo=%s | preIndexed=%v | normalizedJumps=%v | size=%d | jumps=%d {\n",
		testId, algo, preIndexed, normalizeJumps, size, jumps)

	// Prints the current test configuration to the console.
	fmt.Printf("\n%s\n", label)

	name, db := corpus.CreateDatabaseCaches(testId, false, size, jumps)

	legalInputs := "./../../misc/searchLegalInputs.json"
	res, err := corpus.ApplyLegalInputsDir(db, legalInputs, algo, preIndexed, normalizeJumps, parallel, size, jumps)
	if err != nil {
		log.Fatalf(err.Error())
	}

	label += res.String()
	label += "}\n\n"

	if err = os.Remove(name); err != nil {
		log.Fatalf("error removing corpus file: %v", err)
	}
	return label
}
