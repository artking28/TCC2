package main

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	mgu "github.com/artking28/myGoUtils"
	"github.com/tcc2-davi-arthur/models"
)

// --- File Configuration Constants ---
const (
	DatabasePath  = "./data.db"           // Path to the database being used or tested.
	ResultsOutput = "./misc/results.json" // Path to the file where benchmark results will be saved.
)

// --- N-gram Size Constants ---
const (
	Unigram = 1 // Represents a 1-word n-gram.
	Bigram  = 2 // Represents a 2-word n-gram (consecutive words or with jumps).
	Trigram = 3 // Represents a 3-word n-gram (consecutive words or with jumps).
)

// --- Maximum Jump Limit Constants ---
const (
	MaxTrigramJumps = 2 // Maximum allowed jumps between words for a Trigram.
	MaxBigramJumps  = 4 // Maximum allowed jumps between words for a Bigram.
)

// TestParams stores the configuration and results of a single benchmark cycle.
type TestParams struct {
	Algo           models.Algo       `json:"algo"`           // The weighting algorithm used (e.g., TdIdf, Bm25).
	PreIndexed     bool              `json:"preIndexed"`     // Indicates if the test used pre-indexing (in-memory cache) or queried the database directly.
	NormalizeJumps bool              `json:"normalizeJumps"` // Indicates if n-grams with the same terms but different jumps were unified.
	Size           int               `json:"size"`           // N-gram size (1 for Unigram, 2 for Bigram, 3 for Trigram).
	Jumps          int               `json:"jumps"`          // Number of allowed jumps for the n-gram.
	DurationMS     int64             `json:"durationMs"`     // Test duration in milliseconds.
	Results        map[string]string `json:"results"`        // A map to store additional results (currently unused but ready for use).
}

// NewTestParams is a constructor function for TestParams.
// Initializes the TestParams structure with the provided configuration parameters.
func NewTestParams(algo models.Algo, preIndexed bool, normalizeJumps bool, size int, jumps int) TestParams {
	return TestParams{
		Algo:           algo,
		PreIndexed:     preIndexed,
		NormalizeJumps: normalizeJumps,
		Size:           size,
		Jumps:          jumps,
		Results:        make(map[string]string), // Initializes the results map.
	}
}

// TestAll is the main test/benchmark function executed by the Go test runner.
// Iterates over all parameter combinations and saves results to a JSON file.
func TestAll(t *testing.T) {

	var results []TestParams // Slice to collect results from all executed tests.
	var id int64 = 1         // Unique counter to identify each test.

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

					// Set up current test parameters with TF-IDF
					p0 := NewTestParams(models.TdIdf, preIndexed, normalize, size, jump)

					// Execute the test with this parameter combination
					BaseTest(t, id, p0, &results)
					id++ // Increment test ID

					// Set up current test parameters with BM25
					p1 := NewTestParams(models.Bm25, preIndexed, normalize, size, jump)

					// Execute the test with this parameter combination
					BaseTest(t, id, p1, &results)
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
		t.Fatalf("error serializing results: %v", err)
	}

	// Save the JSON content to the ResultsOutput file.
	err = os.WriteFile(ResultsOutput, content, 0644) // 0644 are file permissions (read/write for owner, read for others).
	if err != nil {
		// If saving fails, fail the test.
		t.Fatalf("error saving results.json: %v", err)
	}
}

// BaseTest executes a full benchmark and validation cycle.
// Records execution time and adds parameters and duration to the results slice.
//
// Parameters:
// t: *testing.T - The test object provided by Go.
// testId: int64 - Unique identifier for the test.
// params: TestParams - Parameter configuration for this specific test.
// results: *[]TestParams - Pointer to the slice where results will be stored.
func BaseTest(t *testing.T, testId int64, params TestParams, results *[]TestParams) {
	start := time.Now() // Marks the start of the test execution.

	// Prints the current test configuration to the console.
	fmt.Printf(
		"\n[TEST %02d] → algo=%s | preIndexed=%v | normalizeJumps=%v | size=%d | jumps=%d\n",
		testId, params.Algo, params.PreIndexed, params.NormalizeJumps, params.Size, params.Jumps,
	)

	// NOTE: Actual indexing/search and validation logic would be here,
	// but it was omitted in the provided code. Only the timer is present.

	elapsed := time.Since(start)               // Calculates elapsed time since 'start'.
	params.DurationMS = elapsed.Milliseconds() // Saves duration in milliseconds in the params structure.

	// Prints test completion and elapsed time.
	// The value '1' for "n-grams processed" is a placeholder and should be replaced with the actual n-grams count.
	fmt.Printf("[OK %02d] Completed in %s (%d n-grams processed)\n", testId, elapsed.String(), 1)
	*results = append(*results, params) // Adds the params structure (now with duration) to the results slice.
}
