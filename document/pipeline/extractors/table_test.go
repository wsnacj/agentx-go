package extractors

import "testing"

func TestRunTableCandidatesSelectsCurrentPeriodColumn(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `(RMB in millions)
Item 2025 2024
Revenue 751,766 660,257
Profit attributable to equity holders 224,842 194,073`,
		FieldKey:     "Revenue",
		RowLabels:    []string{"Revenue"},
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one table candidate, got %#v", got)
	}
	if got[0].Value != "751,766" || got[0].RowLabel != "Revenue" || got[0].ColumnLabel != "2025" {
		t.Fatalf("unexpected table candidate: %#v", got[0])
	}
	if got[0].Unit != "(RMB in millions)" || got[0].UnitSource != "(RMB in millions)" {
		t.Fatalf("expected unit provenance, got %#v", got[0])
	}
}

func TestRunTableCandidatesAvoidsSuffixOnlyRowMatches(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `Item 2025 2024
Cost of Revenue 123 100
Revenue 999 888`,
		FieldKey:     "Revenue",
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 || got[0].Value != "999" || got[0].RowLabel != "Revenue" {
		t.Fatalf("expected revenue row only, got %#v", got)
	}
}

func TestRunTableCandidatesSupportsPipeRowsAndNegativeValues(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `| Item | 2025 | 2024 |
| Net Profit | (100) | - |`,
		FieldKey:     "NetProfit",
		RowLabels:    []string{"Net Profit"},
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one table candidate, got %#v", got)
	}
	if got[0].Value != "(100)" || got[0].ColumnLabel != "2025" {
		t.Fatalf("unexpected pipe table candidate: %#v", got[0])
	}
}

func TestRunTableCandidatesSkipsNullishSelectedValues(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `Item 2025 2024
Revenue - 100
Net Profit null 90`,
		FieldKey:     "Revenue",
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 0 {
		t.Fatalf("expected nullish current value to be skipped, got %#v", got)
	}
}

func TestRunTableCandidatesSupportsWrappedRowLabels(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `Item 2025 2024
Profit attributable to
owners of the Company 224,842 194,073`,
		FieldKey:     "Profit Attributable To Owners Of The Company",
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one wrapped-label candidate, got %#v", got)
	}
	if got[0].Value != "224,842" || got[0].RowLabel != "Profit attributable to owners of the Company" {
		t.Fatalf("unexpected wrapped-label candidate: %#v", got[0])
	}
}

func TestRunTableCandidatesSelectsCurrentPeriodFromFiveYearWrappedOwnerLoss(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `(RMB in thousands)
Year ended December 31,
2021 2022 2023 2024 2025
(Loss)/profit for the year attributable to
equity holders of the Company (23,538,379) (6,686,110) 13,855,828 35,807,179 (23,355,015)`,
		FieldKey: "Profit Attributable To Owners Of The Company",
		RowLabels: []string{
			"(Loss)/profit for the year attributable to equity holders of the Company",
			"Equity holders of the Company",
		},
		ColumnLabels: []string{"2025"},
	})
	found := false
	for _, candidate := range got {
		if candidate.Value == "(23,355,015)" &&
			candidate.ColumnLabel == "2025" &&
			candidate.Unit == "(RMB in thousands)" &&
			candidate.RowLabel == "(Loss)/profit for the year attributable to equity holders of the Company" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected five-year wrapped owner-loss candidate, got %#v", got)
	}
}

func TestRunTableCandidatesStitchesLabelAndValuesAcrossPageBoundary(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `Item 2025 2024
Operating cash flow
---PAGE---
751,766 660,257`,
		FieldKey:     "Operating cash flow",
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one cross-page table candidate, got %#v", got)
	}
	if got[0].Value != "751,766" || got[0].RowLabel != "Operating cash flow" || got[0].ColumnLabel != "2025" {
		t.Fatalf("unexpected cross-page table candidate: %#v", got[0])
	}
}

func TestRunTableCandidatesStitchesSplitLabelAcrossPageBoundary(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `Item 2025 2024
Profit attributable to
Page 12
---PAGE---
owners of the Company 224,842 194,073`,
		FieldKey:     "Profit Attributable To Owners Of The Company",
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one split-label cross-page candidate, got %#v", got)
	}
	if got[0].Value != "224,842" || got[0].RowLabel != "Profit attributable to owners of the Company" || got[0].ColumnLabel != "2025" {
		t.Fatalf("unexpected split-label cross-page candidate: %#v", got[0])
	}
}

func TestRunTableCandidatesDoesNotStitchChapterTitleAcrossSamePageHeader(t *testing.T) {
	got := RunTableCandidates(TableInput{
		Text: `Summary
Item 2025 2024
Total Revenue 100 90`,
		FieldKey:     "Revenue",
		RowLabels:    []string{"Total Revenue"},
		ColumnLabels: []string{"2025"},
	})
	if len(got) != 1 {
		t.Fatalf("expected one same-page table candidate, got %#v", got)
	}
	if got[0].RowLabel != "Total Revenue" || got[0].Value != "100" {
		t.Fatalf("chapter title should not be stitched into row label: %#v", got[0])
	}
}
