package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ToolEventSchemaV1 = "tool_event_v1"

	ToolEventKindStarted                  = "started"
	ToolEventKindArgumentsRepaired        = "arguments_repaired"
	ToolEventKindAuthorized               = "authorized"
	ToolEventKindExecuting                = "executing"
	ToolEventKindCompleted                = "completed"
	ToolEventKindFailed                   = "failed"
	ToolEventKindRetried                  = "retried"
	ToolEventKindResultMiddlewareObserved = "result_middleware_observed"
	ToolEventKindResultMiddlewareApplied  = "result_middleware_applied"
	ToolEventKindProviderFallback         = "provider_fallback"
)

type ToolEvent struct {
	SchemaVersion         string                          `json:"schema_version"`
	Timestamp             time.Time                       `json:"timestamp"`
	Kind                  string                          `json:"kind"`
	SourceEvent           string                          `json:"source_event"`
	SourceEventID         string                          `json:"source_event_id,omitempty"`
	Level                 Level                           `json:"level,omitempty"`
	SessionID             string                          `json:"session_id,omitempty"`
	RunID                 string                          `json:"run_id,omitempty"`
	Round                 int                             `json:"round,omitempty"`
	Tool                  string                          `json:"tool,omitempty"`
	Status                string                          `json:"status,omitempty"`
	DurationMs            int64                           `json:"duration_ms,omitempty"`
	Cached                bool                            `json:"cached,omitempty"`
	RetryCount            int                             `json:"retry_count,omitempty"`
	ErrorClass            string                          `json:"error_class,omitempty"`
	ErrorCode             string                          `json:"error_code,omitempty"`
	Reason                string                          `json:"reason,omitempty"`
	ExecutionContractID   string                          `json:"execution_contract_id,omitempty"`
	ExecutionContractDiff []string                        `json:"execution_contract_diff,omitempty"`
	RuntimeDecision       *ToolRuntimeDecisionProjection  `json:"runtime_decision,omitempty"`
	Repair                *ToolRepairProjection           `json:"repair,omitempty"`
	SoftRejection         *ToolSoftRejectionProjection    `json:"soft_rejection,omitempty"`
	ResultMiddleware      *ToolResultMiddlewareProjection `json:"result_middleware,omitempty"`
	ProviderFallback      *ToolProviderFallbackProjection `json:"provider_fallback,omitempty"`
}

type ToolRuntimeDecisionProjection struct {
	Action             string `json:"action,omitempty"`
	Reason             string `json:"reason,omitempty"`
	Detail             string `json:"detail,omitempty"`
	DecisionSubject    string `json:"decision_subject,omitempty"`
	TargetKind         string `json:"target_kind,omitempty"`
	Checked            bool   `json:"checked,omitempty"`
	Allowed            bool   `json:"allowed,omitempty"`
	Denied             bool   `json:"denied,omitempty"`
	RequiresConfirm    bool   `json:"requires_confirm,omitempty"`
	Degraded           bool   `json:"degraded,omitempty"`
	PolicySource       string `json:"policy_source,omitempty"`
	ControlSource      string `json:"control_source,omitempty"`
	EnforcementSurface string `json:"enforcement_surface,omitempty"`
}

type ToolRepairProjection struct {
	Surface string   `json:"surface,omitempty"`
	Kinds   []string `json:"kinds,omitempty"`
	Applied bool     `json:"applied,omitempty"`
}

type ToolSoftRejectionProjection struct {
	Action       string   `json:"action,omitempty"`
	Source       string   `json:"source,omitempty"`
	Surface      string   `json:"surface,omitempty"`
	Reason       string   `json:"reason,omitempty"`
	Detail       string   `json:"detail,omitempty"`
	PolicySource string   `json:"policy_source,omitempty"`
	Count        int      `json:"count,omitempty"`
	Actions      []string `json:"actions,omitempty"`
	Sources      []string `json:"sources,omitempty"`
	Reasons      []string `json:"reasons,omitempty"`
}

