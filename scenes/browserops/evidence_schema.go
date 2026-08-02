package browserops

import (
	"strings"
)

const (
	BrowserArtifactTypePageSnapshot   = "browser.page.snapshot"
	BrowserArtifactTypePageScreenshot = "browser.page.screenshot"
	BrowserArtifactTypeActionTrace    = "browser.action.trace"
	BrowserArtifactTypeSubmitReport   = "browser.submit.report"
	BrowserArtifactTypeDownloadedFile = "browser.download.file"
	BrowserArtifactTypeTaskPlan       = "browser.task.plan"

	BrowserEvidenceKindFinalURL       = "final_url"
	BrowserEvidenceKindPageSnapshot   = "page_snapshot"
	BrowserEvidenceKindScreenshot     = "screenshot"
	BrowserEvidenceKindActionTrace    = "action_trace"
	BrowserEvidenceKindDownloadedFile = "downloaded_file"
	BrowserEvidenceKindFailureReason  = "failure_reason"
	BrowserEvidenceKindTaskPlan       = "task_plan"
)

// BrowserEvidenceBundle is the pack-owned evidence contract for browser
// operation workflows. It intentionally stays in the Browser Ops Domain Kit so
// generic AgentX runtime does not learn browser-task business semantics.
type BrowserEvidenceBundle struct {
	TargetURL      string                         `json:"target_url,omitempty"`
	FinalURL       *BrowserFinalURLEvidence       `json:"final_url,omitempty"`
	PageSnapshot   *BrowserPageSnapshotEvidence   `json:"page_snapshot,omitempty"`
	Screenshot     *BrowserScreenshotEvidence     `json:"screenshot,omitempty"`
	ActionTrace    *BrowserActionTraceEvidence    `json:"action_trace,omitempty"`
	DownloadedFile *BrowserDownloadedFileEvidence `json:"downloaded_file,omitempty"`
	TaskPlan       *BrowserTaskPlanEvidence       `json:"task_plan,omitempty"`
	FailureReasons []BrowserFailureReasonEvidence `json:"failure_reasons,omitempty"`
	ArtifactRefs   []BrowserArtifactRef           `json:"artifact_refs,omitempty"`
}

type BrowserFinalURLEvidence struct {
	URL        string `json:"url,omitempty"`
	TargetURL  string `json:"target_url,omitempty"`
	SourceTool string `json:"source_tool,omitempty"`
	Ready      bool   `json:"ready,omitempty"`
}

type BrowserPageSnapshotEvidence struct {
	Text         string             `json:"text,omitempty"`
	Format       string             `json:"format,omitempty"`
	ElementCount int                `json:"element_count,omitempty"`
	SourceTool   string             `json:"source_tool,omitempty"`
	Artifact     BrowserArtifactRef `json:"artifact,omitempty"`
	Ready        bool               `json:"ready,omitempty"`
}

type BrowserScreenshotEvidence struct {
	Path       string             `json:"path,omitempty"`
	FinalURL   string             `json:"final_url,omitempty"`
	FullPage   bool               `json:"full_page,omitempty"`
	SourceTool string             `json:"source_tool,omitempty"`
	Artifact   BrowserArtifactRef `json:"artifact,omitempty"`
	Ready      bool               `json:"ready,omitempty"`
}

type BrowserActionTraceEvidence struct {
	TraceID      string             `json:"trace_id,omitempty"`
	Path         string             `json:"path,omitempty"`
	Status       string             `json:"status,omitempty"`
	ActionCount  int                `json:"action_count,omitempty"`
	FailureCount int                `json:"failure_count,omitempty"`
	SourceTool   string             `json:"source_tool,omitempty"`
	Artifact     BrowserArtifactRef `json:"artifact,omitempty"`
	Ready        bool               `json:"ready,omitempty"`
}

type BrowserDownloadedFileEvidence struct {
	Path              string             `json:"path,omitempty"`
	FinalURL          string             `json:"final_url,omitempty"`
	Status            string             `json:"status,omitempty"`
	ContentType       string             `json:"content_type,omitempty"`
	ByteSize          int64              `json:"byte_size,omitempty"`
	SuggestedFilename string             `json:"suggested_filename,omitempty"`
	Mode              string             `json:"mode,omitempty"`
	SourceTool        string             `json:"source_tool,omitempty"`
	Artifact          BrowserArtifactRef `json:"artifact,omitempty"`
	Ready             bool               `json:"ready,omitempty"`
}

type BrowserFailureReasonEvidence struct {
	Code         string `json:"code,omitempty"`
	Message      string `json:"message,omitempty"`
	Action       string `json:"action,omitempty"`
	NodeID       string `json:"node_id,omitempty"`
	Tool         string `json:"tool,omitempty"`
	Recoverable  bool   `json:"recoverable,omitempty"`
	EvidenceKind string `json:"evidence_kind,omitempty"`
}

