package browserops

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// BrowserVisualEvidenceEvaluationInput is the pack-owned deterministic gate for
// browser screenshot/snapshot evidence readiness. It does not inspect pixels or
// drive retries; it only checks whether the workflow produced enough evidence
// for a visual/page-state evaluator to make a defensible decision.
type BrowserVisualEvidenceEvaluationInput struct {
	SnapshotText          string
	ScreenshotPath        string
	FinalURL              string
	TargetURL             string
	RequiredSnapshotTerms []string
	RequireSnapshot       bool
	RequireScreenshot     bool
	RequireFinalURL       bool
	MinSnapshotChars      int
}

// BrowserVisualEvidenceEvaluation is the stable output shape for the browser
// submit evidence gate. The result is eval evidence, not runtime policy.
type BrowserVisualEvidenceEvaluation struct {
	Passed              bool     `json:"passed"`
	Score               float64  `json:"score"`
	Summary             string   `json:"summary"`
	VisualEvidenceReady bool     `json:"visual_evidence_ready"`
	SnapshotReady       bool     `json:"snapshot_ready"`
	ScreenshotReady     bool     `json:"screenshot_ready"`
	FinalURLReady       bool     `json:"final_url_ready"`
	Evidence            []string `json:"evidence"`
	FailureReasons      []string `json:"failure_reasons"`
}

func EvaluateBrowserVisualEvidenceGate(input BrowserVisualEvidenceEvaluationInput) BrowserVisualEvidenceEvaluation {
	requireSnapshot, requireScreenshot, requireFinalURL := browserVisualEvidenceRequirements(input)
	requiredChecks := 0
	passedChecks := 0
	evidence := []string{}
	reasons := []string{}

	snapshotReady := true
	if requireSnapshot {
		requiredChecks++
		snapshotReady = browserVisualEvidenceSnapshotReady(input, &evidence, &reasons)
		if snapshotReady {
			passedChecks++
		}
	}

	screenshotReady := true
	if requireScreenshot {
		requiredChecks++
		screenshotReady = browserVisualEvidenceScreenshotReady(input.ScreenshotPath, &evidence, &reasons)
		if screenshotReady {
			passedChecks++
		}
	}

	finalURLReady := true
	if requireFinalURL {
		requiredChecks++
		finalURLReady = browserVisualEvidenceFinalURLReady(input.FinalURL, input.TargetURL, &evidence, &reasons)
		if finalURLReady {
			passedChecks++
		}
	}

	if requiredChecks == 0 {
		requiredChecks = 1
		passedChecks = 1
		evidence = append(evidence, "no_required_visual_evidence")
	}

	reasons = uniqueBrowserActionFailureReasons(reasons)
	evidence = uniqueBrowserActionFailureReasons(evidence)
	score := float64(passedChecks) / float64(requiredChecks)
	out := BrowserVisualEvidenceEvaluation{
		Passed:              len(reasons) == 0,
		Score:               score,
		VisualEvidenceReady: len(reasons) == 0,
		SnapshotReady:       snapshotReady,
		ScreenshotReady:     screenshotReady,
		FinalURLReady:       finalURLReady,
		Evidence:            evidence,
		FailureReasons:      reasons,
	}
	out.Summary = browserVisualEvidenceEvaluationSummary(out)
	return out
}

func browserVisualEvidenceRequirements(input BrowserVisualEvidenceEvaluationInput) (bool, bool, bool) {
	requireSnapshot := input.RequireSnapshot
	requireScreenshot := input.RequireScreenshot
	requireFinalURL := input.RequireFinalURL || strings.TrimSpace(input.TargetURL) != ""
	if !requireSnapshot && !requireScreenshot && !requireFinalURL {
		requireSnapshot = true
		requireScreenshot = true
	}
	return requireSnapshot, requireScreenshot, requireFinalURL
}

func browserVisualEvidenceSnapshotReady(input BrowserVisualEvidenceEvaluationInput, evidence *[]string, reasons *[]string) bool {
	snapshot := strings.TrimSpace(input.SnapshotText)
	if snapshot == "" {
		*reasons = append(*reasons, "snapshot_missing")
		return false
	}
	minChars := input.MinSnapshotChars
	if minChars < 0 {
		minChars = 0
	}
	if minChars > 0 && len(snapshot) < minChars {
		*reasons = append(*reasons, "snapshot_too_short")
		return false
	}
	lowerSnapshot := strings.ToLower(snapshot)
	missingTerm := false
	for _, term := range input.RequiredSnapshotTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		if !strings.Contains(lowerSnapshot, strings.ToLower(term)) {
			*reasons = append(*reasons, "snapshot_term_missing:"+term)
			missingTerm = true
			continue
		}
		*evidence = append(*evidence, "snapshot_term:"+term)
	}
	if missingTerm {
		return false
	}
	*evidence = append(*evidence, "snapshot_ready")
	return true
}

func browserVisualEvidenceScreenshotReady(path string, evidence *[]string, reasons *[]string) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		*reasons = append(*reasons, "screenshot_missing")
		return false
	}
	if strings.HasPrefix(path, "data:image/") || strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		*evidence = append(*evidence, "screenshot_ready")
		return true
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".webp":
		*evidence = append(*evidence, "screenshot_ready")
		return true
	default:
		*reasons = append(*reasons, "screenshot_unsupported_path")
		return false
	}
}

func browserVisualEvidenceFinalURLReady(finalURL string, targetURL string, evidence *[]string, reasons *[]string) bool {
	finalURL = strings.TrimSpace(finalURL)
	targetURL = strings.TrimSpace(targetURL)
	if finalURL == "" {
		*reasons = append(*reasons, "final_url_missing")
		return false
	}
	if targetURL != "" && !browserVisualEvidenceURLMatches(finalURL, targetURL) {
		*reasons = append(*reasons, "final_url_target_mismatch")
		return false
	}
	*evidence = append(*evidence, "final_url_ready")
	return true
}

func browserVisualEvidenceURLMatches(finalURL string, targetURL string) bool {
	if strings.TrimSpace(targetURL) == "" {
		return true
	}
	if strings.TrimSpace(finalURL) == strings.TrimSpace(targetURL) || strings.HasPrefix(strings.TrimSpace(finalURL), strings.TrimSpace(targetURL)) {
		return true
	}
	finalParsed, finalErr := url.Parse(finalURL)
	targetParsed, targetErr := url.Parse(targetURL)
	if finalErr != nil || targetErr != nil {
		return false
	}
	if strings.EqualFold(finalParsed.Host, targetParsed.Host) && finalParsed.Host != "" {
		return true
	}
	return false
}

func browserVisualEvidenceEvaluationSummary(eval BrowserVisualEvidenceEvaluation) string {
	if eval.Passed {
		return fmt.Sprintf("visual evidence ready: score=%.2f snapshot=%t screenshot=%t final_url=%t",
			eval.Score,
			eval.SnapshotReady,
			eval.ScreenshotReady,
			eval.FinalURLReady,
		)
	}
	return fmt.Sprintf("visual evidence incomplete: score=%.2f failures=%s",
		eval.Score,
		strings.Join(eval.FailureReasons, ","),
	)
}
