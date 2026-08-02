package browserops

import (
	"fmt"
	"strings"
)

type BrowserPageStateEvaluationInput struct {
	Bundle              BrowserEvidenceBundle
	SnapshotText        string
	ScreenshotPath      string
	FinalURL            string
	TargetURL           string
	RequiredText        []string
	ForbiddenText       []string
	URLContains         string
	RequireScreenshot   bool
	RequireFinalURL     bool
	MinSnapshotChars    int
	AllowFailureReasons bool
}

type BrowserPageStateEvaluation struct {
	Passed              bool                  `json:"passed"`
	Score               float64               `json:"score"`
	Summary             string                `json:"summary"`
	EvidenceReady       bool                  `json:"evidence_ready"`
	SnapshotReady       bool                  `json:"snapshot_ready"`
	ScreenshotReady     bool                  `json:"screenshot_ready"`
	FinalURLReady       bool                  `json:"final_url_ready"`
	RequiredTextReady   bool                  `json:"required_text_ready"`
	ForbiddenTextReady  bool                  `json:"forbidden_text_ready"`
	URLExpectationReady bool                  `json:"url_expectation_ready"`
	FailureReasonsReady bool                  `json:"failure_reasons_ready"`
	Evidence            []string              `json:"evidence"`
	FailureReasons      []string              `json:"failure_reasons"`
	EvidenceBundle      BrowserEvidenceBundle `json:"evidence_bundle,omitempty"`
}

func EvaluateBrowserPageStateEvidence(input BrowserPageStateEvaluationInput) BrowserPageStateEvaluation {
	bundle := normalizeBrowserPageStateEvidenceBundle(input)
	requiredText := normalizeBrowserPageStateTerms(input.RequiredText)
	forbiddenText := normalizeBrowserPageStateTerms(input.ForbiddenText)
	urlContains := strings.TrimSpace(input.URLContains)
	targetURL := firstNonEmptyBrowserEvidenceString(input.TargetURL, bundle.TargetURL)
	requireFinalURL := input.RequireFinalURL || targetURL != "" || urlContains != ""

	requiredChecks := 0
	passedChecks := 0
	evidence := []string{}
	reasons := []string{}
	out := BrowserPageStateEvaluation{
		ScreenshotReady:     !input.RequireScreenshot,
		FinalURLReady:       !requireFinalURL,
		RequiredTextReady:   len(requiredText) == 0,
		ForbiddenTextReady:  len(forbiddenText) == 0,
		URLExpectationReady: urlContains == "",
		FailureReasonsReady: true,
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
	out.Passed = out.EvidenceReady && out.RequiredTextReady && out.ForbiddenTextReady && out.URLExpectationReady && len(out.FailureReasons) == 0
	out.Summary = browserPageStateEvaluationSummary(out)
	return out
}

func BrowserPageStateEvaluationSchema() map[string]any {
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
			"required_text_ready":   map[string]any{"type": "boolean"},
			"forbidden_text_ready":  map[string]any{"type": "boolean"},
			"url_expectation_ready": map[string]any{"type": "boolean"},
			"failure_reasons_ready": map[string]any{"type": "boolean"},
			"evidence":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"failure_reasons":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"evidence_bundle":       BrowserEvidenceBundleSchema(),
		},
	}
}

func normalizeBrowserPageStateEvidenceBundle(input BrowserPageStateEvaluationInput) BrowserEvidenceBundle {
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

func browserPageStateSnapshotReady(snapshot *BrowserPageSnapshotEvidence, minChars int, evidence *[]string, reasons *[]string) bool {
	if snapshot == nil || strings.TrimSpace(snapshot.Text) == "" {
		*reasons = append(*reasons, "snapshot_missing")
		return false
	}
	text := strings.TrimSpace(snapshot.Text)
	if minChars < 0 {
		minChars = 0
	}
	if minChars > 0 && len(text) < minChars {
		*reasons = append(*reasons, "snapshot_too_short")
		return false
	}
	*evidence = append(*evidence, "snapshot_ready")
	return true
}

func browserPageStateRequiredTextReady(snapshot *BrowserPageSnapshotEvidence, terms []string, evidence *[]string, reasons *[]string) bool {
	if snapshot == nil || strings.TrimSpace(snapshot.Text) == "" {
		*reasons = append(*reasons, "required_text_snapshot_missing")
		return false
	}
	lowerSnapshot := strings.ToLower(snapshot.Text)
	ready := true
	for _, term := range terms {
		if !strings.Contains(lowerSnapshot, strings.ToLower(term)) {
			*reasons = append(*reasons, "required_text_missing:"+term)
			ready = false
			continue
		}
		*evidence = append(*evidence, "required_text:"+term)
	}
	return ready
}

func browserPageStateForbiddenTextReady(snapshot *BrowserPageSnapshotEvidence, terms []string, evidence *[]string, reasons *[]string) bool {
	if snapshot == nil || strings.TrimSpace(snapshot.Text) == "" {
		*reasons = append(*reasons, "forbidden_text_snapshot_missing")
		return false
	}
	lowerSnapshot := strings.ToLower(snapshot.Text)
	ready := true
	for _, term := range terms {
		if strings.Contains(lowerSnapshot, strings.ToLower(term)) {
			*reasons = append(*reasons, "forbidden_text_present:"+term)
			ready = false
			continue
		}
		*evidence = append(*evidence, "forbidden_text_absent:"+term)
	}
	return ready
}

func browserPageStateURLContainsReady(finalURL *BrowserFinalURLEvidence, contains string, evidence *[]string, reasons *[]string) bool {
	contains = strings.TrimSpace(contains)
	if contains == "" {
		return true
	}
	if finalURL == nil || strings.TrimSpace(finalURL.URL) == "" {
		*reasons = append(*reasons, "final_url_missing")
		return false
	}
	if !strings.Contains(strings.ToLower(finalURL.URL), strings.ToLower(contains)) {
		*reasons = append(*reasons, "final_url_contains_missing:"+contains)
		return false
	}
	*evidence = append(*evidence, "final_url_contains:"+contains)
	return true
}

func normalizeBrowserPageStateTerms(values []string) []string {
	out := []string{}
	seen := map[string]bool{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func browserPageStateEvaluationSummary(eval BrowserPageStateEvaluation) string {
	if eval.Passed {
		return fmt.Sprintf("page state verified: score=%.2f evidence=%t required_text=%t forbidden_text=%t url=%t",
			eval.Score,
			eval.EvidenceReady,
			eval.RequiredTextReady,
			eval.ForbiddenTextReady,
			eval.URLExpectationReady,
		)
	}
	return fmt.Sprintf("page state verification failed: score=%.2f failures=%s",
		eval.Score,
		strings.Join(eval.FailureReasons, ","),
	)
}