type BrowserTaskPlanEvidence struct {
	TaskID     string                              `json:"task_id,omitempty"`
	Goal       string                              `json:"goal,omitempty"`
	CaseType   string                              `json:"case_type,omitempty"`
	WorkflowID string                              `json:"workflow_id,omitempty"`
	Status     string                              `json:"status,omitempty"`
	SourceTool string                              `json:"source_tool,omitempty"`
	Subtasks   []BrowserSubtaskObservationEvidence `json:"subtasks,omitempty"`
	Artifact   BrowserArtifactRef                  `json:"artifact,omitempty"`
	Ready      bool                                `json:"ready,omitempty"`
}

type BrowserSubtaskObservationEvidence struct {
	ID           string               `json:"id,omitempty"`
	Kind         string               `json:"kind,omitempty"`
	Description  string               `json:"description,omitempty"`
	Status       string               `json:"status,omitempty"`
	EvidenceKind string               `json:"evidence_kind,omitempty"`
	Required     bool                 `json:"required,omitempty"`
	EvidenceRefs []BrowserArtifactRef `json:"evidence_refs,omitempty"`
}

type BrowserArtifactRef struct {
	ArtifactID string `json:"artifact_id,omitempty"`
	Type       string `json:"type,omitempty"`
	Path       string `json:"path,omitempty"`
	StorageRef string `json:"storage_ref,omitempty"`
	URL        string `json:"url,omitempty"`
	Digest     string `json:"digest,omitempty"`
	MIMEType   string `json:"mime_type,omitempty"`
	Role       string `json:"role,omitempty"`
	SourceTool string `json:"source_tool,omitempty"`
}

type BrowserEvidenceRequirements struct {
	TargetURL             string
	RequiredSnapshotTerms []string
	MinSnapshotChars      int
	RequireSnapshot       bool
	RequireScreenshot     bool
	RequireFinalURL       bool
	RequireActionTrace    bool
	RequireDownloadedFile bool
	RequireArtifactRefs   bool
	RequireTaskPlan       bool
	AllowFailureReasons   bool
	MinSubtasks           int
	RequiredSubtaskKinds  []string
}

type BrowserEvidenceReadiness struct {
	Passed              bool     `json:"passed"`
	Score               float64  `json:"score"`
	EvidenceReady       bool     `json:"evidence_ready"`
	SnapshotReady       bool     `json:"snapshot_ready"`
	ScreenshotReady     bool     `json:"screenshot_ready"`
	FinalURLReady       bool     `json:"final_url_ready"`
	ActionTraceReady    bool     `json:"action_trace_ready"`
	DownloadedFileReady bool     `json:"downloaded_file_ready"`
	ArtifactRefsReady   bool     `json:"artifact_refs_ready"`
	TaskPlanReady       bool     `json:"task_plan_ready"`
	FailureReasonsReady bool     `json:"failure_reasons_ready"`
	SubtaskCount        int      `json:"subtask_count,omitempty"`
	SubtaskKinds        []string `json:"subtask_kinds,omitempty"`
	Evidence            []string `json:"evidence"`
	FailureReasons      []string `json:"failure_reasons"`
}

