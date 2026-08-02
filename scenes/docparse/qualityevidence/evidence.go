// Package qualityevidence evaluates privacy-safe document extraction fixtures.
package qualityevidence

import (
	"strings"

	"github.com/wsnacj/agentx-go/scenes/docparse/representation"
)

const SchemaVersion = "agentx.docparse.quality_evidence.v1"

const (
	StatusSuccess = "success"
	StatusBounded = "bounded"
	StatusSkipped = "skipped"
)

// PageExpectation defines page-local claims that extraction must preserve.
type PageExpectation struct {
	Page            int      `json:"page"`
	RequiredPhrases []string `json:"required_phrases"`
}

// Spec identifies a committed fixture without exposing a machine-local path.
type Spec struct {
	FixtureID    string            `json:"fixture_id"`
	FixtureClass string            `json:"fixture_class"`
	PrivacySafe  bool              `json:"privacy_safe"`
	Pages        []PageExpectation `json:"pages"`
}

// Report is a repeatable extraction-quality readback.
type Report struct {
	SchemaVersion        string  `json:"schema_version"`
	FixtureID            string  `json:"fixture_id"`
	FixtureClass         string  `json:"fixture_class"`
	PrivacySafe          bool    `json:"privacy_safe"`
	Status               string  `json:"status"`
	Passed               bool    `json:"passed"`
	Reason               string  `json:"reason,omitempty"`
	TextSource           string  `json:"text_source,omitempty"`
	OCRStatus            string  `json:"ocr_status"`
	ExpectedPages        int     `json:"expected_pages"`
	ObservedPages        int     `json:"observed_pages"`
	MatchedPages         int     `json:"matched_pages"`
	PageCoverage         float64 `json:"page_coverage"`
	RequiredPhrases      int     `json:"required_phrases"`
	MatchedPhrases       int     `json:"matched_phrases"`
	RequiredPhraseRecall float64 `json:"required_phrase_recall"`
	Truncated            bool    `json:"truncated"`
}

// Evaluate measures page coverage and required-phrase recall for one
// production document representation.
func Evaluate(spec Spec, doc representation.Document) Report {
	report := Report{
		SchemaVersion: SchemaVersion,
		FixtureID:     strings.TrimSpace(spec.FixtureID),
		FixtureClass:  strings.TrimSpace(spec.FixtureClass),
		PrivacySafe:   spec.PrivacySafe,
		TextSource:    strings.TrimSpace(doc.TextSource),
		OCRStatus:     ocrStatus(doc.TextSource),
		ExpectedPages: len(spec.Pages),
		ObservedPages: len(doc.Pages),
	}
	if len(doc.Pages) == 0 {
		report.Status = StatusBounded
		report.Reason = "document_representation_has_no_pages"
		return report
	}

	pages := make(map[int]string, len(doc.Pages))
	maxObservedPage := 0
	for _, page := range doc.Pages {
		pages[page.Number] = normalize(page.Text)
		if page.Number > maxObservedPage {
			maxObservedPage = page.Number
		}
	}
	maxExpectedPage := 0
	for _, expected := range spec.Pages {
		if expected.Page > maxExpectedPage {
			maxExpectedPage = expected.Page
		}
		pageText, ok := pages[expected.Page]
		if ok {
			report.MatchedPages++
		}
		for _, phrase := range expected.RequiredPhrases {
			phrase = normalize(phrase)
			if phrase == "" {
				continue
			}
			report.RequiredPhrases++
			if ok && strings.Contains(pageText, phrase) {
				report.MatchedPhrases++
			}
		}
	}
	if report.ExpectedPages > 0 {
		report.PageCoverage = float64(report.MatchedPages) / float64(report.ExpectedPages)
	} else {
		report.PageCoverage = 1
	}
	if report.RequiredPhrases > 0 {
		report.RequiredPhraseRecall = float64(report.MatchedPhrases) / float64(report.RequiredPhrases)
	} else {
		report.RequiredPhraseRecall = 1
	}
	report.Truncated = maxExpectedPage > maxObservedPage
	report.Passed = spec.PrivacySafe && report.PageCoverage == 1 && report.RequiredPhraseRecall == 1 && !report.Truncated
	if report.Passed {
		report.Status = StatusSuccess
		return report
	}
	report.Status = StatusBounded
	switch {
	case !spec.PrivacySafe:
		report.Reason = "fixture_not_marked_privacy_safe"
	case report.Truncated:
		report.Reason = "expected_page_range_truncated"
	case report.PageCoverage < 1:
		report.Reason = "expected_page_missing"
	default:
		report.Reason = "required_phrase_recall_below_one"
	}
	return report
}

// Skipped creates an explicit readback for an opt-in OCR/provider lane that
// was not executed. A skip never claims quality success.
func Skipped(spec Spec, reason string) Report {
	return Report{
		SchemaVersion: SchemaVersion,
		FixtureID:     strings.TrimSpace(spec.FixtureID),
		FixtureClass:  strings.TrimSpace(spec.FixtureClass),
		PrivacySafe:   spec.PrivacySafe,
		Status:        StatusSkipped,
		Reason:        strings.TrimSpace(reason),
		OCRStatus:     StatusSkipped,
		ExpectedPages: len(spec.Pages),
	}
}

func ocrStatus(source string) string {
	switch strings.TrimSpace(source) {
	case representation.TextSourceOCRX, representation.TextSourceOCRXTable:
		return StatusSuccess
	case representation.TextSourcePDFTextLayer:
		return "not_required_text_layer_satisfied"
	default:
		return "not_exercised"
	}
}

func normalize(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(value)), " "))
}