type ToolResultMiddlewareProjection struct {
	Mode             string                            `json:"mode,omitempty"`
	ObservationOnly  bool                              `json:"observation_only,omitempty"`
	WouldTransform   bool                              `json:"would_transform,omitempty"`
	Reasons          []string                          `json:"reasons,omitempty"`
	Strategies       []string                          `json:"strategies,omitempty"`
	ContentBytes     int64                             `json:"content_bytes,omitempty"`
	ContentLines     int64                             `json:"content_lines,omitempty"`
	ExternalContent  bool                              `json:"external_content,omitempty"`
	UntrustedContent bool                              `json:"untrusted_content,omitempty"`
	ErrorPreserved   bool                              `json:"error_preserved,omitempty"`
	OutputSchema     *ToolResultOutputSchemaProjection `json:"output_schema,omitempty"`
}

type ToolResultOutputSchemaProjection struct {
	Present                bool     `json:"present"`
	Closed                 bool     `json:"closed"`
	PropertyCount          int64    `json:"property_count,omitempty"`
	Required               []string `json:"required,omitempty"`
	MatchedTopLevelKeys    []string `json:"matched_top_level_keys,omitempty"`
	MissingRequired        []string `json:"missing_required,omitempty"`
	UnexpectedTopLevelKeys []string `json:"unexpected_top_level_keys,omitempty"`
	Drift                  bool     `json:"drift"`
}

