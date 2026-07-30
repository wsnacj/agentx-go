package telemetry

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"
)

const (
	SemanticRunEventSchemaV1 = "semantic_run_event_v1"

	SemanticRunEventKindRunInterrupted    = "run.interrupted"
	SemanticRunEventKindRunResumed        = "run.resumed"
	SemanticRunEventKindApprovalRequested = "approval.requested"
	SemanticRunEventKindApprovalResolved  = "approval.resolved"

	SemanticRunSourceEventRunStart                 = "run.start"
	SemanticRunSourceEventRunFinish                = "run.finish"
	SemanticRunSourceEventCheckpointUpsert         = "checkpoint.upsert"
	SemanticRunSourceEventHookPermissionRequest    = "hook.permission_request"
	SemanticRunSourceEventToolGuardianReviewStart  = "tool.guardian_review.start"
	SemanticRunSourceEventToolApproval             = "tool.approval"
	SemanticRunSourceEventToolGuardianReview       = "tool.guardian_review"
	SemanticRunSourceEventToolGuardianReviewFinish = "tool.guardian_review.finish"
	SemanticRunSourceEventToolRuntimeDecision      = "tool.runtime_decision"
)

var semanticRunProjectableSourceEvents = []string{
	SemanticRunSourceEventRunStart,
	SemanticRunSourceEventRunFinish,
	SemanticRunSourceEventCheckpointUpsert,
	SemanticRunSourceEventHookPermissionRequest,
	SemanticRunSourceEventToolGuardianReviewStart,
	SemanticRunSourceEventToolApproval,
	SemanticRunSourceEventToolGuardianReview,
	SemanticRunSourceEventToolGuardianReviewFinish,
	SemanticRunSourceEventToolRuntimeDecision,
}

type SemanticRunEvent struct {
	SchemaVersion string                            `json:"schema_version"`
	Timestamp     time.Time                         `json:"timestamp"`
	Kind          string                            `json:"kind"`
	SourceEvent   string                            `json:"source_event"`
	SourceEventID string                            `json:"source_event_id,omitempty"`
	Level         Level                             `json:"level,omitempty"`
	SessionID     string                            `json:"session_id,omitempty"`
	RunID         string                            `json:"run_id,omitempty"`
	BranchID      string                            `json:"branch_id,omitempty"`
	NodeExecID    string                            `json:"node_exec_id,omitempty"`
	Round         int                               `json:"round,omitempty"`
	Status        string                            `json:"status,omitempty"`
	Reason        string                            `json:"reason,omitempty"`
	Stage         string                            `json:"stage,omitempty"`
	Tool          string                            `json:"tool,omitempty"`
	Approval      *SemanticRunApprovalProjection    `json:"approval,omitempty"`
	Checkpoint    *SemanticRunCheckpointProjection  `json:"checkpoint,omitempty"`
	Termination   *SemanticRunTerminationProjection `json:"termination,omitempty"`
}

