package telemetry

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestProjectToolEventsMapsFinishToTypedEvents(t *testing.T) {
	events := ProjectToolEvents(Event{
		Component: "tool",
		Name:      "tool.finish",
		Tool:      "open_page",
		Status:    "ok",
		Attrs: map[string]any{
			"duration_ms":                                      42,
			"cached":                                           true,
			"retry_count":                                      2,
			"execution_contract_id":                            "contract-1",
			"execution_contract_diff":                          []any{"visibility.allow_tools"},
			"result_middleware_observed":                       true,
			"result_middleware_mode":                           "observe_only",
			"result_middleware_observation_only":               true,
			"result_middleware_would_transform":                true,
			"result_middleware_reasons":                        "large_result,runtime_guard_truncated",
			"result_middleware_strategies":                     "summarize_with_raw_ref",
			"result_middleware_content_bytes":                  float64(8192),
			"result_middleware_content_lines":                  300,
			"result_middleware_external_content":               true,
			"result_middleware_untrusted_content":              true,
			"result_middleware_output_schema_present":          true,
			"result_middleware_output_schema_closed":           true,
			"result_middleware_output_schema_property_count":   3,
			"result_middleware_output_schema_required":         "query,count",
			"result_middleware_output_schema_missing_required": "count",
			"result_middleware_output_schema_unexpected_keys":  "extra",
			"result_middleware_output_schema_drift":            true,
			"provider_fallback_from":                           "perplexity",
			"provider_fallback_to":                             "brave",
			"provider_fallback_reason":                         "credential_missing",
			"soft_rejection_action":                            "reject_content",
			"soft_rejection_source":                            "tool_output_guard",
			"soft_rejection_surface":                           "open_page",
			"soft_rejection_reason":                            "runtime_guard_truncated",
			"soft_rejection_detail":                            "head_tail",
			"soft_rejection_policy_source":                     "execution_contract",
			"soft_rejection_count":                             2,
			"soft_rejection_actions":                           "reject_content,halt",
			"soft_rejection_sources":                           "tool_output_guard,approval",
			"soft_rejection_reasons":                           "runtime_guard_truncated,policy_deny",
		},
	})
	kinds := make([]string, 0, len(events))
	for _, event := range events {
		kinds = append(kinds, event.Kind)
	}
	wantKinds := []string{
		ToolEventKindRetried,
		ToolEventKindProviderFallback,
		ToolEventKindResultMiddlewareObserved,
		ToolEventKindCompleted,
	}
	if !reflect.DeepEqual(kinds, wantKinds) {
		t.Fatalf("unexpected typed event kinds: got=%#v want=%#v events=%#v", kinds, wantKinds, events)
	}
	if events[0].RetryCount != 2 || !events[0].Cached || events[0].DurationMs != 42 {
		t.Fatalf("expected retry/cache/duration projection, got %#v", events[0])
	}
	if events[1].ProviderFallback == nil ||
		events[1].ProviderFallback.From != "perplexity" ||
		events[1].ProviderFallback.To != "brave" ||
		events[1].ProviderFallback.Reason != "credential_missing" {
		t.Fatalf("expected provider fallback projection, got %#v", events[1])
	}
	if events[2].ResultMiddleware == nil ||
		!events[2].ResultMiddleware.ObservationOnly ||
		!events[2].ResultMiddleware.WouldTransform ||
		events[2].ResultMiddleware.ContentBytes != 8192 ||
		!reflect.DeepEqual(events[2].ResultMiddleware.Reasons, []string{"large_result", "runtime_guard_truncated"}) {
		t.Fatalf("expected result middleware projection, got %#v", events[2])
	}
	if events[2].ResultMiddleware.OutputSchema == nil ||
		!events[2].ResultMiddleware.OutputSchema.Closed ||
		events[2].ResultMiddleware.OutputSchema.PropertyCount != 3 ||
		!events[2].ResultMiddleware.OutputSchema.Drift ||
		!reflect.DeepEqual(events[2].ResultMiddleware.OutputSchema.MissingRequired, []string{"count"}) {
		t.Fatalf("expected output schema projection, got %#v", events[2].ResultMiddleware.OutputSchema)
	}
	if events[3].ExecutionContractID != "contract-1" ||
		!reflect.DeepEqual(events[3].ExecutionContractDiff, []string{"visibility.allow_tools"}) {
		t.Fatalf("expected execution contract projection, got %#v", events[3])
	}
	if events[3].ProviderFallback != nil || events[3].ResultMiddleware != nil {
		t.Fatalf("final completed event should not duplicate specialist projections, got %#v", events[3])
	}
	if events[3].SoftRejection == nil ||
		events[3].SoftRejection.Action != "reject_content" ||
		events[3].SoftRejection.Source != "tool_output_guard" ||
		events[3].SoftRejection.Reason != "runtime_guard_truncated" ||
		events[3].SoftRejection.PolicySource != "execution_contract" ||
		events[3].SoftRejection.Count != 2 ||
		!reflect.DeepEqual(events[3].SoftRejection.Actions, []string{"reject_content", "halt"}) ||
		!reflect.DeepEqual(events[3].SoftRejection.Sources, []string{"tool_output_guard", "approval"}) ||
		!reflect.DeepEqual(events[3].SoftRejection.Reasons, []string{"runtime_guard_truncated", "policy_deny"}) {
		t.Fatalf("expected soft rejection projection, got %#v", events[3].SoftRejection)
	}
}

