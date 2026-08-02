package fuzzy

import "testing"

func TestFuzzyFindAllTargetBoundaries(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog"
	target := "brown fox"
	cands := FuzzyFindAllTargetBoundaries(text, target, 4, 2, 70, 70, 75, 2, true)
	if len(cands) == 0 {
		t.Fatalf("expected candidates")
	}
	if cands[0].MatchedSubstr == "" {
		t.Fatalf("matched substring empty")
	}
	if cands[0].OverallScore < 70 {
		t.Fatalf("low overall score: %d", cands[0].OverallScore)
	}
}
