// Package rapidfuzz owns the pure-Go similarity primitive used by OCR diff.
package rapidfuzz

import (
	"math"

	"github.com/agext/levenshtein"
)

// Ratio returns a rounded similarity percentage in the inclusive range 0..100.
func Ratio(left, right string) int {
	return int(math.Round(levenshtein.Similarity(left, right, nil) * 100))
}
