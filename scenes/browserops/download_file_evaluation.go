package browserops

import (
	"fmt"
	"path/filepath"
	"strings"
)

type BrowserDownloadFileEvaluationInput struct {
	Bundle              BrowserEvidenceBundle
	TargetURL           string
	DownloadURL         string
	DownloadedPath      string
	ScreenshotPath      string
	FinalURL            string
	Status              string
	ContentType         string
	Bytes               int64
	ExpectedFilename    string
	ExpectedContentType string
	URLContains         string
	MinBytes            int64
	RequireFinalURL     bool
	RequireScreenshot   bool
	AllowFailureReasons bool
}

type BrowserDownloadFileEvaluation struct {
	Passed              bool                  `json:"passed"`
	Score               float64               `json:"score"`
	Summary             string                `json:"summary"`
	EvidenceReady       bool                  `json:"evidence_ready"`
	DownloadedFileReady bool                  `json:"downloaded_file_ready"`
	ScreenshotReady     bool                  `json:"screenshot_ready"`
	FinalURLReady       bool                  `json:"final_url_ready"`
	FilenameReady       bool                  `json:"filename_ready"`
	ContentTypeReady    bool                  `json:"content_type_ready"`
	ByteSizeReady       bool                  `json:"byte_size_ready"`
	URLExpectationReady bool                  `json:"url_expectation_ready"`
	FailureReasonsReady bool                  `json:"failure_reasons_ready"`
	Evidence            []string              `json:"evidence"`
	FailureReasons      []string              `json:"failure_reasons"`
	EvidenceBundle      BrowserEvidenceBundle `json:"evidence_bundle,omitempty"`
}

func EvaluateBrowserDownloadFileEvidence(input BrowserDownloadFileEvaluationInput) BrowserDownloadFileEvaluation {
	bundle := normalizeBrowserDownloadFileEvidenceBundle(input)
	targetURL := firstNonEmptyBrowserEvidenceString(input.TargetURL, bundle.TargetURL, input.DownloadURL)
	urlContains := strings.TrimSpace(input.URLContains)
	expectedFilename := strings.TrimSpace(input.ExpectedFilename)
	expectedContentType := strings.TrimSpace(input.ExpectedContentType)
	requireFinalURL := input.RequireFinalURL || targetURL != "" || urlContains != ""
	minBytes := input.MinBytes
	if minBytes < 0 {
		minBytes = 0
	}

	requiredChecks := 0
	passedChecks := 0
	evidence := []string{}
	reasons := []string{}
	out := BrowserDownloadFileEvaluation{
		ScreenshotReady:     !input.RequireScreenshot,
		FinalURLReady:       !requireFinalURL,
		FilenameReady:       expectedFilename == "",
		ContentTypeReady:    expectedContentType == "",
		ByteSizeReady:       minBytes == 0,
		URLExpectationReady: urlContains == "",
		FailureReasonsReady: true,
		EvidenceBundle:      bundle,
	}

	requiredChecks++
	out.DownloadedFileReady = browserDownloadFileReady(bundle.DownloadedFile, &evidence, &reasons)
	if out.DownloadedFileReady {
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
	if expectedFilename != "" {
		requiredChecks++
		out.FilenameReady = browserDownloadFilenameReady(bundle.DownloadedFile, expectedFilename, &evidence, &reasons)
		if out.FilenameReady {
			passedChecks++
		}
	}
	if expectedContentType != "" {
		requiredChecks++
		out.ContentTypeReady = browserDownloadContentTypeReady(bundle.DownloadedFile, expectedContentType, &evidence, &reasons)
		if out.ContentTypeReady {
			passedChecks++
		}
	}
	if minBytes > 0 {
		requiredChecks++
		out.ByteSizeReady = browserDownloadByteSizeReady(bundle.DownloadedFile, minBytes, &evidence, &reasons)
		if out.ByteSizeReady {
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
	out.EvidenceReady = out.DownloadedFileReady && out.ScreenshotReady && out.FinalURLReady && out.FailureReasonsReady
	out.Passed = out.EvidenceReady && out.FilenameReady && out.ContentTypeReady &&
		out.ByteSizeReady && out.URLExpectationReady && len(out.FailureReasons) == 0
	out.Summary = browserDownloadFileEvaluationSummary(out)
	return out
}

func BrowserDownloadFileEvaluationSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"passed":                map[string]any{"type": "boolean"},
			"score":                 map[string]any{"type": "number"},
			"summary":               map[string]any{"type": "string"},
			"evidence_ready":        map[string]any{"type": "boolean"},
			"downloaded_file_ready": map[string]any{"type": "boolean"},
			"screenshot_ready":      map[string]any{"type": "boolean"},
			"final_url_ready":       map[string]any{"type": "boolean"},
			"filename_ready":        map[string]any{"type": "boolean"},
			"content_type_ready":    map[string]any{"type": "boolean"},
			"byte_size_ready":       map[string]any{"type": "boolean"},
			"url_expectation_ready": map[string]any{"type": "boolean"},
			"failure_reasons_ready": map[string]any{"type": "boolean"},
			"evidence":              map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"failure_reasons":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"evidence_bundle":       BrowserEvidenceBundleSchema(),
		},
	}
}

