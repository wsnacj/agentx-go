package browserops

import (
	"fmt"
	"strings"
)

type BrowserSiteSearchResultExpectation struct {
	Title           string `json:"title,omitempty"`
	URLContains     string `json:"url_contains,omitempty"`
	SnippetContains string `json:"snippet_contains,omitempty"`
}

type BrowserSiteSearchEvaluationInput struct {
	Bundle              BrowserEvidenceBundle
	SnapshotText        string
	ScreenshotPath      string
	FinalURL            string
	TargetURL           string
	Query               string
	ExpectedResults     []BrowserSiteSearchResultExpectation
	RequiredText        []string
	ForbiddenText       []string
	URLContains         string
	RequireQueryVisible bool
	RequireScreenshot   bool
	RequireFinalURL     bool
	RequireSearchAction bool
	RequireSubmitted    bool
	MinSnapshotChars    int
	FieldCount          int
	Submitted           bool
	AllowFailureReasons bool
}

type BrowserSiteSearchResultEvaluation struct {
	Title           string   `json:"title,omitempty"`
	URLContains     string   `json:"url_contains,omitempty"`
	SnippetContains string   `json:"snippet_contains,omitempty"`
	TitleReady      bool     `json:"title_ready,omitempty"`
	URLReady        bool     `json:"url_ready,omitempty"`
	SnippetReady    bool     `json:"snippet_ready,omitempty"`
	Passed          bool     `json:"passed"`
	FailureReasons  []string `json:"failure_reasons,omitempty"`
}

type BrowserSiteSearchEvaluation struct {
	Passed              bool                                `json:"passed"`
	Score               float64                             `json:"score"`
	Summary             string                              `json:"summary"`
	EvidenceReady       bool                                `json:"evidence_ready"`
	SnapshotReady       bool                                `json:"snapshot_ready"`
	ScreenshotReady     bool                                `json:"screenshot_ready"`
	FinalURLReady       bool                                `json:"final_url_ready"`
	QueryReady          bool                                `json:"query_ready"`
	SearchActionReady   bool                                `json:"search_action_ready"`
	RequiredTextReady   bool                                `json:"required_text_ready"`
	ForbiddenTextReady  bool                                `json:"forbidden_text_ready"`
	URLExpectationReady bool                                `json:"url_expectation_ready"`
	ResultsReady        bool                                `json:"results_ready"`
	FailureReasonsReady bool                                `json:"failure_reasons_ready"`
	ResultEvaluations   []BrowserSiteSearchResultEvaluation `json:"result_evaluations,omitempty"`
	Evidence            []string                            `json:"evidence"`
	FailureReasons      []string                            `json:"failure_reasons"`
	EvidenceBundle      BrowserEvidenceBundle               `json:"evidence_bundle,omitempty"`
}

