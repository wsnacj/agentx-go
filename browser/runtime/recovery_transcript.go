package browserruntime

import "strings"

const (
	BrowserStepRecoveryTranscriptKind = "browser_step_recovery"

	BrowserStepRecoveryTranscriptPhaseFailedAction        = "failed_action"
	BrowserStepRecoveryTranscriptPhaseActionability       = "actionability"
	BrowserStepRecoveryTranscriptPhaseTargetResolution    = "target_resolution"
	BrowserStepRecoveryTranscriptPhaseSnapshotFreshness   = "snapshot_freshness"
	BrowserStepRecoveryTranscriptPhaseRecommendedNextStep = "recommended_next_step"
)

// BrowserStepRecoveryTranscript is a compact action-local recovery timeline.
// It is derived from the failure evidence already in the payload and does not
// own persistence, scheduling, retries, or browser session state.
type BrowserStepRecoveryTranscript struct {
	Kind            string                              `json:"kind,omitempty"`
	Action          string                              `json:"action,omitempty"`
	Status          string                              `json:"status,omitempty"`
	ReasonCode      string                              `json:"reason_code,omitempty"`
	TargetKind      string                              `json:"target_kind,omitempty"`
	Target          string                              `json:"target,omitempty"`
	FailedCheck     string                              `json:"failed_check,omitempty"`
	SnapshotState   string                              `json:"snapshot_state,omitempty"`
	RecoveryAction  string                              `json:"recovery_action,omitempty"`
	NextStepAlias   string                              `json:"next_step_alias,omitempty"`
	ManualRetryHint string                              `json:"manual_retry_hint,omitempty"`
	Retryable       bool                                `json:"retryable,omitempty"`
	Steps           []BrowserStepRecoveryTranscriptStep `json:"steps,omitempty"`
}

// BrowserStepRecoveryTranscriptStep records one compact step in the recovery
// timeline. EvidenceSource names the payload section the step was derived from.
type BrowserStepRecoveryTranscriptStep struct {
	Index           int    `json:"index,omitempty"`
	Phase           string `json:"phase,omitempty"`
	Action          string `json:"action,omitempty"`
	State           string `json:"state,omitempty"`
	Reason          string `json:"reason,omitempty"`
	EvidenceSource  string `json:"evidence_source,omitempty"`
	FailedCheck     string `json:"failed_check,omitempty"`
	RecoveryAction  string `json:"recovery_action,omitempty"`
	NextStepAlias   string `json:"next_step_alias,omitempty"`
	ManualRetryHint string `json:"manual_retry_hint,omitempty"`
}

// BrowserStepRecoveryTranscriptRequest carries the already-built failure
// evidence needed to project a compact recovery transcript.
type BrowserStepRecoveryTranscriptRequest struct {
	Action            string
	Status            string
	ReasonCode        string
	Message           string
	RecoveryAction    string
	Retryable         bool
	TargetKind        string
	Target            string
	ResolverOutcome   *BrowserElementResolverOutcome
	Actionability     *BrowserActionabilityReport
	SnapshotFreshness *BrowserSnapshotFreshness
}

