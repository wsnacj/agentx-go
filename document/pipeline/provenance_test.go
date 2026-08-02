package pipeline

import (
	"reflect"
	"testing"
)

func TestPageRefsForEvidenceReturnsAllMatchedEvidenceLines(t *testing.T) {
	pages := []string{
		"Statement page\nProfit attributable to owners",
		"Continuation page\n2025 2024\n224,842 194,073",
	}
	got := pageRefsForEvidence(pages, "Profit attributable to owners\n224,842 194,073")
	if !reflect.DeepEqual(got, []int{1, 2}) {
		t.Fatalf("expected cross-page evidence refs, got %#v", got)
	}
}

func TestPageRefsForEvidenceDeduplicatesAndSortsRefs(t *testing.T) {
	pages := []string{
		"Revenue 100\nProfit 20",
		"Other page",
	}
	got := pageRefsForEvidence(pages, "Revenue 100\nProfit 20\nRevenue 100")
	if !reflect.DeepEqual(got, []int{1}) {
		t.Fatalf("expected deduplicated page refs, got %#v", got)
	}
}
