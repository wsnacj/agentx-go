package browserops

import (
	"fmt"
	"sort"
	"strings"
)

type BrowserStructuredDataFieldExpectation struct {
	Name         string `json:"name,omitempty"`
	Type         string `json:"type,omitempty"`
	Description  string `json:"description,omitempty"`
	Required     bool   `json:"required,omitempty"`
	ExpectedText string `json:"expected_text,omitempty"`
}

type BrowserStructuredDataEvaluationInput struct {
	Bundle              BrowserEvidenceBundle
	SnapshotText        string
	ScreenshotPath      string
	FinalURL            string
	TargetURL           string
	Fields              []BrowserStructuredDataFieldExpectation
	ExtractedData       map[string]any
	URLContains         string
	RequireScreenshot   bool
	RequireFinalURL     bool
	MinSnapshotChars    int
	AllowFailureReasons bool
}

type BrowserStructuredDataFieldResult struct {
	Name              string   `json:"name,omitempty"`
	ExpectedText      string   `json:"expected_text,omitempty"`
	ExtractedValue    string   `json:"extracted_value,omitempty"`
	PresentInSnapshot bool     `json:"present_in_snapshot,omitempty"`
	Passed            bool     `json:"passed"`
	FailureReasons    []string `json:"failure_reasons,omitempty"`
}

type BrowserStructuredDataEvaluation struct {
	Passed              bool                               `json:"passed"`
	Score               float64                            `json:"score"`
	Summary             string                             `json:"summary"`
	EvidenceReady       bool                               `json:"evidence_ready"`
	SnapshotReady       bool                               `json:"snapshot_ready"`
	ScreenshotReady     bool                               `json:"screenshot_ready"`
	FinalURLReady       bool                               `json:"final_url_ready"`
	URLExpectationReady bool                               `json:"url_expectation_ready"`
	FieldsReady         bool                               `json:"fields_ready"`
	FailureReasonsReady bool                               `json:"failure_reasons_ready"`
	ExtractedData       map[string]any                     `json:"extracted_data,omitempty"`
	FieldResults        []BrowserStructuredDataFieldResult `json:"field_results,omitempty"`
	Evidence            []string                           `json:"evidence"`
	FailureReasons      []string                           `json:"failure_reasons"`
	EvidenceBundle      BrowserEvidenceBundle              `json:"evidence_bundle,omitempty"`
}

func EvaluateBrowserStructuredDataEvidence(input BrowserStructuredDataEvaluationInput) BrowserStructuredDataEvaluation {
	bundle := normalizeBrowserStructuredDataEvidenceBundle(input)
	fields := normalizeBrowserStructuredDataFields(input.Fields)
	targetURL := firstNonEmptyBrowserEvidenceString(input.TargetURL, bundle.TargetURL)
	urlContains := strings.TrimSpace(input.URLContains)
	requireFinalURL := input.RequireFinalURL || targetURL != "" || urlContains != ""

	requiredChecks := 0
	passedChecks := 0
	evidence := []string{}
	reasons := []string{}
	out := BrowserStructuredDataEvaluation{
		ScreenshotReady:     !input.RequireScreenshot,
		FinalURLReady:       !requireFinalURL,
		URLExpectationReady: urlContains == "",
		FieldsReady:         len(fields) > 0,
		FailureReasonsReady: true,
		ExtractedData:       cloneStructuredDataMap(input.ExtractedData),
		EvidenceBundle:      bundle,
	}

	requiredChecks++
	out.SnapshotReady = browserPageStateSnapshotReady(bundle.PageSnapshot, input.MinSnapshotChars, &evidence, &reasons)
	if out.SnapshotReady {
		passedChecks++
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
	if urlContains != "" {
		requiredChecks++
		out.URLExpectationReady = browserPageStateURLContainsReady(bundle.FinalURL, urlContains, &evidence, &reasons)
		if out.URLExpectationReady {
			passedChecks++
		}
	}
	if len(fields) == 0 {
		requiredChecks++
		out.FieldsReady = false
		reasons = append(reasons, "fields_missing")
	} else {
		for _, field := range fields {
			requiredChecks++
			result := evaluateBrowserStructuredDataField(bundle.PageSnapshot, input.ExtractedData, field)
			out.FieldResults = append(out.FieldResults, result)
			if result.Passed {
				passedChecks++
				evidence = append(evidence, "field_ready:"+field.Name)
			} else {
				out.FieldsReady = false
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
	out.Passed = out.EvidenceReady && out.URLExpectationReady && out.FieldsReady && len(out.FailureReasons) == 0
	out.Summary = browserStructuredDataEvaluationSummary(out)
	return out
}

func BrowserStructuredDataEvaluationSchema() map[string]any {
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
			"url_expectation_ready": map[string]any{"type": "boolean"},
			"fields_ready":          map[string]any{"type": "boolean"},
			"failure_reasons_ready": map[string]any{"type": "boolean"},
			"extracted_data":        map[string]any{"type": "object"},
			"field_results": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name":                map[string]any{"type": "string"},
						"expected_text":       map[string]any{"type": "string"},
						"extracted_value":     map[string]any{"type": "string"},
						"present_in_snapshot": map[string]any{"type": "boolean"},
						"passed":              map[string]any{"type": "boolean"},
						"failure_reasons":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
					},
				},
			},
			"evidence":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"failure_reasons": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"evidence_bundle": BrowserEvidenceBundleSchema(),
		},
	}
}