type SemanticRunApprovalProjection struct {
	Tool               string `json:"tool,omitempty"`
	Action             string `json:"action,omitempty"`
	Decision           string `json:"decision,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Detail             string `json:"detail,omitempty"`
	Stage              string `json:"stage,omitempty"`
	ReviewID           string `json:"review_id,omitempty"`
	Reviewer           string `json:"reviewer,omitempty"`
	Source             string `json:"source,omitempty"`
	RuntimeOwner       string `json:"runtime_owner,omitempty"`
	Risk               string `json:"risk,omitempty"`
	DecisionSubject    string `json:"decision_subject,omitempty"`
	TargetKind         string `json:"target_kind,omitempty"`
	PolicySource       string `json:"policy_source,omitempty"`
	ControlSource      string `json:"control_source,omitempty"`
	EnforcementSurface string `json:"enforcement_surface,omitempty"`
	Checked            bool   `json:"checked,omitempty"`
	Allowed            bool   `json:"allowed,omitempty"`
	Denied             bool   `json:"denied,omitempty"`
	RequiresConfirm    bool   `json:"requires_confirm,omitempty"`
	Degraded           bool   `json:"degraded,omitempty"`
	Timeout            bool   `json:"timeout,omitempty"`
	FailClosed         bool   `json:"fail_closed,omitempty"`
}

type SemanticRunCheckpointProjection struct {
	Stage                string `json:"stage,omitempty"`
	ResumeEnvelopeSchema string `json:"resume_envelope_schema,omitempty"`
	ResumeMode           string `json:"resume_mode,omitempty"`
	InterruptionKind     string `json:"interruption_kind,omitempty"`
	PendingToolCallCount int    `json:"pending_tool_call_count,omitempty"`
	PendingApprovalCount int    `json:"pending_approval_count,omitempty"`
	HasPendingToolCalls  bool   `json:"has_pending_tool_calls,omitempty"`
	HasPendingApprovals  bool   `json:"has_pending_approvals,omitempty"`
	LastErrorPresent     bool   `json:"last_error_present,omitempty"`
	TerminationPresent   bool   `json:"termination_present,omitempty"`
}

type SemanticRunTerminationProjection struct {
	Kind            string `json:"kind,omitempty"`
	CheckpointStage string `json:"checkpoint_stage,omitempty"`
	EventName       string `json:"event_name,omitempty"`
	EventStatus     string `json:"event_status,omitempty"`
	ReplyPersisted  bool   `json:"reply_persisted,omitempty"`
}

type SemanticRunEventJSONLSink struct {
	mu   sync.Mutex
	path string
}

func NewSemanticRunEventJSONLSink(path string) (*SemanticRunEventJSONLSink, error) {
	absPath, err := preparePrivateJSONLPath(path, "semantic run event jsonl")
	if err != nil {
		return nil, err
	}
	return &SemanticRunEventJSONLSink{path: absPath}, nil
}

func (s *SemanticRunEventJSONLSink) Emit(_ context.Context, event Event) error {
	if s == nil {
		return nil
	}
	runEvents := ProjectSemanticRunEvents(event)
	if len(runEvents) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := openPrivateJSONLAppend(s.path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, runEvent := range runEvents {
		payload, err := json.Marshal(normalizeSemanticRunEvent(runEvent))
		if err != nil {
			return err
		}
		if _, err := file.Write(append(payload, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func ProjectSemanticRunEvents(event Event) []SemanticRunEvent {
	event = normalizeEvent(event)
	base := baseSemanticRunEvent(event)
	switch strings.TrimSpace(event.Name) {
	case SemanticRunSourceEventRunStart:
		if !attrBool(event.Attrs, "resume") {
			return nil
		}
		base.Kind = SemanticRunEventKindRunResumed
		base.Reason = "resume_requested"
		return []SemanticRunEvent{base}
	case SemanticRunSourceEventRunFinish:
		termination := semanticRunTerminationProjection(event.Attrs)
		if !semanticRunFinishWasInterrupted(event, termination) {
			return nil
		}
		base.Kind = SemanticRunEventKindRunInterrupted
		base.Reason = semanticRunFinishInterruptionReason(event, termination)
		base.Termination = termination
		if termination != nil && termination.CheckpointStage != "" {
			base.Stage = termination.CheckpointStage
		}
		return []SemanticRunEvent{base}
	case SemanticRunSourceEventCheckpointUpsert:
		checkpoint := semanticRunCheckpointProjection(event.Attrs)
		if checkpoint == nil || !semanticRunCheckpointWasInterrupted(checkpoint) {
			return nil
		}
		base.Kind = SemanticRunEventKindRunInterrupted
		base.Stage = checkpoint.Stage
		base.Reason = semanticRunCheckpointInterruptionReason(checkpoint)
		base.Checkpoint = checkpoint
		return []SemanticRunEvent{base}
	case SemanticRunSourceEventHookPermissionRequest:
		base.Kind = SemanticRunEventKindApprovalRequested
		base.Reason = firstAttrString(event.Attrs, "approval_reason", "approval_detail", "reason")
		base.Approval = semanticRunApprovalProjection(event)
		return []SemanticRunEvent{base}
	case SemanticRunSourceEventToolGuardianReviewStart:
		base.Kind = SemanticRunEventKindApprovalRequested
		base.Reason = firstAttrString(event.Attrs, "rationale", "reason")
		base.Approval = semanticRunApprovalProjection(event)
		return []SemanticRunEvent{base}
	case SemanticRunSourceEventToolApproval, SemanticRunSourceEventToolGuardianReview, SemanticRunSourceEventToolGuardianReviewFinish:
		base.Kind = SemanticRunEventKindApprovalResolved
		base.Reason = firstAttrString(event.Attrs, "reason", "approval_reason", "rationale")
		base.Approval = semanticRunApprovalProjection(event)
		return []SemanticRunEvent{base}
	case SemanticRunSourceEventToolRuntimeDecision:
		approval := semanticRunApprovalProjection(event)
		out := []SemanticRunEvent{}
		if approval != nil && approval.RequiresConfirm {
			requested := base
			requested.Kind = SemanticRunEventKindApprovalRequested
			requested.Reason = firstNonEmptyString(approval.Reason, "runtime_confirmation_required")
			requested.Approval = approval
			out = append(out, requested)
		}
		if semanticRunApprovalDecisionResolved(approval, event.Status) {
			resolved := base
			resolved.Kind = SemanticRunEventKindApprovalResolved
			resolved.Reason = firstNonEmptyString(approval.Reason, strings.ToLower(strings.TrimSpace(event.Status)))
			resolved.Approval = approval
			out = append(out, resolved)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func SemanticRunEventProjectableSourceEvents() []string {
	out := make([]string, len(semanticRunProjectableSourceEvents))
	copy(out, semanticRunProjectableSourceEvents)
	return out
}

func IsSemanticRunEventProjectableSourceEvent(name string) bool {
	normalized := strings.TrimSpace(name)
	for _, sourceEvent := range semanticRunProjectableSourceEvents {
		if normalized == sourceEvent {
			return true
		}
	}
	return false
}

func baseSemanticRunEvent(event Event) SemanticRunEvent {
	attrs := event.Attrs
	round := event.Round
	if round == 0 {
		round = int(attrInt64(attrs, "round"))
	}
	out := SemanticRunEvent{
		SchemaVersion: SemanticRunEventSchemaV1,
		Timestamp:     event.Timestamp,
		SourceEvent:   strings.TrimSpace(event.Name),
		Level:         event.Level,
		SessionID:     strings.TrimSpace(event.SessionID),
		RunID:         firstAttrString(attrs, "run_id"),
		Round:         round,
		Status:        strings.ToLower(strings.TrimSpace(event.Status)),
		Stage:         firstAttrString(attrs, "stage", "termination_checkpoint_stage", "primary_execution_termination_checkpoint_stage"),
		Tool:          strings.ToLower(firstNonEmptyString(event.Tool, firstAttrString(attrs, "tool"))),
	}
	return normalizeSemanticRunEvent(out)
}

func normalizeSemanticRunEvent(event SemanticRunEvent) SemanticRunEvent {
	out := event
	out.SchemaVersion = strings.TrimSpace(out.SchemaVersion)
	if out.SchemaVersion == "" {
		out.SchemaVersion = SemanticRunEventSchemaV1
	}
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now().UTC()
	} else {
		out.Timestamp = out.Timestamp.UTC()
	}
	out.Kind = NormalizeSemanticRunEventKind(out.Kind)
	out.SourceEvent = strings.TrimSpace(out.SourceEvent)
	out.SourceEventID = strings.TrimSpace(out.SourceEventID)
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.RunID = strings.TrimSpace(out.RunID)
	out.BranchID = strings.TrimSpace(out.BranchID)
	out.NodeExecID = strings.TrimSpace(out.NodeExecID)
	out.Status = strings.ToLower(strings.TrimSpace(out.Status))
	out.Reason = strings.TrimSpace(out.Reason)
	out.Stage = strings.TrimSpace(out.Stage)
	out.Tool = strings.ToLower(strings.TrimSpace(out.Tool))
	if out.Level == "" {
		out.Level = LevelInfo
	}
	return out
}

func NormalizeSemanticRunEventKind(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case SemanticRunEventKindRunInterrupted,
		SemanticRunEventKindRunResumed,
		SemanticRunEventKindApprovalRequested,
		SemanticRunEventKindApprovalResolved:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func semanticRunCheckpointProjection(attrs map[string]any) *SemanticRunCheckpointProjection {
	if len(attrs) == 0 {
		return nil
	}
	pendingToolCalls := int(firstAttrInt64(attrs, "pending_tool_call_count", "pending_calls_count"))
	pendingApprovals := int(firstAttrInt64(attrs, "pending_approval_count", "pending_approvals_count"))
	out := &SemanticRunCheckpointProjection{
		Stage:                firstAttrString(attrs, "stage"),
		ResumeEnvelopeSchema: firstAttrString(attrs, "resume_envelope_schema"),
		ResumeMode:           firstAttrString(attrs, "resume_mode"),
		InterruptionKind:     firstAttrString(attrs, "interruption_kind"),
		PendingToolCallCount: pendingToolCalls,
		PendingApprovalCount: pendingApprovals,
		HasPendingToolCalls:  attrBool(attrs, "has_pending_tool_calls") || pendingToolCalls > 0,
		HasPendingApprovals:  attrBool(attrs, "has_pending_approvals") || pendingApprovals > 0,
		LastErrorPresent:     attrBool(attrs, "last_error_present"),
		TerminationPresent:   attrBool(attrs, "termination_present"),
	}
	if out.Stage == "" &&
		out.ResumeEnvelopeSchema == "" &&
		out.ResumeMode == "" &&
		out.InterruptionKind == "" &&
		out.PendingToolCallCount == 0 &&
		out.PendingApprovalCount == 0 &&
		!out.HasPendingToolCalls &&
		!out.HasPendingApprovals &&
		!out.LastErrorPresent &&
		!out.TerminationPresent {
		return nil
	}
	return out
}

func semanticRunCheckpointWasInterrupted(checkpoint *SemanticRunCheckpointProjection) bool {
	if checkpoint == nil {
		return false
	}
	if checkpoint.HasPendingToolCalls || checkpoint.HasPendingApprovals || checkpoint.TerminationPresent {
		return true
	}
	return semanticRunCheckpointStageIsInterrupted(checkpoint.Stage)
}

func semanticRunCheckpointStageIsInterrupted(stage string) bool {
	normalized := strings.ToLower(strings.TrimSpace(stage))
	if normalized == "" {
		return false
	}
	switch normalized {
	case "tool_pending",
		"tool_error",
		"budget_stop",
		"scheduler_queue_stall_break",
		"tool_failure_fuse_break",
		"tool_loop_break",
		"max_rounds_break",
		"run_incomplete",
		"run_interrupted",
		"run_paused":
		return true
	default:
		return strings.HasSuffix(normalized, "_pending") ||
			strings.HasSuffix(normalized, "_break") ||
			strings.HasSuffix(normalized, "_stop")
	}
}

func semanticRunCheckpointInterruptionReason(checkpoint *SemanticRunCheckpointProjection) string {
	if checkpoint == nil {
		return ""
	}
	if checkpoint.HasPendingToolCalls {
		return "pending_tool_calls"
	}
	if checkpoint.HasPendingApprovals {
		return "pending_approvals"
	}
	if checkpoint.TerminationPresent {
		return "checkpoint_termination"
	}
	if strings.TrimSpace(checkpoint.Stage) != "" {
		return "checkpoint_stage=" + strings.TrimSpace(checkpoint.Stage)
	}
	return ""
}

func semanticRunTerminationProjection(attrs map[string]any) *SemanticRunTerminationProjection {
	if len(attrs) == 0 {
		return nil
	}
	out := &SemanticRunTerminationProjection{
		Kind:            firstAttrString(attrs, "termination_kind", "primary_execution_termination_kind"),
		CheckpointStage: firstAttrString(attrs, "termination_checkpoint_stage", "primary_execution_termination_checkpoint_stage"),
		EventName:       firstAttrString(attrs, "termination_event_name", "primary_execution_termination_event_name"),
		EventStatus:     firstAttrString(attrs, "termination_event_status", "primary_execution_termination_event_status"),
		ReplyPersisted:  attrBool(attrs, "termination_reply_persisted") || attrBool(attrs, "primary_execution_termination_reply_persisted"),
	}
	if out.Kind == "" && out.CheckpointStage == "" && out.EventName == "" && out.EventStatus == "" && !out.ReplyPersisted {
		return nil
	}
	return out
}

func semanticRunFinishWasInterrupted(event Event, termination *SemanticRunTerminationProjection) bool {
	finalStatus := strings.ToLower(firstAttrString(event.Attrs, "final_status"))
	if termination != nil && strings.TrimSpace(termination.CheckpointStage) != "" {
		return true
	}
	switch finalStatus {
	case "incomplete", "interrupted", "paused", "pending", "requires_approval":
		return true
	default:
		return false
	}
}

func semanticRunFinishInterruptionReason(event Event, termination *SemanticRunTerminationProjection) string {
	if termination != nil {
		if termination.CheckpointStage != "" {
			return "termination_checkpoint_stage=" + termination.CheckpointStage
		}
		if termination.Kind != "" {
			return "termination_kind=" + termination.Kind
		}
	}
	if finalStatus := firstAttrString(event.Attrs, "final_status"); finalStatus != "" {
		return "final_status=" + finalStatus
	}
	return ""
}

func semanticRunApprovalProjection(event Event) *SemanticRunApprovalProjection {
	attrs := event.Attrs
	if len(attrs) == 0 && strings.TrimSpace(event.Tool) == "" && strings.TrimSpace(event.Status) == "" {
		return nil
	}
	status := strings.ToLower(strings.TrimSpace(event.Status))
	checked := attrBool(attrs, "checked") || attrBool(attrs, "approval_checked")
	allowed := attrBool(attrs, "allowed") || attrBool(attrs, "approval_allowed")
	denied := attrBool(attrs, "denied")
	requiresConfirm := attrBool(attrs, "requires_confirm")
	switch status {
	case "approved", "allowed", "ok":
		checked = true
		allowed = true
	case "denied", "rejected", "blocked":
		checked = true
		denied = true
	case "confirm_required":
		checked = true
		requiresConfirm = true
	}
	out := &SemanticRunApprovalProjection{
		Tool:               strings.ToLower(firstNonEmptyString(event.Tool, firstAttrString(attrs, "tool"))),
		Action:             firstAttrString(attrs, "action"),
		Decision:           semanticRunApprovalDecision(event, allowed, denied, requiresConfirm),
		Reason:             firstAttrString(attrs, "reason", "approval_reason", "rationale"),
		Detail:             firstAttrString(attrs, "detail", "approval_detail"),
		Stage:              firstAttrString(attrs, "stage", "approval_stage"),
		ReviewID:           firstAttrString(attrs, "review_id"),
		Reviewer:           firstAttrString(attrs, "reviewer"),
		Source:             firstAttrString(attrs, "source"),
		RuntimeOwner:       firstAttrString(attrs, "runtime_owner"),
		Risk:               firstAttrString(attrs, "risk"),
		DecisionSubject:    firstAttrString(attrs, "decision_subject"),
		TargetKind:         firstAttrString(attrs, "target_kind"),
		PolicySource:       firstAttrString(attrs, "policy_source"),
		ControlSource:      firstAttrString(attrs, "control_source"),
		EnforcementSurface: firstAttrString(attrs, "enforcement_surface"),
		Checked:            checked,
		Allowed:            allowed,
		Denied:             denied,
		RequiresConfirm:    requiresConfirm,
		Degraded:           attrBool(attrs, "degraded"),
		Timeout:            attrBool(attrs, "timeout"),
		FailClosed:         attrBool(attrs, "fail_closed"),
	}
	if out.Tool == "" &&
		out.Action == "" &&
		out.Decision == "" &&
		out.Reason == "" &&
		out.Stage == "" &&
		out.ReviewID == "" &&
		!out.Checked &&
		!out.Allowed &&
		!out.Denied &&
		!out.RequiresConfirm &&
		!out.Degraded &&
		!out.Timeout &&
		!out.FailClosed {
		return nil
	}
	return out
}

func semanticRunApprovalDecision(event Event, allowed bool, denied bool, requiresConfirm bool) string {
	if outcome := firstAttrString(event.Attrs, "outcome", "decision"); outcome != "" {
		return strings.ToLower(strings.TrimSpace(outcome))
	}
	status := strings.ToLower(strings.TrimSpace(event.Status))
	switch {
	case denied:
		return "denied"
	case requiresConfirm:
		return "confirm_required"
	case allowed:
		return "approved"
	case status != "":
		return status
	default:
		return ""
	}
}

func semanticRunApprovalDecisionResolved(approval *SemanticRunApprovalProjection, status string) bool {
	if approval == nil {
		return false
	}
	if approval.Allowed || approval.Denied || approval.Degraded || approval.Timeout || approval.FailClosed {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "approved", "allowed", "denied", "rejected", "blocked", "degraded":
		return true
	default:
		return false
	}
}

func firstAttrInt64(attrs map[string]any, keys ...string) int64 {
	for _, key := range keys {
		if value := attrInt64(attrs, key); value != 0 {
			return value
		}
	}
	return 0
}