type ToolProviderFallbackProjection struct {
	From   string `json:"from,omitempty"`
	To     string `json:"to,omitempty"`
	Reason string `json:"reason,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

type ToolEventJSONLSink struct {
	mu   sync.Mutex
	path string
}

func NewToolEventJSONLSink(path string) (*ToolEventJSONLSink, error) {
	absPath, err := preparePrivateJSONLPath(path, "tool event jsonl")
	if err != nil {
		return nil, err
	}
	return &ToolEventJSONLSink{path: absPath}, nil
}

func (s *ToolEventJSONLSink) Emit(_ context.Context, event Event) error {
	if s == nil {
		return nil
	}
	toolEvents := ProjectToolEvents(event)
	if len(toolEvents) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := openPrivateJSONLAppend(s.path)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, toolEvent := range toolEvents {
		payload, err := json.Marshal(normalizeToolEvent(toolEvent))
		if err != nil {
			return err
		}
		if _, err := file.Write(append(payload, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func ProjectToolEvents(event Event) []ToolEvent {
	event = normalizeEvent(event)
	if strings.TrimSpace(event.Component) != "tool" && !strings.HasPrefix(strings.TrimSpace(event.Name), "tool.") {
		return nil
	}
	base := baseToolEvent(event)
	switch strings.TrimSpace(event.Name) {
	case "tool.start":
		base.Kind = ToolEventKindStarted
		return []ToolEvent{base}
	case "tool.approval", "tool.runtime_decision", "tool.guardian_review.start", "tool.guardian_review", "tool.guardian_review.finish":
		base.Kind = ToolEventKindAuthorized
		return []ToolEvent{base}
	case "tool.repair":
		base.Kind = ToolEventKindArgumentsRepaired
		return []ToolEvent{base}
	case "tool.finish":
		return projectToolFinishEvents(base, event)
	default:
		return nil
	}
}

func projectToolFinishEvents(base ToolEvent, event Event) []ToolEvent {
	out := []ToolEvent{}
	if base.RetryCount > 0 {
		retried := base
		retried.Kind = ToolEventKindRetried
		out = append(out, retried)
	}
	if fallback := toolProviderFallbackProjection(event.Attrs); fallback != nil {
		provider := base
		provider.Kind = ToolEventKindProviderFallback
		provider.ProviderFallback = fallback
		out = append(out, provider)
	}
	if middleware := toolResultMiddlewareProjection(event.Attrs); middleware != nil {
		middlewareEvent := base
		if attrBool(event.Attrs, "result_middleware_applied") {
			middlewareEvent.Kind = ToolEventKindResultMiddlewareApplied
		} else {
			middlewareEvent.Kind = ToolEventKindResultMiddlewareObserved
		}
		middlewareEvent.ResultMiddleware = middleware
		out = append(out, middlewareEvent)
	}
	final := base
	if strings.EqualFold(strings.TrimSpace(event.Status), "error") || strings.TrimSpace(base.ErrorClass) != "" {
		final.Kind = ToolEventKindFailed
	} else {
		final.Kind = ToolEventKindCompleted
	}
	out = append(out, final)
	return out
}

func baseToolEvent(event Event) ToolEvent {
	attrs := event.Attrs
	out := ToolEvent{
		SchemaVersion:         ToolEventSchemaV1,
		Timestamp:             event.Timestamp,
		SourceEvent:           strings.TrimSpace(event.Name),
		Level:                 event.Level,
		SessionID:             strings.TrimSpace(event.SessionID),
		RunID:                 firstAttrString(attrs, "run_id", "result_middleware_run_id"),
		Round:                 event.Round,
		Tool:                  strings.ToLower(strings.TrimSpace(event.Tool)),
		Status:                strings.ToLower(strings.TrimSpace(event.Status)),
		DurationMs:            attrInt64(attrs, "duration_ms"),
		Cached:                attrBool(attrs, "cached"),
		RetryCount:            int(attrInt64(attrs, "retry_count")),
		ErrorClass:            firstAttrString(attrs, "error_class"),
		ErrorCode:             firstAttrString(attrs, "error_code", "repair_error_code"),
		Reason:                firstAttrString(attrs, "reason", "approval_reason"),
		ExecutionContractID:   firstAttrString(attrs, "execution_contract_id"),
		ExecutionContractDiff: attrStringList(attrs, "execution_contract_diff"),
		RuntimeDecision:       toolRuntimeDecisionProjection(attrs),
		Repair:                toolRepairProjection(attrs),
		SoftRejection:         toolSoftRejectionProjection(attrs),
	}
	return normalizeToolEvent(out)
}

func normalizeToolEvent(event ToolEvent) ToolEvent {
	out := event
	out.SchemaVersion = strings.TrimSpace(out.SchemaVersion)
	if out.SchemaVersion == "" {
		out.SchemaVersion = ToolEventSchemaV1
	}
	if out.Timestamp.IsZero() {
		out.Timestamp = time.Now().UTC()
	} else {
		out.Timestamp = out.Timestamp.UTC()
	}
	out.Kind = NormalizeToolEventKind(out.Kind)
	out.SourceEvent = strings.TrimSpace(out.SourceEvent)
	out.SourceEventID = strings.TrimSpace(out.SourceEventID)
	out.SessionID = strings.TrimSpace(out.SessionID)
	out.RunID = strings.TrimSpace(out.RunID)
	out.Tool = strings.ToLower(strings.TrimSpace(out.Tool))
	out.Status = strings.ToLower(strings.TrimSpace(out.Status))
	out.ErrorClass = strings.TrimSpace(out.ErrorClass)
	out.ErrorCode = strings.TrimSpace(out.ErrorCode)
	out.Reason = strings.TrimSpace(out.Reason)
	out.ExecutionContractID = strings.TrimSpace(out.ExecutionContractID)
	out.ExecutionContractDiff = normalizeStringList(out.ExecutionContractDiff)
	out.SoftRejection = normalizeToolSoftRejectionProjection(out.SoftRejection)
	if out.Level == "" {
		out.Level = LevelInfo
	}
	return out
}

func NormalizeToolEventKind(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case ToolEventKindStarted,
		ToolEventKindArgumentsRepaired,
		ToolEventKindAuthorized,
		ToolEventKindExecuting,
		ToolEventKindCompleted,
		ToolEventKindFailed,
		ToolEventKindRetried,
		ToolEventKindResultMiddlewareObserved,
		ToolEventKindResultMiddlewareApplied,
		ToolEventKindProviderFallback:
		return strings.ToLower(strings.TrimSpace(raw))
	default:
		return ""
	}
}

func toolRuntimeDecisionProjection(attrs map[string]any) *ToolRuntimeDecisionProjection {
	if len(attrs) == 0 {
		return nil
	}
	action := firstAttrString(attrs, "action")
	checked := attrBool(attrs, "checked")
	allowed := attrBool(attrs, "allowed")
	denied := attrBool(attrs, "denied")
	requiresConfirm := attrBool(attrs, "requires_confirm")
	degraded := attrBool(attrs, "degraded")
	if action == "" && !checked && !allowed && !denied && !requiresConfirm && !degraded {
		return nil
	}
	return &ToolRuntimeDecisionProjection{
		Action:             action,
		Reason:             firstAttrString(attrs, "reason"),
		Detail:             firstAttrString(attrs, "detail"),
		DecisionSubject:    firstAttrString(attrs, "decision_subject"),
		TargetKind:         firstAttrString(attrs, "target_kind"),
		Checked:            checked,
		Allowed:            allowed,
		Denied:             denied,
		RequiresConfirm:    requiresConfirm,
		Degraded:           degraded,
		PolicySource:       firstAttrString(attrs, "policy_source"),
		ControlSource:      firstAttrString(attrs, "control_source"),
		EnforcementSurface: firstAttrString(attrs, "enforcement_surface"),
	}
}

func toolRepairProjection(attrs map[string]any) *ToolRepairProjection {
	if len(attrs) == 0 {
		return nil
	}
	attempted := attrBool(attrs, "repair_attempted")
	applied := attrBool(attrs, "repair_applied")
	surface := firstAttrString(attrs, "repair_surface")
	kinds := attrCSVOrStringList(attrs, "repair_kinds")
	if !attempted && !applied && surface == "" && len(kinds) == 0 {
		return nil
	}
	return &ToolRepairProjection{
		Surface: surface,
		Kinds:   kinds,
		Applied: applied,
	}
}

func toolSoftRejectionProjection(attrs map[string]any) *ToolSoftRejectionProjection {
	if len(attrs) == 0 {
		return nil
	}
	action := firstAttrString(attrs, "soft_rejection_action")
	source := firstAttrString(attrs, "soft_rejection_source")
	reason := firstAttrString(attrs, "soft_rejection_reason")
	if action == "" && source == "" && reason == "" {
		return nil
	}
	count := int(attrInt64(attrs, "soft_rejection_count"))
	if count <= 0 {
		count = 1
	}
	return normalizeToolSoftRejectionProjection(&ToolSoftRejectionProjection{
		Action:       action,
		Source:       source,
		Surface:      firstAttrString(attrs, "soft_rejection_surface"),
		Reason:       reason,
		Detail:       firstAttrString(attrs, "soft_rejection_detail"),
		PolicySource: firstAttrString(attrs, "soft_rejection_policy_source"),
		Count:        count,
		Actions:      attrCSVOrStringList(attrs, "soft_rejection_actions"),
		Sources:      attrCSVOrStringList(attrs, "soft_rejection_sources"),
		Reasons:      attrCSVOrStringList(attrs, "soft_rejection_reasons"),
	})
}

func normalizeToolSoftRejectionProjection(in *ToolSoftRejectionProjection) *ToolSoftRejectionProjection {
	if in == nil {
		return nil
	}
	out := *in
	out.Action = strings.ToLower(strings.TrimSpace(out.Action))
	out.Source = strings.ToLower(strings.TrimSpace(out.Source))
	out.Surface = strings.TrimSpace(out.Surface)
	out.Reason = strings.TrimSpace(out.Reason)
	out.Detail = strings.TrimSpace(out.Detail)
	out.PolicySource = strings.TrimSpace(out.PolicySource)
	out.Actions = normalizeStringList(out.Actions)
	out.Sources = normalizeStringList(out.Sources)
	out.Reasons = normalizeStringList(out.Reasons)
	if out.Count <= 0 {
		out.Count = 1
	}
	if out.Action == "" && out.Source == "" && out.Reason == "" &&
		len(out.Actions) == 0 && len(out.Sources) == 0 && len(out.Reasons) == 0 {
		return nil
	}
	return &out
}

func toolResultMiddlewareProjection(attrs map[string]any) *ToolResultMiddlewareProjection {
	if len(attrs) == 0 || !attrBool(attrs, "result_middleware_observed") {
		return nil
	}
	return &ToolResultMiddlewareProjection{
		Mode:             firstAttrString(attrs, "result_middleware_mode"),
		ObservationOnly:  attrBool(attrs, "result_middleware_observation_only"),
		WouldTransform:   attrBool(attrs, "result_middleware_would_transform"),
		Reasons:          attrCSVOrStringList(attrs, "result_middleware_reasons"),
		Strategies:       attrCSVOrStringList(attrs, "result_middleware_strategies"),
		ContentBytes:     attrInt64(attrs, "result_middleware_content_bytes"),
		ContentLines:     attrInt64(attrs, "result_middleware_content_lines"),
		ExternalContent:  attrBool(attrs, "result_middleware_external_content"),
		UntrustedContent: attrBool(attrs, "result_middleware_untrusted_content"),
		ErrorPreserved:   attrBool(attrs, "result_middleware_error_preserved"),
		OutputSchema:     toolResultOutputSchemaProjection(attrs),
	}
}

func toolResultOutputSchemaProjection(attrs map[string]any) *ToolResultOutputSchemaProjection {
	if len(attrs) == 0 || !attrBool(attrs, "result_middleware_output_schema_present") {
		return nil
	}
	return &ToolResultOutputSchemaProjection{
		Present:                true,
		Closed:                 attrBool(attrs, "result_middleware_output_schema_closed"),
		PropertyCount:          attrInt64(attrs, "result_middleware_output_schema_property_count"),
		Required:               attrCSVOrStringList(attrs, "result_middleware_output_schema_required"),
		MatchedTopLevelKeys:    attrCSVOrStringList(attrs, "result_middleware_output_schema_matched_keys"),
		MissingRequired:        attrCSVOrStringList(attrs, "result_middleware_output_schema_missing_required"),
		UnexpectedTopLevelKeys: attrCSVOrStringList(attrs, "result_middleware_output_schema_unexpected_keys"),
		Drift:                  attrBool(attrs, "result_middleware_output_schema_drift"),
	}
}

func toolProviderFallbackProjection(attrs map[string]any) *ToolProviderFallbackProjection {
	if len(attrs) == 0 {
		return nil
	}
	from := firstAttrString(attrs, "provider_fallback_from", "provider_from")
	to := firstAttrString(attrs, "provider_fallback_to", "provider_to")
	reason := firstAttrString(attrs, "provider_fallback_reason", "fallback_reason")
	hint := firstAttrString(attrs, "provider_fallback_hint", "fallback_hint")
	if from == "" && to == "" && reason == "" && hint == "" {
		return nil
	}
	return &ToolProviderFallbackProjection{
		From:   from,
		To:     to,
		Reason: reason,
		Hint:   hint,
	}
}

func firstAttrString(attrs map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := attrString(attrs, key); value != "" {
			return value
		}
	}
	return ""
}

func attrString(attrs map[string]any, key string) string {
	if len(attrs) == 0 {
		return ""
	}
	value, ok := attrs[key]
	if !ok {
		return ""
	}
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func attrBool(attrs map[string]any, key string) bool {
	if len(attrs) == 0 {
		return false
	}
	value, ok := attrs[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed
	default:
		return false
	}
}

func attrInt64(attrs map[string]any, key string) int64 {
	if len(attrs) == 0 {
		return 0
	}
	value, ok := attrs[key]
	if !ok {
		return 0
	}
	switch typed := value.(type) {
	case int:
		return int64(typed)
	case int8:
		return int64(typed)
	case int16:
		return int64(typed)
	case int32:
		return int64(typed)
	case int64:
		return typed
	case uint:
		return int64(typed)
	case uint8:
		return int64(typed)
	case uint16:
		return int64(typed)
	case uint32:
		return int64(typed)
	case uint64:
		return int64(typed)
	case float32:
		return int64(typed)
	case float64:
		return int64(typed)
	case json.Number:
		n, _ := typed.Int64()
		return n
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return n
	default:
		return 0
	}
}

func attrCSVOrStringList(attrs map[string]any, key string) []string {
	value := attrStringList(attrs, key)
	if len(value) > 0 {
		return value
	}
	raw := attrString(attrs, key)
	if raw == "" {
		return nil
	}
	return normalizeStringList(strings.Split(raw, ","))
}

func attrStringList(attrs map[string]any, key string) []string {
	if len(attrs) == 0 {
		return nil
	}
	value, ok := attrs[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return normalizeStringList(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, strings.TrimSpace(fmt.Sprint(item)))
		}
		return normalizeStringList(out)
	default:
		return nil
	}
}

func normalizeStringList(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