func normalizeBrowserStructuredDataEvidenceBundle(input BrowserStructuredDataEvaluationInput) BrowserEvidenceBundle {
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

func evaluateBrowserStructuredDataField(snapshot *BrowserPageSnapshotEvidence, extracted map[string]any, field BrowserStructuredDataFieldExpectation) BrowserStructuredDataFieldResult {
	name := strings.TrimSpace(field.Name)
	reasonName := browserStructuredDataReasonName(name)
	result := BrowserStructuredDataFieldResult{
		Name:         name,
		ExpectedText: strings.TrimSpace(field.ExpectedText),
	}
	if name == "" {
		result.FailureReasons = append(result.FailureReasons, "field_missing_name")
		return result
	}
	if snapshot != nil {
		result.PresentInSnapshot = browserStructuredDataTextContains(snapshot.Text, result.ExpectedText)
	}
	result.ExtractedValue = strings.TrimSpace(browserStructuredDataExtractedValue(extracted, name))
	if field.Required && result.ExtractedValue == "" {
		result.FailureReasons = append(result.FailureReasons, "field_"+reasonName+":extracted_value_missing")
	}
	if result.ExpectedText != "" {
		if !result.PresentInSnapshot {
			result.FailureReasons = append(result.FailureReasons, "field_"+reasonName+":expected_text_missing")
		}
		if result.ExtractedValue != "" && !browserStructuredDataTextContains(result.ExtractedValue, result.ExpectedText) {
			result.FailureReasons = append(result.FailureReasons, "field_"+reasonName+":extracted_value_mismatch")
		}
	}
	result.Passed = len(result.FailureReasons) == 0
	return result
}

func normalizeBrowserStructuredDataFields(values []BrowserStructuredDataFieldExpectation) []BrowserStructuredDataFieldExpectation {
	out := make([]BrowserStructuredDataFieldExpectation, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		value.Name = strings.TrimSpace(value.Name)
		value.Type = strings.TrimSpace(value.Type)
		value.Description = strings.TrimSpace(value.Description)
		value.ExpectedText = strings.TrimSpace(value.ExpectedText)
		if value.Name == "" || seen[value.Name] {
			continue
		}
		seen[value.Name] = true
		out = append(out, value)
	}
	return out
}

func browserStructuredDataExtractedValue(values map[string]any, name string) string {
	if len(values) == 0 || strings.TrimSpace(name) == "" {
		return ""
	}
	if value, ok := values[name]; ok {
		return browserStructuredDataValueString(value)
	}
	lowerName := strings.ToLower(name)
	for key, value := range values {
		if strings.ToLower(strings.TrimSpace(key)) == lowerName {
			return browserStructuredDataValueString(value)
		}
	}
	return ""
}

func browserStructuredDataValueString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case []byte:
		return string(typed)
	case fmt.Stringer:
		return typed.String()
	case nil:
		return ""
	default:
		return fmt.Sprint(typed)
	}
}

func browserStructuredDataTextContains(haystack string, needle string) bool {
	needle = strings.TrimSpace(needle)
	if needle == "" {
		return true
	}
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

func browserStructuredDataReasonName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "unknown"
	}
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	return name
}

func cloneStructuredDataMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]any, len(values))
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		out[key] = values[key]
	}
	return out
}

func browserStructuredDataEvaluationSummary(eval BrowserStructuredDataEvaluation) string {
	if eval.Passed {
		return fmt.Sprintf("structured data extracted: score=%.2f fields=%d evidence=%t",
			eval.Score,
			len(eval.FieldResults),
			eval.EvidenceReady,
		)
	}
	return fmt.Sprintf("structured data extraction failed: score=%.2f failures=%s",
		eval.Score,
		strings.Join(eval.FailureReasons, ","),
	)
}