func TestProjectToolEventsMapsRuntimeDecisionAndRepair(t *testing.T) {
	decisionEvents := ProjectToolEvents(Event{
		Component: "tool",
		Name:      "tool.runtime_decision",
		Tool:      "exec",
		Status:    "allowed",
		Attrs: map[string]any{
			"action":              "exec",
			"checked":             true,
			"allowed":             true,
			"reason":              "exec_allowed",
			"detail":              "runtime exec call passed shared preflight",
			"decision_subject":    "exec_command",
			"target_kind":         "command",
			"policy_source":       "tool_call_policy",
			"control_source":      "execution_contract",
			"enforcement_surface": "runtime",
		},
	})
	if len(decisionEvents) != 1 || decisionEvents[0].Kind != ToolEventKindAuthorized {
		t.Fatalf("expected authorized event, got %#v", decisionEvents)
	}
	if decisionEvents[0].RuntimeDecision == nil ||
		!decisionEvents[0].RuntimeDecision.Checked ||
		!decisionEvents[0].RuntimeDecision.Allowed ||
		decisionEvents[0].RuntimeDecision.Reason != "exec_allowed" ||
		decisionEvents[0].RuntimeDecision.Detail != "runtime exec call passed shared preflight" ||
		decisionEvents[0].RuntimeDecision.DecisionSubject != "exec_command" ||
		decisionEvents[0].RuntimeDecision.TargetKind != "command" ||
		decisionEvents[0].RuntimeDecision.PolicySource != "tool_call_policy" {
		t.Fatalf("expected runtime decision projection, got %#v", decisionEvents[0])
	}

	repairEvents := ProjectToolEvents(Event{
		Component: "tool",
		Name:      "tool.repair",
		Tool:      "browser_click",
		Status:    "applied",
		Attrs: map[string]any{
			"repair_attempted": true,
			"repair_applied":   true,
			"repair_surface":   "browser_click",
			"repair_kinds":     "use_declared_hint",
		},
	})
	if len(repairEvents) != 1 || repairEvents[0].Kind != ToolEventKindArgumentsRepaired {
		t.Fatalf("expected arguments_repaired event, got %#v", repairEvents)
	}
	if repairEvents[0].Repair == nil ||
		!repairEvents[0].Repair.Applied ||
		repairEvents[0].Repair.Surface != "browser_click" ||
		!reflect.DeepEqual(repairEvents[0].Repair.Kinds, []string{"use_declared_hint"}) {
		t.Fatalf("expected repair projection, got %#v", repairEvents[0])
	}
}

func TestToolEventJSONLSinkFiltersAndProjectsToolEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tool-events.jsonl")
	sink, err := NewToolEventJSONLSink(path)
	if err != nil {
		t.Fatalf("new tool event jsonl sink: %v", err)
	}
	if err := sink.Emit(context.Background(), Event{Component: "runner", Name: "run.start"}); err != nil {
		t.Fatalf("emit non-tool event: %v", err)
	}
	if err := sink.Emit(context.Background(), Event{
		Component: "tool",
		Name:      "tool.finish",
		Tool:      "exec",
		Status:    "error",
		Attrs: map[string]any{
			"error_class": "invalid_args",
			"error_code":  "missing_script",
		},
	}); err != nil {
		t.Fatalf("emit tool event: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tool events jsonl: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one projected tool event, got %d: %s", len(lines), string(raw))
	}
	var event ToolEvent
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("decode projected tool event: %v", err)
	}
	if event.SchemaVersion != ToolEventSchemaV1 ||
		event.Kind != ToolEventKindFailed ||
		event.Tool != "exec" ||
		event.ErrorClass != "invalid_args" ||
		event.ErrorCode != "missing_script" {
		t.Fatalf("unexpected projected tool event: %#v", event)
	}
}

