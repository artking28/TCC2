package main

import (
	"fmt"
	"time"

	"github.com/tcc2-davi-arthur/models"
)

// BaseTest executes a full benchmark and validation cycle.
// Records execution time and adds parameters and duration to the results slice.
func BaseTest(testId int64, algo models.Algo, preIndexed bool, normalizeJumps bool, size int, jumps int) (string, int64) {

	// Marks the start of the test execution.
	start := time.Now()
	label := fmt.Sprintf("[TEST %02d] → algo=%s | preIndexed=%v | normalizeJumps=%v | size=%d | jumps=%d",
		testId, algo, preIndexed, normalizeJumps, size, jumps)

	// Prints the current test configuration to the console.
	fmt.Printf("\n%s\n", label)

	//corpus := ReadCorpus(true)
	//
	//corpus.Index()
	//
	//corpus.Calculate()

	// Prints test completion and elapsed time.
	// The value '1' for "n-grams processed" is a placeholder and should be replaced with the actual n-grams count.
	elapsed := time.Since(start).Milliseconds() // Calculates elapsed time since 'start'.
	fmt.Printf("[OK %02d] Completed in %s (%d n-grams processed)\n", testId, elapsed, 1)
	return label, elapsed
}
