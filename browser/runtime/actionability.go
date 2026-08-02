package browserruntime

import "strings"

const (
	BrowserActionabilityStatusPassed      = "passed"
	BrowserActionabilityStatusFailed      = "failed"
	BrowserActionabilityStatusPartial     = "partial"
	BrowserActionabilityStatusSkipped     = "skipped"
	BrowserActionabilityStatusNotReported = "not_reported"
)

// BrowserActionabilityCheck captures one Playwright-style precondition for a
// browser action. Backends may progressively fill precise checks; the shared
// contract still makes missing evidence explicit instead of silently implying
// success.
type BrowserActionabilityCheck struct {
	Name     string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"`
	Required bool   `json:"required,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

// BrowserActionabilityReport is the shared actionability surface for target
// actions. It is intentionally backend-agnostic: concrete CDP/Playwright-like
// backends can report richer checks later while tools already expose a stable
// failure and diagnostic shape.
type BrowserActionabilityReport struct {
	Action           string                      `json:"action,omitempty"`
	Status           string                      `json:"status,omitempty"`
	TargetKind       string                      `json:"target_kind,omitempty"`
	Target           string                      `json:"target,omitempty"`
	FailedCheck      string                      `json:"failed_check,omitempty"`
	FailureReason    string                      `json:"failure_reason,omitempty"`
	RetryDisposition string                      `json:"retry_disposition,omitempty"`
	ManualRetryHint  string                      `json:"manual_retry_hint,omitempty"`
	RecoveryAction   string                      `json:"recovery_action,omitempty"`
	Checks           []BrowserActionabilityCheck `json:"checks,omitempty"`
}

// BrowserActionFailureArtifact is the compact trace-like index attached to a
// failed action. It does not own persistence; it summarizes the evidence that
// already exists in the action result so callers can inspect failures without
// parsing free-form notes.
type BrowserActionFailureArtifact struct {
	Kind                 string                         `json:"kind,omitempty"`
	Action               string                         `json:"action,omitempty"`
	Status               string                         `json:"status,omitempty"`
	ReasonCode           string                         `json:"reason_code,omitempty"`
	Message              string                         `json:"message,omitempty"`
	RecoveryAction       string                         `json:"recovery_action,omitempty"`
	FinalURL             string                         `json:"final_url,omitempty"`
	Title                string                         `json:"title,omitempty"`
	TargetKind           string                         `json:"target_kind,omitempty"`
	Target               string                         `json:"target,omitempty"`
	FailedCheck          string                         `json:"failed_check,omitempty"`
	SnapshotAvailable    bool                           `json:"snapshot_available,omitempty"`
	SnapshotFormat       string                         `json:"snapshot_format,omitempty"`
	SnapshotRefs         string                         `json:"snapshot_refs,omitempty"`
	SnapshotFrame        string                         `json:"snapshot_frame,omitempty"`
	SnapshotElementCount int                            `json:"snapshot_element_count,omitempty"`
	SnapshotTruncated    bool                           `json:"snapshot_truncated,omitempty"`
	SnapshotFreshness    *BrowserSnapshotFreshness      `json:"snapshot_freshness,omitempty"`
	RecoveryTranscript   *BrowserStepRecoveryTranscript `json:"recovery_transcript,omitempty"`
	ScreenshotPath       string                         `json:"screenshot_path,omitempty"`
	ArtifactPath         string                         `json:"artifact_path,omitempty"`
	ResolverOutcome      *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	ConsoleMessageCount  int                            `json:"console_message_count,omitempty"`
	ErrorCount           int                            `json:"error_count,omitempty"`
	ConsoleMessages      []BrowserConsoleMessage        `json:"console_messages,omitempty"`
	Errors               []BrowserErrorEntry            `json:"errors,omitempty"`
}

// BrowserActionFailureEvidence is emitted for failed browser actions or failed
// target resolution. It keeps the actionable evidence compact enough for tool
// payloads while preserving the resolver outcome and snapshot availability.
type BrowserActionFailureEvidence struct {
	Action               string                         `json:"action,omitempty"`
	Status               string                         `json:"status,omitempty"`
	ReasonCode           string                         `json:"reason_code,omitempty"`
	Message              string                         `json:"message,omitempty"`
	RecoveryAction       string                         `json:"recovery_action,omitempty"`
	Retryable            bool                           `json:"retryable,omitempty"`
	ResolverOutcome      *BrowserElementResolverOutcome `json:"resolver_outcome,omitempty"`
	Actionability        *BrowserActionabilityReport    `json:"actionability,omitempty"`
	SnapshotAvailable    bool                           `json:"snapshot_available,omitempty"`
	SnapshotFormat       string                         `json:"snapshot_format,omitempty"`
	SnapshotRefs         string                         `json:"snapshot_refs,omitempty"`
	SnapshotFrame        string                         `json:"snapshot_frame,omitempty"`
	SnapshotElementCount int                            `json:"snapshot_element_count,omitempty"`
	SnapshotTruncated    bool                           `json:"snapshot_truncated,omitempty"`
	SnapshotFreshness    *BrowserSnapshotFreshness      `json:"snapshot_freshness,omitempty"`
	RecoveryTranscript   *BrowserStepRecoveryTranscript `json:"recovery_transcript,omitempty"`
	ArtifactPath         string                         `json:"artifact_path,omitempty"`
	Artifact             *BrowserActionFailureArtifact  `json:"artifact,omitempty"`
}

// BrowserActionabilityReportRequest carries the known evidence for building a
// shared actionability report.
type BrowserActionabilityReportRequest struct {
	Action          string
	ElementRef      string
	Selector        string
	ResolverOutcome *BrowserElementResolverOutcome
}

// BrowserActionFailureEvidenceRequest carries action-local evidence used to
// construct a compact failure artifact for tool payloads.
type BrowserActionFailureEvidenceRequest struct {
	Action            string
	Status            string
	Note              string
	RecoveryAction    string
	ElementRef        string
	Selector          string
	ResolverOutcome   *BrowserElementResolverOutcome
	Actionability     *BrowserActionabilityReport
	SnapshotFormat    string
	SnapshotRefs      string
	SnapshotFrame     string
	ElementCount      int
	SnapshotText      string
	SnapshotTruncated bool
	ArtifactPath      string
	FinalURL          string
	Title             string
	ConsoleMessages   []BrowserConsoleMessage
	Errors            []BrowserErrorEntry
}

// BuildBrowserActionabilityReport builds the common target-action
// actionability surface. It reports target resolution precisely when resolver
// evidence exists and marks backend-specific checks as not_reported until a
// backend supplies concrete evidence.
func BuildBrowserActionabilityReport(req BrowserActionabilityReportRequest) *BrowserActionabilityReport {
	action := normalizeBrowserActionabilityAction(req.Action)
	if action == "" {
		return nil
	}
	targetKind, target := browserActionabilityTarget(req.ElementRef, req.Selector)
	outcome := browserElementResolverOutcomeNormalizedClone(req.ResolverOutcome)
	if target == "" && outcome == nil {
		return nil
	}

	report := &BrowserActionabilityReport{
		Action:     action,
		TargetKind: targetKind,
		Target:     target,
		Status:     BrowserActionabilityStatusPartial,
	}

	resolveCheck := BrowserActionabilityCheck{
		Name:     "resolve_target",
		Required: true,
		Status:   BrowserActionabilityStatusNotReported,
	}
	resolutionFailed := false
	if outcome != nil {
		resolveCheck.Detail = strings.TrimSpace(outcome.Note)
		switch strings.TrimSpace(outcome.Status) {
		case "matched":
			resolveCheck.Status = BrowserActionabilityStatusPassed
		case "":
			resolveCheck.Status = BrowserActionabilityStatusNotReported
		default:
			resolveCheck.Status = BrowserActionabilityStatusFailed
			resolutionFailed = true
			report.Status = BrowserActionabilityStatusFailed
			report.FailedCheck = "resolve_target"
			report.FailureReason = browserActionabilityFailureReason(outcome)
			report.RetryDisposition = strings.TrimSpace(outcome.RetryDisposition)
			report.ManualRetryHint = strings.TrimSpace(outcome.ManualRetryHint)
			report.RecoveryAction = strings.TrimSpace(outcome.RecoveryAction)
		}
	}
	report.Checks = append(report.Checks, resolveCheck)

	for _, name := range browserActionabilityRequiredChecks(action) {
		check := BrowserActionabilityCheck{
			Name:     name,
			Required: true,
			Status:   BrowserActionabilityStatusNotReported,
		}
		if resolutionFailed {
			check.Status = BrowserActionabilityStatusSkipped
			check.Detail = "blocked by target resolution"
		}
		report.Checks = append(report.Checks, check)
	}

	if !resolutionFailed && browserActionabilityAllRequiredPassed(report.Checks) {
		report.Status = BrowserActionabilityStatusPassed
	}
	return report
}

// BuildBrowserActionFailureEvidence returns nil for successful actions and a
// compact evidence object for target/action failures.
func BuildBrowserActionFailureEvidence(req BrowserActionFailureEvidenceRequest) *BrowserActionFailureEvidence {
	action := normalizeBrowserActionabilityAction(req.Action)
	if action == "" {
		return nil
	}
	outcome := browserElementResolverOutcomeNormalizedClone(req.ResolverOutcome)
	actionability := req.Actionability
	if actionability == nil {
		actionability = BuildBrowserActionabilityReport(BrowserActionabilityReportRequest{
			Action:          action,
			ElementRef:      req.ElementRef,
			Selector:        req.Selector,
			ResolverOutcome: outcome,
		})
	}

	reasonCode := ""
	retryable := false
	if actionability != nil && actionability.Status == BrowserActionabilityStatusFailed {
		reasonCode = firstNonEmptyString(strings.TrimSpace(actionability.FailureReason), "actionability_failed")
		retryable = browserActionFailureRetryable(reasonCode)
	}
	if reasonCode == "" && outcome != nil {
		status := strings.TrimSpace(outcome.Status)
		if status != "" && status != "matched" {
			reasonCode = "resolver_" + status
			retryable = browserActionFailureRetryable(status)
		}
	}
	if reasonCode == "" {
		status := strings.TrimSpace(req.Status)
		if browserActionStatusIndicatesFailure(status) {
			reasonCode = "action_" + status
			retryable = browserActionFailureRetryable(status)
		}
	}
	if reasonCode == "" {
		return nil
	}

	recoveryAction := strings.TrimSpace(req.RecoveryAction)
	if recoveryAction == "" && actionability != nil {
		recoveryAction = strings.TrimSpace(actionability.RecoveryAction)
	}
	if recoveryAction == "" && outcome != nil {
		recoveryAction = strings.TrimSpace(outcome.RecoveryAction)
	}
	message := strings.TrimSpace(req.Note)
	if message == "" && outcome != nil {
		message = strings.TrimSpace(outcome.Note)
	}
	if message == "" && actionability != nil {
		message = strings.TrimSpace(actionability.FailureReason)
	}

	snapshotAvailable := strings.TrimSpace(req.SnapshotText) != "" ||
		strings.TrimSpace(req.SnapshotFormat) != "" ||
		strings.TrimSpace(req.SnapshotRefs) != "" ||
		req.ElementCount > 0

	evidence := &BrowserActionFailureEvidence{
		Action:               action,
		Status:               strings.TrimSpace(req.Status),
		ReasonCode:           reasonCode,
		Message:              message,
		RecoveryAction:       recoveryAction,
		Retryable:            retryable,
		ResolverOutcome:      outcome,
		Actionability:        actionability,
		SnapshotAvailable:    snapshotAvailable,
		SnapshotFormat:       strings.TrimSpace(req.SnapshotFormat),
		SnapshotRefs:         strings.TrimSpace(req.SnapshotRefs),
		SnapshotFrame:        strings.TrimSpace(req.SnapshotFrame),
		SnapshotElementCount: req.ElementCount,
		SnapshotTruncated:    req.SnapshotTruncated,
		SnapshotFreshness: BuildBrowserSnapshotFreshness(BrowserSnapshotFreshnessRequest{
			Action:                    action,
			TargetKind:                browserActionFailureEvidenceTargetKind(req.ElementRef, req.Selector),
			Target:                    firstNonEmptyString(strings.TrimSpace(req.ElementRef), strings.TrimSpace(req.Selector)),
			SnapshotSource:            browserActionFailureEvidenceSnapshotSource(snapshotAvailable),
			PageURL:                   req.FinalURL,
			PageTitle:                 req.Title,
			ResolverOutcome:           outcome,
			RecoverySnapshotAvailable: snapshotAvailable,
		}),
		ArtifactPath: strings.TrimSpace(req.ArtifactPath),
	}
	targetKind, target := browserActionFailureArtifactTarget(req, evidence.Actionability)
	evidence.RecoveryTranscript = BuildBrowserStepRecoveryTranscript(BrowserStepRecoveryTranscriptRequest{
		Action:            evidence.Action,
		Status:            evidence.Status,
		ReasonCode:        evidence.ReasonCode,
		Message:           evidence.Message,
		RecoveryAction:    evidence.RecoveryAction,
		Retryable:         evidence.Retryable,
		TargetKind:        targetKind,
		Target:            target,
		ResolverOutcome:   evidence.ResolverOutcome,
		Actionability:     evidence.Actionability,
		SnapshotFreshness: evidence.SnapshotFreshness,
	})
	evidence.Artifact = buildBrowserActionFailureArtifact(req, evidence)
	return evidence
}

func buildBrowserActionFailureArtifact(
	req BrowserActionFailureEvidenceRequest,
	evidence *BrowserActionFailureEvidence,
) *BrowserActionFailureArtifact {
	if evidence == nil {
		return nil
	}
	targetKind, target := browserActionFailureArtifactTarget(req, evidence.Actionability)
	consoleMessages := cloneBrowserConsoleMessages(req.ConsoleMessages)
	errors := cloneBrowserErrorEntries(req.Errors)
	artifact := &BrowserActionFailureArtifact{
		Kind:                 "trace_like",
		Action:               strings.TrimSpace(evidence.Action),
		Status:               strings.TrimSpace(evidence.Status),
		ReasonCode:           strings.TrimSpace(evidence.ReasonCode),
		Message:              strings.TrimSpace(evidence.Message),
		RecoveryAction:       strings.TrimSpace(evidence.RecoveryAction),
		FinalURL:             strings.TrimSpace(req.FinalURL),
		Title:                strings.TrimSpace(req.Title),
		TargetKind:           targetKind,
		Target:               target,
		FailedCheck:          browserActionFailureArtifactFailedCheck(evidence.Actionability),
		SnapshotAvailable:    evidence.SnapshotAvailable,
		SnapshotFormat:       strings.TrimSpace(evidence.SnapshotFormat),
		SnapshotRefs:         strings.TrimSpace(evidence.SnapshotRefs),
		SnapshotFrame:        strings.TrimSpace(evidence.SnapshotFrame),
		SnapshotElementCount: evidence.SnapshotElementCount,
		SnapshotTruncated:    evidence.SnapshotTruncated,
		SnapshotFreshness:    cloneBrowserSnapshotFreshness(evidence.SnapshotFreshness),
		RecoveryTranscript:   cloneBrowserStepRecoveryTranscript(evidence.RecoveryTranscript),
		ScreenshotPath:       browserActionFailureArtifactScreenshotPath(evidence),
		ArtifactPath:         strings.TrimSpace(evidence.ArtifactPath),
		ResolverOutcome:      browserElementResolverOutcomeNormalizedClone(evidence.ResolverOutcome),
		ConsoleMessageCount:  len(consoleMessages),
		ErrorCount:           len(errors),
		ConsoleMessages:      consoleMessages,
		Errors:               errors,
	}
	if browserActionFailureArtifactEmpty(*artifact) {
		return nil
	}
	return artifact
}

func browserActionFailureArtifactScreenshotPath(evidence *BrowserActionFailureEvidence) string {
	if evidence == nil || !strings.EqualFold(strings.TrimSpace(evidence.Action), "screenshot") {
		return ""
	}
	return strings.TrimSpace(evidence.ArtifactPath)
}

func browserActionFailureArtifactTarget(
	req BrowserActionFailureEvidenceRequest,
	actionability *BrowserActionabilityReport,
) (string, string) {
	if actionability != nil {
		if kind := strings.TrimSpace(actionability.TargetKind); kind != "" {
			return kind, strings.TrimSpace(actionability.Target)
		}
		if target := strings.TrimSpace(actionability.Target); target != "" {
			return "", target
		}
	}
	return browserActionabilityTarget(req.ElementRef, req.Selector)
}

func browserActionFailureEvidenceTargetKind(elementRef string, selector string) string {
	targetKind, _ := browserActionabilityTarget(elementRef, selector)
	return targetKind
}

func browserActionFailureEvidenceSnapshotSource(snapshotAvailable bool) string {
	if snapshotAvailable {
		return BrowserSnapshotFreshnessSourceRecoverySnapshot
	}
	return ""
}

func browserActionFailureArtifactFailedCheck(actionability *BrowserActionabilityReport) string {
	if actionability == nil {
		return ""
	}
	return strings.TrimSpace(actionability.FailedCheck)
}

func browserActionFailureArtifactEmpty(artifact BrowserActionFailureArtifact) bool {
	return strings.TrimSpace(artifact.Kind) == "" &&
		strings.TrimSpace(artifact.Action) == "" &&
		strings.TrimSpace(artifact.Status) == "" &&
		strings.TrimSpace(artifact.ReasonCode) == "" &&
		strings.TrimSpace(artifact.Message) == "" &&
		strings.TrimSpace(artifact.RecoveryAction) == "" &&
		strings.TrimSpace(artifact.FinalURL) == "" &&
		strings.TrimSpace(artifact.Title) == "" &&
		strings.TrimSpace(artifact.TargetKind) == "" &&
		strings.TrimSpace(artifact.Target) == "" &&
		strings.TrimSpace(artifact.FailedCheck) == "" &&
		!artifact.SnapshotAvailable &&
		strings.TrimSpace(artifact.SnapshotFormat) == "" &&
		strings.TrimSpace(artifact.SnapshotRefs) == "" &&
		strings.TrimSpace(artifact.SnapshotFrame) == "" &&
		artifact.SnapshotElementCount == 0 &&
		!artifact.SnapshotTruncated &&
		artifact.SnapshotFreshness == nil &&
		artifact.RecoveryTranscript == nil &&
		strings.TrimSpace(artifact.ScreenshotPath) == "" &&
		strings.TrimSpace(artifact.ArtifactPath) == "" &&
		artifact.ResolverOutcome == nil &&
		artifact.ConsoleMessageCount == 0 &&
		artifact.ErrorCount == 0 &&
		len(artifact.ConsoleMessages) == 0 &&
		len(artifact.Errors) == 0
}

func cloneBrowserConsoleMessages(messages []BrowserConsoleMessage) []BrowserConsoleMessage {
	if len(messages) == 0 {
		return nil
	}
	return append([]BrowserConsoleMessage(nil), messages...)
}

func cloneBrowserErrorEntries(entries []BrowserErrorEntry) []BrowserErrorEntry {
	if len(entries) == 0 {
		return nil
	}
	return append([]BrowserErrorEntry(nil), entries...)
}

func normalizeBrowserActionabilityAction(action string) string {
	action = strings.ToLower(strings.TrimSpace(action))
	switch action {
	case "type_text":
		return "type"
	case "click", "type", "screenshot", "highlight", "upload", "hover", "drag", "select", "fill":
		return action
	default:
		return action
	}
}

func browserActionabilityTarget(elementRef string, selector string) (string, string) {
	if ref := strings.TrimSpace(elementRef); ref != "" {
		return "ref", ref
	}
	if selector := strings.TrimSpace(selector); selector != "" {
		return "selector", selector
	}
	return "", ""
}

func browserActionabilityRequiredChecks(action string) []string {
	switch strings.TrimSpace(action) {
	case "click":
		return []string{"attached", "visible", "stable", "receives_events", "enabled", "frame_hit_target", "navigation_wait"}
	case "type", "fill":
		return []string{"attached", "visible", "stable", "enabled", "editable"}
	case "select":
		return []string{"attached", "visible", "stable", "enabled"}
	case "upload":
		return []string{"attached", "enabled"}
	case "hover", "drag":
		return []string{"attached", "visible", "stable", "receives_events", "frame_hit_target"}
	case "screenshot", "highlight":
		return []string{"attached", "visible", "stable"}
	default:
		return nil
	}
}

func browserActionabilityAllRequiredPassed(checks []BrowserActionabilityCheck) bool {
	if len(checks) == 0 {
		return false
	}
	for _, check := range checks {
		if check.Required && check.Status != BrowserActionabilityStatusPassed {
			return false
		}
	}
	return true
}

func browserActionabilityFailureReason(outcome *BrowserElementResolverOutcome) string {
	if outcome == nil {
		return ""
	}
	status := strings.TrimSpace(outcome.Status)
	blockedBy := strings.TrimSpace(outcome.BlockedBy)
	switch {
	case status != "" && blockedBy != "":
		return "resolver_" + status + "_" + blockedBy
	case status != "":
		return "resolver_" + status
	default:
		return ""
	}
}

func browserActionStatusIndicatesFailure(status string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "", "ok", "opened", "navigated", "extracted", "snapshotted", "captured", "clicked", "typed", "evaluated", "highlighted", "uploaded", "pressed", "hovered", "dragged", "selected", "filled", "resized", "waited", "listed", "focused", "closed", "review_required", "armed", "started", "stopped":
		return false
	default:
		return strings.Contains(status, "error") ||
			strings.Contains(status, "failed") ||
			strings.Contains(status, "unresolved") ||
			strings.Contains(status, "blocked") ||
			strings.Contains(status, "timeout") ||
			strings.Contains(status, "invalid") ||
			strings.Contains(status, "not_found")
	}
}

func browserActionFailureRetryable(reason string) bool {
	reason = strings.ToLower(strings.TrimSpace(reason))
	return strings.Contains(reason, "unresolved") ||
		strings.Contains(reason, "resolution_failed") ||
		strings.Contains(reason, "page_binding") ||
		strings.Contains(reason, "stale") ||
		strings.Contains(reason, "timeout") ||
		strings.Contains(reason, "not_reported")
}
