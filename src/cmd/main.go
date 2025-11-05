package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
)

// ResultsOutput represents path to the file where benchmark results will be saved.
const ResultsOutput = "./misc/results.json"

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

	var id int64 = 1 // Unique counter to identify each test.
	results := map[string]int64{}

	// Defines n-gram sizes to be tested and their respective maximum jump limits.
	sizes := []mgu.Pair[int, int]{
		mgu.NewPair(Unigram, 0),               // Unigram: 0 jumps (always).
		mgu.NewPair(Bigram, MaxBigramJumps),   // Bigram: up to 4 jumps.
		mgu.NewPair(Trigram, MaxTrigramJumps), // Trigram: up to 2 jumps.
	}

	// Main loop: iterates through each configured n-gram type (Unigram, Bigram, Trigram).
	for _, s := range sizes {
		size, maxJumps := s.Left, s.Right // Unpacks n-gram size and jump limit.

		// For each n-gram size, test all possible jump levels (from 0 to maxJumps).
		for jump := 0; jump <= maxJumps; jump++ {

			// Test both variations: with and without pre-indexing (in-memory cache vs database query).
			for _, preIndexed := range []bool{false, true} {

				// For each pre-index mode, test jump normalization on and off.
				for _, normalize := range []bool{false, true} {

					// Execute the test with this parameter combination
					resultL, resultT := BaseTest(id, models.TdIdf, preIndexed, normalize, size, jump)
					results[resultL] = resultT
					id++ // Increment test ID

					// Execute the test with this parameter combination
					resultL, resultT = BaseTest(id, models.Bm25, preIndexed, normalize, size, jump)
					results[resultL] = resultT
					id++ // Increment test ID
				}
			}
		}
	}

	// --- Saving Results ---

	// Print an empty line for better console formatting.
	fmt.Println()

	// Serialize the 'results' slice to indented JSON (human-readable).
	content, err := json.MarshalIndent(results, "", "   ")
	if err != nil {
		// If serialization fails, fail the test.
		log.Fatalf("error serializing results: %v", err)
	}

	// Save the JSON content to the ResultsOutput file.
	err = os.WriteFile(ResultsOutput, content, 0644) // 0644 are file permissions (read/write for owner, read for others).
	if err != nil {
		// If saving fails, fail the test.
		log.Fatalf("error saving results.json: %v", err)
	}
}