func TestReplayToolEventsFromStoredRecordsProjectsTrace(t *testing.T) {
	toolPayload, err := json.Marshal(Event{
		Component: "tool",
		Name:      "tool.finish",
		Tool:      "exec",
		Status:    "error",
		Attrs: map[string]any{
			"retry_count": 1,
			"error_class": "runtime",
			"error_code":  "exec_context_canceled",
		},
	})
	if err != nil {
		t.Fatalf("marshal tool payload: %v", err)
	}
	runPayload, err := json.Marshal(Event{Component: "runner", Name: "run.start"})
	if err != nil {
		t.Fatalf("marshal run payload: %v", err)
	}
	events, trace := ReplayToolEventsFromStoredRecords([]StoredRawEventRecord{
		{
			EventID:     "event-tool",
			RunID:       "run-1",
			Name:        "tool.finish",
			PayloadJSON: string(toolPayload),
			CreatedAt:   1710000000123,
		},
		{
			EventID:     "event-run",
			RunID:       "run-1",
			Name:        "run.start",
			PayloadJSON: string(runPayload),
			CreatedAt:   1710000000456,
		},
		{
			EventID:     "event-bad",
			RunID:       "run-1",
			Name:        "tool.finish",
			PayloadJSON: `{"component":`,
			CreatedAt:   1710000000789,
		},
	})

	if len(events) != 2 {
		t.Fatalf("expected retry + failed projections, got %#v", events)
	}
	for _, event := range events {
		if event.SourceEventID != "event-tool" {
			t.Fatalf("expected projected event to retain source event id, got %#v", event)
		}
		if !event.Timestamp.Equal(time.UnixMilli(1710000000123).UTC()) {
			t.Fatalf("expected created_at fallback timestamp, got %#v", event.Timestamp)
		}
	}
	if trace.SchemaVersion != ToolEventProjectionTraceSchemaV1 ||
		trace.Source != ToolEventProjectionSourceRunstore ||
		trace.SourceEventCount != 3 ||
		trace.DecodedEventCount != 2 ||
		trace.ProjectedEventCount != 2 ||
		trace.InvalidEventCount != 1 ||
		trace.Summary.Failed != 1 ||
		trace.Summary.Retried != 1 {
		t.Fatalf("unexpected projection trace: %#v", trace)
	}
	if len(trace.Sources) != 3 ||
		trace.Sources[0].EventID != "event-tool" ||
		!reflect.DeepEqual(trace.Sources[0].ProjectedKinds, []string{ToolEventKindRetried, ToolEventKindFailed}) ||
		trace.Sources[1].ProjectedCount != 0 ||
		trace.Sources[2].Error != "invalid_payload_json" {
		t.Fatalf("unexpected source trace: %#v", trace.Sources)
	}
}

func TestSummarizeToolEventsCoversHighValueSurfacePilot(t *testing.T) {
	events := []ToolEvent{
		{Kind: ToolEventKindCompleted, Tool: "search"},
		{Kind: ToolEventKindProviderFallback, Tool: "open_page"},
		{Kind: ToolEventKindCompleted, Tool: "browser_click"},
		{
			Kind:   ToolEventKindAuthorized,
			Tool:   "write",
			Reason: "protected_metadata_write_denied",
			RuntimeDecision: &ToolRuntimeDecisionProjection{
				Checked:         true,
				Denied:          true,
				Reason:          "protected_metadata_write_denied",
				DecisionSubject: "protected_metadata",
				PolicySource:    "tool_call_policy",
			},
		},
		{
			Kind: ToolEventKindResultMiddlewareObserved,
			Tool: "pdf_extract",
			ResultMiddleware: &ToolResultMiddlewareProjection{
				OutputSchema: &ToolResultOutputSchemaProjection{Present: true, Drift: true},
			},
		},
		{
			Kind:       ToolEventKindFailed,
			Tool:       "exec",
			ErrorClass: "invalid_args",
			SoftRejection: &ToolSoftRejectionProjection{
				Action: "reject_content",
				Source: "argument_repair",
				Reason: "missing_script",
			},
		},
		{Kind: ToolEventKindRetried, Tool: "exec", RetryCount: 1},
	}
	summary := SummarizeToolEvents(events)
	if summary.TotalEvents != 7 ||
		summary.BySurface[ToolEventSurfaceRetrieval] != 2 ||
		summary.BySurface[ToolEventSurfaceBrowser] != 1 ||
		summary.BySurface[ToolEventSurfacePDF] != 1 ||
		summary.BySurface[ToolEventSurfaceExec] != 2 {
		t.Fatalf("unexpected surface summary: %#v", summary)
	}
	if summary.Failed != 1 ||
		summary.Retried != 1 ||
		summary.ProviderFallbacks != 1 ||
		summary.ResultMiddlewareObserved != 1 {
		t.Fatalf("unexpected event counters: %#v", summary)
	}
	if summary.OutputSchemaObserved != 1 || summary.OutputSchemaDrift != 1 {
		t.Fatalf("unexpected output schema counters: %#v", summary)
	}
	if summary.RuntimeDecisions != 1 ||
		summary.RuntimeDecisionDenied != 1 ||
		summary.ByRuntimeDecisionReason["protected_metadata_write_denied"] != 1 ||
		summary.ByRuntimeDecisionSubject["protected_metadata"] != 1 ||
		summary.ByRuntimeDecisionPolicySource["tool_call_policy"] != 1 {
		t.Fatalf("unexpected runtime decision counters: %#v", summary)
	}
	if summary.SoftRejections != 1 ||
		summary.SoftRejectContent != 1 ||
		summary.SoftHalt != 0 ||
		summary.BySoftRejectionSource["argument_repair"] != 1 ||
		summary.BySoftRejectionReason["missing_script"] != 1 {
		t.Fatalf("unexpected soft rejection counters: %#v", summary)
	}
}