func BuildBrowserEvidenceBundleFromState(state map[string]any) BrowserEvidenceBundle {
	targetURL := browserEvidenceStateString(state, "target_url", "case.input.target_url")
	finalURL := firstNonEmptyBrowserEvidenceString(
		browserEvidenceStateString(state, "review.final_url"),
		browserEvidenceStateString(state, "review.evidence_final_url"),
	)
	snapshot := browserEvidenceStateString(state, "review.snapshot")
	snapshotPath := browserEvidenceStateString(state, "review.snapshot_path")
	screenshotPath := browserEvidenceStateString(state, "review.evidence_path", "review.screenshot_path")
	tracePath := browserEvidenceStateString(state, "review.trace_path", "review.action_trace_path")
	downloadPath := browserEvidenceStateString(state, "download.path", "download.file_path", "download.output_path")
	downloadFinalURL := firstNonEmptyBrowserEvidenceString(browserEvidenceStateString(state, "download.final_url"), finalURL)
	downloadContentType := browserEvidenceStateString(state, "download.content_type", "download.mime_type")
	downloadStatus := browserEvidenceStateString(state, "download.status")
	downloadFilename := browserEvidenceStateString(state, "download.suggested_filename", "download.filename")
	downloadMode := browserEvidenceStateString(state, "download.mode", "download.kind")
	taskPlan := browserTaskPlanEvidenceFromState(firstNonNilBrowserEvidenceStateAny(
		state,
		"review.task_plan",
		"browser.task_plan",
		"task_plan",
	))

	bundle := BrowserEvidenceBundle{
		TargetURL:      targetURL,
		TaskPlan:       taskPlan,
		FailureReasons: browserFailureReasonEvidenceFromState(browserEvidenceStateAny(state, "review.failure_reasons")),
		ArtifactRefs:   browserArtifactRefsFromState(browserEvidenceStateAny(state, "review.artifact_refs")),
	}
	if finalURL != "" {
		bundle.FinalURL = &BrowserFinalURLEvidence{
			URL:        finalURL,
			TargetURL:  targetURL,
			SourceTool: "browser_capture_page_snapshot",
			Ready:      true,
		}
	}
	if snapshot != "" {
		bundle.PageSnapshot = &BrowserPageSnapshotEvidence{
			Text:       snapshot,
			Format:     firstNonEmptyBrowserEvidenceString(browserEvidenceStateString(state, "review.snapshot_format"), "aria"),
			SourceTool: "browser_capture_page_snapshot",
			Artifact: BrowserArtifactRef{
				ArtifactID: browserEvidenceStateString(state, "review.snapshot_artifact_id"),
				Type:       BrowserArtifactTypePageSnapshot,
				Path:       snapshotPath,
				StorageRef: browserEvidenceStateString(state, "review.snapshot_storage_ref"),
				Role:       BrowserEvidenceKindPageSnapshot,
				SourceTool: "browser_capture_page_snapshot",
			},
			Ready: true,
		}
		bundle.ArtifactRefs = appendBrowserArtifactRef(bundle.ArtifactRefs, bundle.PageSnapshot.Artifact)
	}
	if screenshotPath != "" {
		bundle.Screenshot = &BrowserScreenshotEvidence{
			Path:       screenshotPath,
			FinalURL:   finalURL,
			FullPage:   true,
			SourceTool: "browser_capture_submission_evidence",
			Artifact: BrowserArtifactRef{
				ArtifactID: browserEvidenceStateString(state, "review.evidence_artifact_id", "review.screenshot_artifact_id"),
				Type:       BrowserArtifactTypePageScreenshot,
				Path:       screenshotPath,
				StorageRef: browserEvidenceStateString(state, "review.evidence_storage_ref", "review.screenshot_storage_ref"),
				Role:       BrowserEvidenceKindScreenshot,
				SourceTool: "browser_capture_submission_evidence",
			},
			Ready: true,
		}
		bundle.ArtifactRefs = appendBrowserArtifactRef(bundle.ArtifactRefs, bundle.Screenshot.Artifact)
	}
	if tracePath != "" {
		bundle.ActionTrace = &BrowserActionTraceEvidence{
			Path:       tracePath,
			Status:     browserEvidenceStateString(state, "review.trace_status"),
			SourceTool: "browser_act",
			Artifact: BrowserArtifactRef{
				ArtifactID: browserEvidenceStateString(state, "review.trace_artifact_id", "review.action_trace_artifact_id"),
				Type:       BrowserArtifactTypeActionTrace,
				Path:       tracePath,
				StorageRef: browserEvidenceStateString(state, "review.trace_storage_ref", "review.action_trace_storage_ref"),
				Role:       BrowserEvidenceKindActionTrace,
				SourceTool: "browser_act",
			},
			Ready: true,
		}
		bundle.ArtifactRefs = appendBrowserArtifactRef(bundle.ArtifactRefs, bundle.ActionTrace.Artifact)
	}
	if downloadPath != "" {
		bundle.DownloadedFile = &BrowserDownloadedFileEvidence{
			Path:              downloadPath,
			FinalURL:          downloadFinalURL,
			Status:            downloadStatus,
			ContentType:       downloadContentType,
			ByteSize:          browserEvidenceStateInt64(state, "download.bytes", "download.byte_size"),
			SuggestedFilename: downloadFilename,
			Mode:              downloadMode,
			SourceTool:        "browser_download_file",
			Artifact: BrowserArtifactRef{
				ArtifactID: browserEvidenceStateString(state, "download.artifact_id"),
				Type:       BrowserArtifactTypeDownloadedFile,
				Path:       downloadPath,
				StorageRef: browserEvidenceStateString(state, "download.storage_ref"),
				MIMEType:   downloadContentType,
				Role:       BrowserEvidenceKindDownloadedFile,
				SourceTool: "browser_download_file",
			},
			Ready: true,
		}
		bundle.ArtifactRefs = appendBrowserArtifactRef(bundle.ArtifactRefs, bundle.DownloadedFile.Artifact)
	}
	return bundle
}