func EvaluateBrowserSiteSearchEvidence(input BrowserSiteSearchEvaluationInput) BrowserSiteSearchEvaluation {
	bundle := normalizeBrowserSiteSearchEvidenceBundle(input)
	expectedResults := normalizeBrowserSiteSearchResults(input.ExpectedResults)
	requiredText := normalizeBrowserPageStateTerms(input.RequiredText)
	forbiddenText := normalizeBrowserPageStateTerms(input.ForbiddenText)
	query := strings.TrimSpace(input.Query)
	targetURL := firstNonEmptyBrowserEvidenceString(input.TargetURL, bundle.TargetURL)
	urlContains := strings.TrimSpace(input.URLContains)
	requireFinalURL := input.RequireFinalURL || targetURL != "" || urlContains != ""

	requiredChecks := 0
	passedChecks := 0
	evidence := []string{}
	reasons := []string{}
	out := BrowserSiteSearchEvaluation{
		ScreenshotReady:     !input.RequireScreenshot,
		FinalURLReady:       !requireFinalURL,
		QueryReady:          query != "",
		SearchActionReady:   !input.RequireSearchAction && !input.RequireSubmitted,
		RequiredTextReady:   len(requiredText) == 0,
		ForbiddenTextReady:  len(forbiddenText) == 0,
		URLExpectationReady: urlContains == "",
		ResultsReady:        len(expectedResults) > 0,
		FailureReasonsReady: true,
		EvidenceBundle:      bundle,
	}

	requiredChecks++
	out.SnapshotReady = browserPageStateSnapshotReady(bundle.PageSnapshot, input.MinSnapshotChars, &evidence, &reasons)
	if out.SnapshotReady {
		passedChecks++
	}
	requiredChecks++
	out.QueryReady = browserSiteSearchQueryReady(bundle.PageSnapshot, query, input.RequireQueryVisible, &evidence, &reasons)
	if out.QueryReady {
		passedChecks++
	}
	if input.RequireSearchAction || input.RequireSubmitted {
		requiredChecks++
		out.SearchActionReady = browserSiteSearchActionReady(input.FieldCount, input.Submitted, input.RequireSubmitted, &evidence, &reasons)
		if out.SearchActionReady {
			passedChecks++
		}
	}
	if input.RequireScreenshot {
		requiredChecks++
		out.ScreenshotReady = browserEvidenceScreenshotReady(bundle.Screenshot, &evidence, &reasons)
		if out.ScreenshotReady {
			passedChecks++
		}
	}
	if requireFinalURL {
		requiredChecks++
		out.FinalURLReady = browserEvidenceFinalURLReady(bundle.FinalURL, targetURL, &evidence, &reasons)
		if out.FinalURLReady {
			passedChecks++
		}
	}
	if len(requiredText) > 0 {
		requiredChecks++
		out.RequiredTextReady = browserPageStateRequiredTextReady(bundle.PageSnapshot, requiredText, &evidence, &reasons)
		if out.RequiredTextReady {
			passedChecks++
		}
	}
	if len(forbiddenText) > 0 {
		requiredChecks++
		out.ForbiddenTextReady = browserPageStateForbiddenTextReady(bundle.PageSnapshot, forbiddenText, &evidence, &reasons)
		if out.ForbiddenTextReady {
			passedChecks++
		}
	}
	if urlContains != "" {
		requiredChecks++
		out.URLExpectationReady = browserPageStateURLContainsReady(bundle.FinalURL, urlContains, &evidence, &reasons)
		if out.URLExpectationReady {
			passedChecks++
		}
	}
	if len(expectedResults) == 0 {
		requiredChecks++
		out.ResultsReady = false
		reasons = append(reasons, "expected_results_missing")
	} else {
		for _, expected := range expectedResults {
			requiredChecks++
			result := evaluateBrowserSiteSearchResult(bundle.PageSnapshot, expected)
			out.ResultEvaluations = append(out.ResultEvaluations, result)
			if result.Passed {
				passedChecks++
				evidence = append(evidence, "result_ready:"+browserSiteSearchResultEvidenceLabel(expected))
			} else {
				out.ResultsReady = false
				reasons = append(reasons, result.FailureReasons...)
			}
		}
	}
	if !input.AllowFailureReasons {
		requiredChecks++
		out.FailureReasonsReady = len(bundle.FailureReasons) == 0
		if out.FailureReasonsReady {
			passedChecks++
		} else {
			reasons = append(reasons, "failure_reasons_present")
			for _, reason := range bundle.FailureReasons {
				if code := strings.TrimSpace(reason.Code); code != "" {
					reasons = append(reasons, "failure_reason:"+code)
				}
			}
		}
	}

	out.Evidence = uniqueBrowserActionFailureReasons(evidence)
	out.FailureReasons = uniqueBrowserActionFailureReasons(reasons)
	out.Score = float64(passedChecks) / float64(requiredChecks)
	out.EvidenceReady = out.SnapshotReady && out.ScreenshotReady && out.FinalURLReady && out.FailureReasonsReady
	out.Passed = out.EvidenceReady && out.QueryReady && out.SearchActionReady &&
		out.RequiredTextReady && out.ForbiddenTextReady && out.URLExpectationReady &&
		out.ResultsReady && len(out.FailureReasons) == 0
	out.Summary = browserSiteSearchEvaluationSummary(out)
	return out
}

func BrowserSiteSearchEvaluationSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"passed":                map[string]any{"type": "boolean"},
			"score":                 map[string]any{"type": "number"},
			"summary":               map[string]any{"type": "string"},
			"evidence_ready":        map[string]any{"type": "boolean"},
			"snapshot_ready":        map[string]any{"type": "boolean"},
			"screenshot_ready":      map[string]any{"type": "boolean"},
			"final_url_ready":       map[string]any{"type": "boolean"},
			"query_ready":           map[string]any{"type": "boolean"},
			"search_action_ready":   map[string]any{"type": "boolean"},
			"required_text_ready":   map[string]any{"type": "boolean"},
			"forbidden_text_ready":  map[string]any{"type": "boolean"},
			"url_expectation_ready": map[string]any{"type": "boolean"},
			"results_ready":         map[string]any{"type": "boolean"},
			"failure_reasons_ready": map[string]any{"type": "boolean"},
			"result_evaluations": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"title":            map[string]any{"type": "string"},
						"url_contains":     map[string]any{"type": "string"},
						"snippet_contains": map[string]any{"type": "string"},
						"title_ready":      map[string]any{"type": "boolean"},
						"url_ready":        map[string]any{"type": "boolean"},
						"snippet_ready":    map[string]any{"type": "boolean"},
						"passed":           map[string]any{"type": "boolean"},
						"failure_reasons":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"evidence":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"failure_reasons": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"evidence_bundle": BrowserEvidenceBundleSchema(),
		},
	}
}

func normalizeBrowserSiteSearchEvidenceBundle(input BrowserSiteSearchEvaluationInput) BrowserEvidenceBundle {
	bundle := input.Bundle
	if strings.TrimSpace(bundle.TargetURL) == "" {
		bundle.TargetURL = strings.TrimSpace(input.TargetURL)
	}
	if bundle.PageSnapshot == nil && strings.TrimSpace(input.SnapshotText) != "" {
		bundle.PageSnapshot = &BrowserPageSnapshotEvidence{
			Text:       strings.TrimSpace(input.SnapshotText),
			Format:     "aria",
			SourceTool: "browser_capture_page_snapshot",
			Ready:      true,
		}
	}
	if bundle.Screenshot == nil && strings.TrimSpace(input.ScreenshotPath) != "" {
		path := strings.TrimSpace(input.ScreenshotPath)
		bundle.Screenshot = &BrowserScreenshotEvidence{
			Path:       path,
			FinalURL:   strings.TrimSpace(input.FinalURL),
			FullPage:   true,
			SourceTool: "browser_capture_submission_evidence",
			Artifact: BrowserArtifactRef{
				Type:       BrowserArtifactTypePageScreenshot,
				Path:       path,
				Role:       BrowserEvidenceKindScreenshot,
				SourceTool: "browser_capture_submission_evidence",
			},
			Ready: true,
		}
		bundle.ArtifactRefs = appendBrowserArtifactRef(bundle.ArtifactRefs, bundle.Screenshot.Artifact)
	}
	if bundle.FinalURL == nil && strings.TrimSpace(input.FinalURL) != "" {
		bundle.FinalURL = &BrowserFinalURLEvidence{
			URL:        strings.TrimSpace(input.FinalURL),
			TargetURL:  firstNonEmptyBrowserEvidenceString(input.TargetURL, bundle.TargetURL),
			SourceTool: "browser_capture_page_snapshot",
			Ready:      true,
		}
	}
	return bundle
}

func browserSiteSearchQueryReady(snapshot *BrowserPageSnapshotEvidence, query string, requireVisible bool, evidence *[]string, reasons *[]string) bool {
	query = strings.TrimSpace(query)
	if query == "" {
		*reasons = append(*reasons, "query_missing")
		return false
	}
	if requireVisible {
		if snapshot == nil || !browserSiteSearchTextContains(snapshot.Text, query) {
			*reasons = append(*reasons, "query_not_visible")
			return false
		}
		*evidence = append(*evidence, "query_visible:"+query)
		return true
	}
	*evidence = append(*evidence, "query_present")
	return true
}