func normalizeBrowserDownloadFileEvidenceBundle(input BrowserDownloadFileEvaluationInput) BrowserEvidenceBundle {
	bundle := input.Bundle
	if strings.TrimSpace(bundle.TargetURL) == "" {
		bundle.TargetURL = strings.TrimSpace(input.TargetURL)
	}
	finalURL := strings.TrimSpace(input.FinalURL)
	if finalURL == "" && strings.TrimSpace(input.DownloadURL) != "" {
		finalURL = strings.TrimSpace(input.DownloadURL)
	}
	if bundle.FinalURL == nil && finalURL != "" {
		bundle.FinalURL = &BrowserFinalURLEvidence{
			URL:        finalURL,
			TargetURL:  firstNonEmptyBrowserEvidenceString(input.TargetURL, input.DownloadURL, bundle.TargetURL),
			SourceTool: "browser_download_file",
			Ready:      true,
		}
	}
	if bundle.DownloadedFile == nil && strings.TrimSpace(input.DownloadedPath) != "" {
		path := strings.TrimSpace(input.DownloadedPath)
		contentType := strings.TrimSpace(input.ContentType)
		bundle.DownloadedFile = &BrowserDownloadedFileEvidence{
			Path:        path,
			FinalURL:    finalURL,
			Status:      strings.TrimSpace(input.Status),
			ContentType: contentType,
			ByteSize:    input.Bytes,
			SourceTool:  "browser_download_file",
			Artifact: BrowserArtifactRef{
				Type:       BrowserArtifactTypeDownloadedFile,
				Path:       path,
				MIMEType:   contentType,
				Role:       BrowserEvidenceKindDownloadedFile,
				SourceTool: "browser_download_file",
			},
			Ready: true,
		}
		bundle.ArtifactRefs = appendBrowserArtifactRef(bundle.ArtifactRefs, bundle.DownloadedFile.Artifact)
	}
	if bundle.Screenshot == nil && strings.TrimSpace(input.ScreenshotPath) != "" {
		path := strings.TrimSpace(input.ScreenshotPath)
		bundle.Screenshot = &BrowserScreenshotEvidence{
			Path:       path,
			FinalURL:   finalURL,
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
	return bundle
}

func browserDownloadFileReady(file *BrowserDownloadedFileEvidence, evidence *[]string, reasons *[]string) bool {
	if file == nil || strings.TrimSpace(file.Path) == "" {
		*reasons = append(*reasons, "downloaded_file_missing")
		return false
	}
	if status := strings.ToLower(strings.TrimSpace(file.Status)); status != "" && status != "downloaded" && status != "ok" {
		*reasons = append(*reasons, "download_status_not_ready:"+status)
		return false
	}
	*evidence = append(*evidence, "downloaded_file_ready")
	return true
}

func browserDownloadFilenameReady(file *BrowserDownloadedFileEvidence, expected string, evidence *[]string, reasons *[]string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}
	if file == nil {
		*reasons = append(*reasons, "filename_download_missing")
		return false
	}
	filename := strings.TrimSpace(file.SuggestedFilename)
	if filename == "" {
		filename = filepath.Base(strings.TrimSpace(file.Path))
	}
	if !strings.EqualFold(filename, expected) && !strings.Contains(strings.ToLower(filename), strings.ToLower(expected)) {
		*reasons = append(*reasons, "filename_mismatch:"+expected)
		return false
	}
	*evidence = append(*evidence, "filename_ready:"+expected)
	return true
}

func browserDownloadContentTypeReady(file *BrowserDownloadedFileEvidence, expected string, evidence *[]string, reasons *[]string) bool {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return true
	}
	if file == nil || strings.TrimSpace(file.ContentType) == "" {
		*reasons = append(*reasons, "content_type_missing")
		return false
	}
	if !strings.Contains(strings.ToLower(file.ContentType), strings.ToLower(expected)) {
		*reasons = append(*reasons, "content_type_mismatch:"+expected)
		return false
	}
	*evidence = append(*evidence, "content_type_ready:"+expected)
	return true
}

func browserDownloadByteSizeReady(file *BrowserDownloadedFileEvidence, minBytes int64, evidence *[]string, reasons *[]string) bool {
	if minBytes <= 0 {
		return true
	}
	if file == nil || file.ByteSize < minBytes {
		*reasons = append(*reasons, fmt.Sprintf("byte_size_too_small:%d", minBytes))
		return false
	}
	*evidence = append(*evidence, fmt.Sprintf("byte_size_ready:%d", minBytes))
	return true
}

func browserDownloadFileEvaluationSummary(eval BrowserDownloadFileEvaluation) string {
	if eval.Passed {
		return fmt.Sprintf("download file verified: score=%.2f evidence=%t",
			eval.Score,
			eval.EvidenceReady,
		)
	}
	return fmt.Sprintf("download file verification failed: score=%.2f failures=%s",
		eval.Score,
		strings.Join(eval.FailureReasons, ","),
	)
}