// BuildBrowserStepRecoveryTranscript derives an ordered recovery transcript
// from existing failure evidence. It intentionally avoids inventing runtime
// policy; callers still decide whether and how to act on the guidance.
func BuildBrowserStepRecoveryTranscript(req BrowserStepRecoveryTranscriptRequest) *BrowserStepRecoveryTranscript {
	action := normalizeBrowserActionabilityAction(req.Action)
	status := strings.TrimSpace(req.Status)
	reasonCode := strings.TrimSpace(req.ReasonCode)
	message := strings.TrimSpace(req.Message)
	actionability := req.Actionability
	outcome := browserElementResolverOutcomeNormalizedClone(req.ResolverOutcome)
	freshness := cloneBrowserSnapshotFreshness(req.SnapshotFreshness)
	if action == "" && status == "" && reasonCode == "" && message == "" && actionability == nil && outcome == nil && freshness == nil {
		return nil
	}

	transcript := &BrowserStepRecoveryTranscript{
		Kind:       BrowserStepRecoveryTranscriptKind,
		Action:     action,
		Status:     status,
		ReasonCode: reasonCode,
		TargetKind: strings.TrimSpace(req.TargetKind),
		Target:     strings.TrimSpace(req.Target),
		Retryable:  req.Retryable,
	}
	if actionability != nil {
		transcript.FailedCheck = strings.TrimSpace(actionability.FailedCheck)
	}
	transcript.SnapshotState = browserStepRecoverySnapshotState(freshness)
	transcript.RecoveryAction = browserStepRecoveryRecoveryAction(req.RecoveryAction, actionability, outcome, freshness)
	transcript.NextStepAlias = browserStepRecoveryNextStepAlias(actionability, outcome, freshness, transcript.RecoveryAction)
	transcript.ManualRetryHint = browserStepRecoveryManualRetryHint(actionability, outcome)

	transcript.addStep(BrowserStepRecoveryTranscriptStep{
		Phase:          BrowserStepRecoveryTranscriptPhaseFailedAction,
		Action:         action,
		State:          firstNonEmptyString(status, "failed"),
		Reason:         firstNonEmptyString(reasonCode, message),
		EvidenceSource: "failure_evidence",
	})
	if actionability != nil {
		transcript.addStep(BrowserStepRecoveryTranscriptStep{
			Phase:           BrowserStepRecoveryTranscriptPhaseActionability,
			Action:          strings.TrimSpace(actionability.Action),
			State:           strings.TrimSpace(actionability.Status),
			Reason:          strings.TrimSpace(actionability.FailureReason),
			EvidenceSource:  "actionability",
			FailedCheck:     strings.TrimSpace(actionability.FailedCheck),
			RecoveryAction:  strings.TrimSpace(actionability.RecoveryAction),
			ManualRetryHint: strings.TrimSpace(actionability.ManualRetryHint),
		})
	}
	if outcome != nil {
		transcript.addStep(BrowserStepRecoveryTranscriptStep{
			Phase:           BrowserStepRecoveryTranscriptPhaseTargetResolution,
			Action:          strings.TrimSpace(outcome.PrimaryKind),
			State:           strings.TrimSpace(outcome.Status),
			Reason:          firstNonEmptyString(strings.TrimSpace(outcome.BlockedBy), strings.TrimSpace(outcome.Note)),
			EvidenceSource:  "resolver_outcome",
			RecoveryAction:  strings.TrimSpace(outcome.RecoveryAction),
			NextStepAlias:   strings.TrimSpace(outcome.NextStepAlias),
			ManualRetryHint: strings.TrimSpace(outcome.ManualRetryHint),
		})
	}
	if freshness != nil {
		transcript.addStep(BrowserStepRecoveryTranscriptStep{
			Phase:          BrowserStepRecoveryTranscriptPhaseSnapshotFreshness,
			Action:         strings.TrimSpace(freshness.Source),
			State:          strings.TrimSpace(freshness.State),
			Reason:         strings.TrimSpace(freshness.RefreshReason),
			EvidenceSource: "snapshot_freshness",
			RecoveryAction: strings.TrimSpace(freshness.RecoveryAction),
			NextStepAlias:  strings.TrimSpace(freshness.NextStepAlias),
		})
	}
	if transcript.RecoveryAction != "" || transcript.NextStepAlias != "" || transcript.ManualRetryHint != "" {
		transcript.addStep(BrowserStepRecoveryTranscriptStep{
			Phase:           BrowserStepRecoveryTranscriptPhaseRecommendedNextStep,
			Action:          transcript.RecoveryAction,
			State:           "recommended",
			Reason:          firstNonEmptyString(transcript.NextStepAlias, transcript.ManualRetryHint),
			EvidenceSource:  "recovery_guidance",
			RecoveryAction:  transcript.RecoveryAction,
			NextStepAlias:   transcript.NextStepAlias,
			ManualRetryHint: transcript.ManualRetryHint,
		})
	}
	return normalizeBrowserStepRecoveryTranscript(transcript)
}

func cloneBrowserStepRecoveryTranscript(transcript *BrowserStepRecoveryTranscript) *BrowserStepRecoveryTranscript {
	if transcript == nil {
		return nil
	}
	cloned := *transcript
	if len(transcript.Steps) > 0 {
		cloned.Steps = append([]BrowserStepRecoveryTranscriptStep(nil), transcript.Steps...)
	} else {
		cloned.Steps = nil
	}
	return &cloned
}

func (t *BrowserStepRecoveryTranscript) addStep(step BrowserStepRecoveryTranscriptStep) {
	if t == nil {
		return
	}
	step = normalizeBrowserStepRecoveryTranscriptStep(step)
	if browserStepRecoveryTranscriptStepEmpty(step) {
		return
	}
	step.Index = len(t.Steps) + 1
	t.Steps = append(t.Steps, step)
}