func EvaluateBrowserEvidenceBundle(bundle BrowserEvidenceBundle, req BrowserEvidenceRequirements) BrowserEvidenceReadiness {
	req = normalizeBrowserEvidenceRequirements(req, bundle)
	requiredChecks := 0
	passedChecks := 0
	evidence := []string{}
	reasons := []string{}

	out := BrowserEvidenceReadiness{
		SnapshotReady:       !req.RequireSnapshot,
		ScreenshotReady:     !req.RequireScreenshot,
		FinalURLReady:       !req.RequireFinalURL,
		ActionTraceReady:    !req.RequireActionTrace,
		DownloadedFileReady: !req.RequireDownloadedFile,
		ArtifactRefsReady:   !req.RequireArtifactRefs,
		TaskPlanReady:       !req.RequireTaskPlan,
		FailureReasonsReady: true,
		SubtaskCount:        len(browserTaskPlanSubtasks(bundle.TaskPlan)),
		SubtaskKinds:        browserTaskPlanKinds(bundle.TaskPlan),
	}
	if req.RequireSnapshot {
		requiredChecks++
		out.SnapshotReady = browserEvidenceSnapshotReady(bundle.PageSnapshot, req, &evidence, &reasons)
		if out.SnapshotReady {
			passedChecks++
		}
	}
	if req.RequireScreenshot {
		requiredChecks++
		out.ScreenshotReady = browserEvidenceScreenshotReady(bundle.Screenshot, &evidence, &reasons)
		if out.ScreenshotReady {
			passedChecks++
		}
	}
	if req.RequireFinalURL {
		requiredChecks++
		out.FinalURLReady = browserEvidenceFinalURLReady(bundle.FinalURL, req.TargetURL, &evidence, &reasons)
		if out.FinalURLReady {
			passedChecks++
		}
	}
	if req.RequireActionTrace {
		requiredChecks++
		out.ActionTraceReady = browserEvidenceActionTraceReady(bundle.ActionTrace, &evidence, &reasons)
		if out.ActionTraceReady {
			passedChecks++
		}
	}
	if req.RequireDownloadedFile {
		requiredChecks++
		out.DownloadedFileReady = browserEvidenceDownloadedFileReady(bundle.DownloadedFile, &evidence, &reasons)
		if out.DownloadedFileReady {
			passedChecks++
		}
	}
	if req.RequireArtifactRefs {
		requiredChecks++
		out.ArtifactRefsReady = browserArtifactRefsHaveLocator(bundle.ArtifactRefs)
		if out.ArtifactRefsReady {
			passedChecks++
			evidence = append(evidence, "artifact_refs_ready")
		} else {
			reasons = append(reasons, "artifact_refs_missing")
		}
	}
	if req.RequireTaskPlan {
		requiredChecks++
		out.TaskPlanReady = browserEvidenceTaskPlanReady(bundle.TaskPlan, req, &evidence, &reasons)
		if out.TaskPlanReady {
			passedChecks++
		}
	}
	if !req.AllowFailureReasons {
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
	if requiredChecks == 0 {
		requiredChecks = 1
		passedChecks = 1
		evidence = append(evidence, "no_required_browser_evidence")
	}
	out.FailureReasons = uniqueBrowserActionFailureReasons(reasons)
	out.Evidence = uniqueBrowserActionFailureReasons(evidence)
	out.Score = float64(passedChecks) / float64(requiredChecks)
	out.Passed = len(out.FailureReasons) == 0
	out.EvidenceReady = out.Passed
	return out
}

func BrowserVisualEvidenceInputFromBundle(bundle BrowserEvidenceBundle, req BrowserEvidenceRequirements) BrowserVisualEvidenceEvaluationInput {
	out := BrowserVisualEvidenceEvaluationInput{
		TargetURL:             firstNonEmptyBrowserEvidenceString(req.TargetURL, bundle.TargetURL),
		RequiredSnapshotTerms: append([]string(nil), req.RequiredSnapshotTerms...),
		RequireSnapshot:       req.RequireSnapshot,
		RequireScreenshot:     req.RequireScreenshot,
		RequireFinalURL:       req.RequireFinalURL,
		MinSnapshotChars:      req.MinSnapshotChars,
	}
	if bundle.PageSnapshot != nil {
		out.SnapshotText = bundle.PageSnapshot.Text
	}
	if bundle.Screenshot != nil {
		out.ScreenshotPath = bundle.Screenshot.Path
	}
	if bundle.FinalURL != nil {
		out.FinalURL = bundle.FinalURL.URL
	}
	return out
}

func BrowserEvidenceBundleSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"target_url": map[string]any{"type": "string"},
			"final_url": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url":         map[string]any{"type": "string"},
					"target_url":  map[string]any{"type": "string"},
					"source_tool": map[string]any{"type": "string"},
					"ready":       map[string]any{"type": "boolean"},
				},
			},
			"page_snapshot": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"text":          map[string]any{"type": "string"},
					"format":        map[string]any{"type": "string"},
					"element_count": map[string]any{"type": "integer"},
					"source_tool":   map[string]any{"type": "string"},
					"artifact":      browserArtifactRefSchema(),
					"ready":         map[string]any{"type": "boolean"},
				},
			},
			"screenshot": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":        map[string]any{"type": "string"},
					"final_url":   map[string]any{"type": "string"},
					"full_page":   map[string]any{"type": "boolean"},
					"source_tool": map[string]any{"type": "string"},
					"artifact":    browserArtifactRefSchema(),
					"ready":       map[string]any{"type": "boolean"},
				},
			},
			"action_trace": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"trace_id":      map[string]any{"type": "string"},
					"path":          map[string]any{"type": "string"},
					"status":        map[string]any{"type": "string"},
					"action_count":  map[string]any{"type": "integer"},
					"failure_count": map[string]any{"type": "integer"},
					"source_tool":   map[string]any{"type": "string"},
					"artifact":      browserArtifactRefSchema(),
					"ready":         map[string]any{"type": "boolean"},
				},
			},
			"downloaded_file": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":               map[string]any{"type": "string"},
					"final_url":          map[string]any{"type": "string"},
					"status":             map[string]any{"type": "string"},
					"content_type":       map[string]any{"type": "string"},
					"byte_size":          map[string]any{"type": "integer"},
					"suggested_filename": map[string]any{"type": "string"},
					"mode":               map[string]any{"type": "string"},
					"source_tool":        map[string]any{"type": "string"},
					"artifact":           browserArtifactRefSchema(),
					"ready":              map[string]any{"type": "boolean"},
				},
			},
			"task_plan": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"task_id":     map[string]any{"type": "string"},
					"goal":        map[string]any{"type": "string"},
					"case_type":   map[string]any{"type": "string"},
					"workflow_id": map[string]any{"type": "string"},
					"status":      map[string]any{"type": "string"},
					"source_tool": map[string]any{"type": "string"},
					"ready":       map[string]any{"type": "boolean"},
					"artifact":    browserArtifactRefSchema(),
					"subtasks": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"id":            map[string]any{"type": "string"},
								"kind":          map[string]any{"type": "string"},
								"description":   map[string]any{"type": "string"},
								"status":        map[string]any{"type": "string"},
								"evidence_kind": map[string]any{"type": "string"},
								"required":      map[string]any{"type": "boolean"},
								"evidence_refs": map[string]any{
									"type":  "array",
									"items": browserArtifactRefSchema(),
								},
							},
						},
					},
				},
			},
			"failure_reasons": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"code":          map[string]any{"type": "string"},
						"message":       map[string]any{"type": "string"},
						"action":        map[string]any{"type": "string"},
						"node_id":       map[string]any{"type": "string"},
						"tool":          map[string]any{"type": "string"},
						"recoverable":   map[string]any{"type": "boolean"},
						"evidence_kind": map[string]any{"type": "string"},
					},
				},
			},
			"artifact_refs": map[string]any{
				"type":  "array",
				"items": browserArtifactRefSchema(),
			},
		},
	}
}

func browserArtifactRefSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"artifact_id": map[string]any{"type": "string"},
			"type":        map[string]any{"type": "string"},
			"path":        map[string]any{"type": "string"},
			"storage_ref": map[string]any{"type": "string"},
			"url":         map[string]any{"type": "string"},
			"digest":      map[string]any{"type": "string"},
			"mime_type":   map[string]any{"type": "string"},
			"role":        map[string]any{"type": "string"},
			"source_tool": map[string]any{"type": "string"},
		},
	}
}

func normalizeBrowserEvidenceRequirements(req BrowserEvidenceRequirements, bundle BrowserEvidenceBundle) BrowserEvidenceRequirements {
	req.TargetURL = firstNonEmptyBrowserEvidenceString(req.TargetURL, bundle.TargetURL)
	req.RequiredSubtaskKinds = normalizeBrowserTaskPlanKinds(req.RequiredSubtaskKinds)
	if req.MinSubtasks > 0 || len(req.RequiredSubtaskKinds) > 0 {
		req.RequireTaskPlan = true
	}
	if !req.RequireSnapshot && !req.RequireScreenshot && !req.RequireFinalURL && !req.RequireActionTrace && !req.RequireDownloadedFile && !req.RequireArtifactRefs && !req.RequireTaskPlan {
		req.RequireSnapshot = true
		req.RequireScreenshot = true
		req.RequireFinalURL = true
	}
	return req
}