func browserSiteSearchActionReady(fieldCount int, submitted bool, requireSubmitted bool, evidence *[]string, reasons *[]string) bool {
	if fieldCount <= 0 {
		*reasons = append(*reasons, "search_fields_not_filled")
		return false
	}
	if requireSubmitted && !submitted {
		*reasons = append(*reasons, "search_not_submitted")
		return false
	}
	*evidence = append(*evidence, "search_action_ready")
	return true
}

func evaluateBrowserSiteSearchResult(snapshot *BrowserPageSnapshotEvidence, expected BrowserSiteSearchResultExpectation) BrowserSiteSearchResultEvaluation {
	title := strings.TrimSpace(expected.Title)
	urlContains := strings.TrimSpace(expected.URLContains)
	snippetContains := strings.TrimSpace(expected.SnippetContains)
	reasonName := browserSiteSearchReasonName(firstNonEmptyBrowserEvidenceString(title, urlContains, snippetContains))
	result := BrowserSiteSearchResultEvaluation{
		Title:           title,
		URLContains:     urlContains,
		SnippetContains: snippetContains,
		TitleReady:      title == "",
		URLReady:        urlContains == "",
		SnippetReady:    snippetContains == "",
	}
	if title == "" && urlContains == "" && snippetContains == "" {
		result.FailureReasons = append(result.FailureReasons, "result_"+reasonName+":criteria_missing")
		return result
	}
	if snapshot == nil || strings.TrimSpace(snapshot.Text) == "" {
		result.FailureReasons = append(result.FailureReasons, "result_"+reasonName+":snapshot_missing")
		return result
	}
	if title != "" {
		result.TitleReady = browserSiteSearchTextContains(snapshot.Text, title)
		if !result.TitleReady {
			result.FailureReasons = append(result.FailureReasons, "result_"+reasonName+":title_missing")
		}
	}
	if urlContains != "" {
		result.URLReady = browserSiteSearchTextContains(snapshot.Text, urlContains)
		if !result.URLReady {
			result.FailureReasons = append(result.FailureReasons, "result_"+reasonName+":url_missing")
		}
	}
	if snippetContains != "" {
		result.SnippetReady = browserSiteSearchTextContains(snapshot.Text, snippetContains)
		if !result.SnippetReady {
			result.FailureReasons = append(result.FailureReasons, "result_"+reasonName+":snippet_missing")
		}
	}
	result.Passed = len(result.FailureReasons) == 0
	return result
}

func normalizeBrowserSiteSearchResults(values []BrowserSiteSearchResultExpectation) []BrowserSiteSearchResultExpectation {
	out := make([]BrowserSiteSearchResultExpectation, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value.Title = strings.TrimSpace(value.Title)
		value.URLContains = strings.TrimSpace(value.URLContains)
		value.SnippetContains = strings.TrimSpace(value.SnippetContains)
		key := strings.Join([]string{value.Title, value.URLContains, value.SnippetContains}, "\x00")
		if key == "\x00\x00" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func browserSiteSearchTextContains(haystack string, needle string) bool {
	needle = strings.TrimSpace(needle)
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

func browserSiteSearchReasonName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "unknown"
	}
	name = strings.ToLower(name)
	replacer := strings.NewReplacer(" ", "_", ".", "_", "-", "_", "/", "_", ":", "_")
	return replacer.Replace(name)
}

func browserSiteSearchResultEvidenceLabel(expected BrowserSiteSearchResultExpectation) string {
	return browserSiteSearchReasonName(firstNonEmptyBrowserEvidenceString(expected.Title, expected.URLContains, expected.SnippetContains))
}

func browserSiteSearchEvaluationSummary(eval BrowserSiteSearchEvaluation) string {
	if eval.Passed {
		return fmt.Sprintf("site search verified: score=%.2f results=%d evidence=%t",
			eval.Score,
			len(eval.ResultEvaluations),
			eval.EvidenceReady,
		)
	}
	return fmt.Sprintf("site search verification failed: score=%.2f failures=%s",
		eval.Score,
		strings.Join(eval.FailureReasons, ","),
	)
}