func normalizeBrowserStepRecoveryTranscript(transcript *BrowserStepRecoveryTranscript) *BrowserStepRecoveryTranscript {
	if transcript == nil {
		return nil
	}
	transcript.Kind = firstNonEmptyString(strings.TrimSpace(transcript.Kind), BrowserStepRecoveryTranscriptKind)
	transcript.Action = normalizeBrowserActionabilityAction(transcript.Action)
	transcript.Status = strings.TrimSpace(transcript.Status)
	transcript.ReasonCode = strings.TrimSpace(transcript.ReasonCode)
	transcript.TargetKind = strings.TrimSpace(transcript.TargetKind)
	transcript.Target = strings.TrimSpace(transcript.Target)
	transcript.FailedCheck = strings.TrimSpace(transcript.FailedCheck)
	transcript.SnapshotState = strings.TrimSpace(transcript.SnapshotState)
	transcript.RecoveryAction = strings.TrimSpace(transcript.RecoveryAction)
	transcript.NextStepAlias = strings.TrimSpace(transcript.NextStepAlias)
	transcript.ManualRetryHint = strings.TrimSpace(transcript.ManualRetryHint)
	if len(transcript.Steps) == 0 {
		transcript.Steps = nil
	}
	if transcript.Kind == "" &&
		transcript.Action == "" &&
		transcript.Status == "" &&
		transcript.ReasonCode == "" &&
		transcript.TargetKind == "" &&
		transcript.Target == "" &&
		transcript.FailedCheck == "" &&
		transcript.SnapshotState == "" &&
		transcript.RecoveryAction == "" &&
		transcript.NextStepAlias == "" &&
		transcript.ManualRetryHint == "" &&
		!transcript.Retryable &&
		len(transcript.Steps) == 0 {
		return nil
	}
	return transcript
}

func normalizeBrowserStepRecoveryTranscriptStep(step BrowserStepRecoveryTranscriptStep) BrowserStepRecoveryTranscriptStep {
	step.Phase = strings.TrimSpace(step.Phase)
	step.Action = strings.TrimSpace(step.Action)
	step.State = strings.TrimSpace(step.State)
	step.Reason = strings.TrimSpace(step.Reason)
	step.EvidenceSource = strings.TrimSpace(step.EvidenceSource)
	step.FailedCheck = strings.TrimSpace(step.FailedCheck)
	step.RecoveryAction = strings.TrimSpace(step.RecoveryAction)
	step.NextStepAlias = strings.TrimSpace(step.NextStepAlias)
	step.ManualRetryHint = strings.TrimSpace(step.ManualRetryHint)
	return step
}

func browserStepRecoveryTranscriptStepEmpty(step BrowserStepRecoveryTranscriptStep) bool {
	return strings.TrimSpace(step.Phase) == "" &&
		strings.TrimSpace(step.Action) == "" &&
		strings.TrimSpace(step.State) == "" &&
		strings.TrimSpace(step.Reason) == "" &&
		strings.TrimSpace(step.EvidenceSource) == "" &&
		strings.TrimSpace(step.FailedCheck) == "" &&
		strings.TrimSpace(step.RecoveryAction) == "" &&
		strings.TrimSpace(step.NextStepAlias) == "" &&
		strings.TrimSpace(step.ManualRetryHint) == ""
}

func browserStepRecoverySnapshotState(freshness *BrowserSnapshotFreshness) string {
	if freshness == nil {
		return ""
	}
	return strings.TrimSpace(freshness.State)
}

func browserStepRecoveryRecoveryAction(
	reqRecoveryAction string,
	actionability *BrowserActionabilityReport,
	outcome *BrowserElementResolverOutcome,
	freshness *BrowserSnapshotFreshness,
) string {
	if freshness != nil && strings.TrimSpace(freshness.RecoveryAction) != "" {
		return strings.TrimSpace(freshness.RecoveryAction)
	}
	if actionability != nil && strings.TrimSpace(actionability.RecoveryAction) != "" {
		return strings.TrimSpace(actionability.RecoveryAction)
	}
	if outcome != nil && strings.TrimSpace(outcome.RecoveryAction) != "" {
		return strings.TrimSpace(outcome.RecoveryAction)
	}
	return strings.TrimSpace(reqRecoveryAction)
}

func browserStepRecoveryNextStepAlias(
	actionability *BrowserActionabilityReport,
	outcome *BrowserElementResolverOutcome,
	freshness *BrowserSnapshotFreshness,
	recoveryAction string,
) string {
	if freshness != nil && strings.TrimSpace(freshness.NextStepAlias) != "" {
		return strings.TrimSpace(freshness.NextStepAlias)
	}
	if outcome != nil && strings.TrimSpace(outcome.NextStepAlias) != "" {
		return strings.TrimSpace(outcome.NextStepAlias)
	}
	if alias := browserElementResolverNextStepAlias(strings.TrimSpace(recoveryAction)); alias != "" {
		return alias
	}
	if actionability != nil && strings.TrimSpace(actionability.RecoveryAction) != "" {
		return browserElementResolverNextStepAlias(strings.TrimSpace(actionability.RecoveryAction))
	}
	return ""
}

func browserStepRecoveryManualRetryHint(
	actionability *BrowserActionabilityReport,
	outcome *BrowserElementResolverOutcome,
) string {
	if actionability != nil && strings.TrimSpace(actionability.ManualRetryHint) != "" {
		return strings.TrimSpace(actionability.ManualRetryHint)
	}
	if outcome != nil && strings.TrimSpace(outcome.ManualRetryHint) != "" {
		return strings.TrimSpace(outcome.ManualRetryHint)
	}
	return ""
}