func browserEvidenceSnapshotReady(snapshot *BrowserPageSnapshotEvidence, req BrowserEvidenceRequirements, evidence *[]string, reasons *[]string) bool {
	if snapshot == nil || strings.TrimSpace(snapshot.Text) == "" {
		*reasons = append(*reasons, "snapshot_missing")
		return false
	}
	minChars := req.MinSnapshotChars
	if minChars < 0 {
		minChars = 0
	}
	if minChars > 0 && len(strings.TrimSpace(snapshot.Text)) < minChars {
		*reasons = append(*reasons, "snapshot_too_short")
		return false
	}
	lowerSnapshot := strings.ToLower(snapshot.Text)
	missingTerm := false
	for _, term := range req.RequiredSnapshotTerms {
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

func browserEvidenceScreenshotReady(screenshot *BrowserScreenshotEvidence, evidence *[]string, reasons *[]string) bool {
	if screenshot == nil {
		*reasons = append(*reasons, "screenshot_missing")
		return false
	}
	return browserVisualEvidenceScreenshotReady(screenshot.Path, evidence, reasons)
}

func browserEvidenceFinalURLReady(finalURL *BrowserFinalURLEvidence, targetURL string, evidence *[]string, reasons *[]string) bool {
	if finalURL == nil {
		*reasons = append(*reasons, "final_url_missing")
		return false
	}
	return browserVisualEvidenceFinalURLReady(finalURL.URL, firstNonEmptyBrowserEvidenceString(targetURL, finalURL.TargetURL), evidence, reasons)
}

func browserEvidenceActionTraceReady(trace *BrowserActionTraceEvidence, evidence *[]string, reasons *[]string) bool {
	if trace == nil {
		*reasons = append(*reasons, "action_trace_missing")
		return false
	}
	if trace.Ready || strings.TrimSpace(trace.TraceID) != "" || strings.TrimSpace(trace.Path) != "" ||
		browserArtifactRefHasLocator(trace.Artifact) {
		*evidence = append(*evidence, "action_trace_ready")
		return true
	}
	*reasons = append(*reasons, "action_trace_missing")
	return false
}

func browserEvidenceDownloadedFileReady(file *BrowserDownloadedFileEvidence, evidence *[]string, reasons *[]string) bool {
	if file == nil {
		*reasons = append(*reasons, "downloaded_file_missing")
		return false
	}
	if strings.TrimSpace(file.Path) == "" && !browserArtifactRefHasLocator(file.Artifact) {
		*reasons = append(*reasons, "downloaded_file_missing")
		return false
	}
	*evidence = append(*evidence, "downloaded_file_ready")
	return true
}

func browserEvidenceTaskPlanReady(plan *BrowserTaskPlanEvidence, req BrowserEvidenceRequirements, evidence *[]string, reasons *[]string) bool {
	taskReasons := []string{}
	if plan == nil {
		*reasons = append(*reasons, "task_plan_missing")
		return false
	}
	subtasks := browserTaskPlanSubtasks(plan)
	if strings.TrimSpace(plan.TaskID) == "" && strings.TrimSpace(plan.Goal) == "" && len(subtasks) == 0 {
		*reasons = append(*reasons, "task_plan_missing")
		return false
	}
	minSubtasks := req.MinSubtasks
	if minSubtasks < 0 {
		minSubtasks = 0
	}
	if minSubtasks > 0 && len(subtasks) < minSubtasks {
		*reasons = append(*reasons, "task_plan_subtasks_too_few")
		return false
	}
	switch status := strings.ToLower(strings.TrimSpace(plan.Status)); status {
	case "failed", "blocked", "missing", "planned":
		taskReasons = append(taskReasons, "task_plan_not_ready:"+status)
	}
	kinds := map[string]bool{}
	for _, subtask := range subtasks {
		kind := normalizeBrowserTaskPlanKind(subtask.Kind)
		if kind == "" {
			continue
		}
		kinds[kind] = true
		if subtask.Required && !browserSubtaskObservationReady(subtask) {
			id := firstNonEmptyBrowserEvidenceString(subtask.ID, kind)
			taskReasons = append(taskReasons, "task_plan_subtask_not_observed:"+id)
		}
	}
	for _, required := range req.RequiredSubtaskKinds {
		required = normalizeBrowserTaskPlanKind(required)
		if required == "" {
			continue
		}
		if !kinds[required] {
			taskReasons = append(taskReasons, "task_plan_subtask_kind_missing:"+required)
			continue
		}
		*evidence = append(*evidence, "task_subtask_kind:"+required)
	}
	if len(taskReasons) > 0 {
		*reasons = append(*reasons, taskReasons...)
		return false
	}
	*evidence = append(*evidence, "task_plan_ready")
	return true
}

func browserTaskPlanSubtasks(plan *BrowserTaskPlanEvidence) []BrowserSubtaskObservationEvidence {
	if plan == nil {
		return nil
	}
	out := make([]BrowserSubtaskObservationEvidence, 0, len(plan.Subtasks))
	for _, subtask := range plan.Subtasks {
		subtask.ID = strings.TrimSpace(subtask.ID)
		subtask.Kind = normalizeBrowserTaskPlanKind(subtask.Kind)
		subtask.Description = strings.TrimSpace(subtask.Description)
		subtask.Status = strings.TrimSpace(subtask.Status)
		subtask.EvidenceKind = strings.TrimSpace(subtask.EvidenceKind)
		subtask.EvidenceRefs = normalizeBrowserArtifactRefs(subtask.EvidenceRefs)
		if subtask.ID == "" && subtask.Kind == "" && subtask.Description == "" {
			continue
		}
		out = append(out, subtask)
	}
	return out
}

func browserTaskPlanKinds(plan *BrowserTaskPlanEvidence) []string {
	kinds := make([]string, 0)
	seen := map[string]bool{}
	for _, subtask := range browserTaskPlanSubtasks(plan) {
		kind := normalizeBrowserTaskPlanKind(subtask.Kind)
		if kind == "" || seen[kind] {
			continue
		}
		seen[kind] = true
		kinds = append(kinds, kind)
	}
	return kinds
}

func browserTaskPlanFanoutObserved(plan *BrowserTaskPlanEvidence) bool {
	subtasks := browserTaskPlanSubtasks(plan)
	if len(subtasks) < 2 {
		return false
	}
	for _, subtask := range subtasks {
		if subtask.Required && !browserSubtaskObservationReady(subtask) {
			return false
		}
	}
	return true
}

func browserSubtaskObservationReady(subtask BrowserSubtaskObservationEvidence) bool {
	status := strings.ToLower(strings.TrimSpace(subtask.Status))
	switch status {
	case "observed", "ready", "passed", "completed", "captured", "submitted", "downloaded", "extracted", "verified", "opened":
		return true
	case "failed", "blocked", "missing", "planned":
		return false
	}
	return strings.TrimSpace(subtask.EvidenceKind) != "" || len(normalizeBrowserArtifactRefs(subtask.EvidenceRefs)) > 0
}

func browserTaskPlanEvidenceFromState(raw any) *BrowserTaskPlanEvidence {
	switch value := raw.(type) {
	case *BrowserTaskPlanEvidence:
		return value
	case BrowserTaskPlanEvidence:
		return &value
	case map[string]any:
		plan := BrowserTaskPlanEvidence{
			TaskID:     browserAnyString(value["task_id"]),
			Goal:       browserAnyString(value["goal"]),
			CaseType:   browserAnyString(value["case_type"]),
			WorkflowID: browserAnyString(value["workflow_id"]),
			Status:     browserAnyString(value["status"]),
			SourceTool: browserAnyString(value["source_tool"]),
			Artifact:   browserArtifactRefFromAny(value["artifact"]),
			Ready:      browserAnyBool(value["ready"]),
			Subtasks:   browserSubtaskObservationsFromState(value["subtasks"]),
		}
		plan.Subtasks = browserTaskPlanSubtasks(&plan)
		if plan.TaskID == "" && plan.Goal == "" && len(plan.Subtasks) == 0 {
			return nil
		}
		return &plan
	default:
		return nil
	}
}

func browserSubtaskObservationsFromState(raw any) []BrowserSubtaskObservationEvidence {
	switch value := raw.(type) {
	case []BrowserSubtaskObservationEvidence:
		return value
	case []any:
		out := make([]BrowserSubtaskObservationEvidence, 0, len(value))
		for _, item := range value {
			if subtask, ok := browserSubtaskObservationFromAny(item); ok {
				out = append(out, subtask)
			}
		}
		return out
	default:
		if subtask, ok := browserSubtaskObservationFromAny(raw); ok {
			return []BrowserSubtaskObservationEvidence{subtask}
		}
		return nil
	}
}

func browserSubtaskObservationFromAny(raw any) (BrowserSubtaskObservationEvidence, bool) {
	switch value := raw.(type) {
	case BrowserSubtaskObservationEvidence:
		return value, value.ID != "" || value.Kind != "" || value.Description != ""
	case map[string]any:
		subtask := BrowserSubtaskObservationEvidence{
			ID:           browserAnyString(value["id"]),
			Kind:         browserAnyString(value["kind"]),
			Description:  browserAnyString(value["description"]),
			Status:       browserAnyString(value["status"]),
			EvidenceKind: browserAnyString(value["evidence_kind"]),
			Required:     browserAnyBool(value["required"]),
			EvidenceRefs: browserArtifactRefsFromState(value["evidence_refs"]),
		}
		return subtask, subtask.ID != "" || subtask.Kind != "" || subtask.Description != ""
	default:
		return BrowserSubtaskObservationEvidence{}, false
	}
}

func normalizeBrowserTaskPlanKinds(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		kind := normalizeBrowserTaskPlanKind(value)
		if kind == "" || seen[kind] {
			continue
		}
		seen[kind] = true
		out = append(out, kind)
	}
	return out
}

func normalizeBrowserTaskPlanKind(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.ReplaceAll(value, " ", "_")
	return value
}

func browserArtifactRefsFromState(raw any) []BrowserArtifactRef {
	switch value := raw.(type) {
	case []BrowserArtifactRef:
		return normalizeBrowserArtifactRefs(value)
	case []any:
		out := []BrowserArtifactRef{}
		for _, item := range value {
			if ref := browserArtifactRefFromAny(item); browserArtifactRefHasLocator(ref) {
				out = append(out, ref)
			}
		}
		return normalizeBrowserArtifactRefs(out)
	default:
		if ref := browserArtifactRefFromAny(raw); browserArtifactRefHasLocator(ref) {
			return []BrowserArtifactRef{ref}
		}
		return nil
	}
}

func browserArtifactRefFromAny(raw any) BrowserArtifactRef {
	switch value := raw.(type) {
	case BrowserArtifactRef:
		return value
	case map[string]any:
		return BrowserArtifactRef{
			ArtifactID: browserAnyString(value["artifact_id"]),
			Type:       browserAnyString(value["type"]),
			Path:       browserAnyString(value["path"]),
			StorageRef: browserAnyString(value["storage_ref"]),
			URL:        browserAnyString(value["url"]),
			Digest:     browserAnyString(value["digest"]),
			MIMEType:   browserAnyString(value["mime_type"]),
			Role:       browserAnyString(value["role"]),
			SourceTool: browserAnyString(value["source_tool"]),
		}
	default:
		return BrowserArtifactRef{}
	}
}

func browserFailureReasonEvidenceFromState(raw any) []BrowserFailureReasonEvidence {
	switch value := raw.(type) {
	case []BrowserFailureReasonEvidence:
		return value
	case []string:
		out := make([]BrowserFailureReasonEvidence, 0, len(value))
		for _, item := range value {
			if item = strings.TrimSpace(item); item != "" {
				out = append(out, BrowserFailureReasonEvidence{Code: item, EvidenceKind: BrowserEvidenceKindFailureReason})
			}
		}
		return out
	case []any:
		out := []BrowserFailureReasonEvidence{}
		for _, item := range value {
			if reason := browserFailureReasonEvidenceFromAny(item); reason.Code != "" || reason.Message != "" {
				out = append(out, reason)
			}
		}
		return out
	default:
		if reason := browserFailureReasonEvidenceFromAny(raw); reason.Code != "" || reason.Message != "" {
			return []BrowserFailureReasonEvidence{reason}
		}
		return nil
	}
}

func browserFailureReasonEvidenceFromAny(raw any) BrowserFailureReasonEvidence {
	switch value := raw.(type) {
	case BrowserFailureReasonEvidence:
		return value
	case string:
		return BrowserFailureReasonEvidence{Code: strings.TrimSpace(value), EvidenceKind: BrowserEvidenceKindFailureReason}
	case map[string]any:
		return BrowserFailureReasonEvidence{
			Code:         browserAnyString(value["code"]),
			Message:      browserAnyString(value["message"]),
			Action:       browserAnyString(value["action"]),
			NodeID:       browserAnyString(value["node_id"]),
			Tool:         browserAnyString(value["tool"]),
			Recoverable:  browserAnyBool(value["recoverable"]),
			EvidenceKind: firstNonEmptyBrowserEvidenceString(browserAnyString(value["evidence_kind"]), BrowserEvidenceKindFailureReason),
		}
	default:
		return BrowserFailureReasonEvidence{}
	}
}

func appendBrowserArtifactRef(values []BrowserArtifactRef, ref BrowserArtifactRef) []BrowserArtifactRef {
	if !browserArtifactRefHasLocator(ref) {
		return values
	}
	return normalizeBrowserArtifactRefs(append(values, ref))
}

func normalizeBrowserArtifactRefs(values []BrowserArtifactRef) []BrowserArtifactRef {
	out := []BrowserArtifactRef{}
	seen := map[string]bool{}
	for _, value := range values {
		value.ArtifactID = strings.TrimSpace(value.ArtifactID)
		value.Type = strings.TrimSpace(value.Type)
		value.Path = strings.TrimSpace(value.Path)
		value.StorageRef = strings.TrimSpace(value.StorageRef)
		value.URL = strings.TrimSpace(value.URL)
		value.Digest = strings.TrimSpace(value.Digest)
		value.MIMEType = strings.TrimSpace(value.MIMEType)
		value.Role = strings.TrimSpace(value.Role)
		value.SourceTool = strings.TrimSpace(value.SourceTool)
		if !browserArtifactRefHasLocator(value) {
			continue
		}
		key := strings.Join([]string{value.ArtifactID, value.Type, value.Path, value.StorageRef, value.URL, value.Digest, value.Role, value.SourceTool}, "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, value)
	}
	return out
}

func browserArtifactRefsHaveLocator(values []BrowserArtifactRef) bool {
	for _, value := range normalizeBrowserArtifactRefs(values) {
		if browserArtifactRefHasLocator(value) {
			return true
		}
	}
	return false
}

func browserArtifactRefHasLocator(value BrowserArtifactRef) bool {
	return strings.TrimSpace(value.ArtifactID) != "" ||
		strings.TrimSpace(value.Path) != "" ||
		strings.TrimSpace(value.StorageRef) != "" ||
		strings.TrimSpace(value.URL) != ""
}

func browserEvidenceStateString(state map[string]any, paths ...string) string {
	for _, path := range paths {
		if value := browserAnyString(browserEvidenceStateAny(state, path)); value != "" {
			return value
		}
	}
	return ""
}

func firstNonNilBrowserEvidenceStateAny(state map[string]any, paths ...string) any {
	for _, path := range paths {
		if value := browserEvidenceStateAny(state, path); value != nil {
			return value
		}
	}
	return nil
}

func browserEvidenceStateAny(state map[string]any, path string) any {
	if state == nil || strings.TrimSpace(path) == "" {
		return nil
	}
	if value, ok := state[path]; ok {
		return value
	}
	parts := strings.Split(path, ".")
	var current any = state
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil
		}
		switch node := current.(type) {
		case map[string]any:
			current = node[part]
		default:
			return nil
		}
	}
	return current
}

func browserAnyString(raw any) string {
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case []byte:
		return strings.TrimSpace(string(value))
	default:
		return ""
	}
}

func browserAnyBool(raw any) bool {
	value, _ := raw.(bool)
	return value
}

func browserEvidenceStateInt64(state map[string]any, paths ...string) int64 {
	for _, path := range paths {
		if value := browserAnyInt64(browserEvidenceStateAny(state, path)); value > 0 {
			return value
		}
	}
	return 0
}

func browserAnyInt64(raw any) int64 {
	switch value := raw.(type) {
	case int:
		return int64(value)
	case int64:
		return value
	case int32:
		return int64(value)
	case float64:
		return int64(value)
	case float32:
		return int64(value)
	default:
		return 0
	}
}

func firstNonEmptyBrowserEvidenceString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
