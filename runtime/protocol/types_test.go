package protocol

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNormalizeRunEventDefaultsSchemaAndTrims(t *testing.T) {
	event := NormalizeRunEvent(RunEvent{
		Envelope: Envelope{
			Kind:  " tool.call.completed ",
			RunID: " run-1 ",
		},
		ToolName: " Browser_Open ",
		Status:   " Completed ",
		RuntimeDecision: &RuntimeDecisionSnapshot{
			Action:        " Allow ",
			TargetKind:    " Tool ",
			Checked:       true,
			Allowed:       true,
			ControlSource: " execution_contract ",
		},
		Attrs: map[string]any{
			" reason ": "ok",
			"":         "dropped",
		},
	})

	if event.SchemaVersion != RunEventSchemaV1 {
		t.Fatalf("unexpected schema version %q", event.SchemaVersion)
	}
	if event.Kind != "tool.call.completed" || event.RunID != "run-1" {
		t.Fatalf("unexpected envelope: %#v", event.Envelope)
	}
	if event.ToolName != "browser_open" || event.Status != "completed" {
		t.Fatalf("expected normalized tool/status, got tool=%q status=%q", event.ToolName, event.Status)
	}
	if event.RuntimeDecision == nil || event.RuntimeDecision.Action != "allow" || event.RuntimeDecision.TargetKind != "tool" {
		t.Fatalf("expected normalized runtime decision, got %#v", event.RuntimeDecision)
	}
	if len(event.Attrs) != 1 || event.Attrs["reason"] != "ok" {
		t.Fatalf("expected trimmed attrs, got %#v", event.Attrs)
	}
	if err := ValidateRunEvent(event); err != nil {
		t.Fatalf("validate run event: %v", err)
	}
}

func TestValidateRunEventRequiresKindAndRunID(t *testing.T) {
	if err := ValidateRunEvent(RunEvent{}); err == nil || !strings.Contains(err.Error(), "kind is required") {
		t.Fatalf("expected missing kind error, got %v", err)
	}
	if err := ValidateRunEvent(RunEvent{Envelope: Envelope{Kind: "run.started"}}); err == nil || !strings.Contains(err.Error(), "run_id is required") {
		t.Fatalf("expected missing run_id error, got %v", err)
	}
}

func TestTraceSpanReservedSchemaNormalizesValidatesAndMarshals(t *testing.T) {
	span := NormalizeTraceSpan(TraceSpan{
		Envelope: Envelope{
			Kind:         " Trace.Span ",
			RunID:        " run-1 ",
			TraceID:      " trace-1 ",
			SpanID:       " span-7 ",
			ParentSpanID: " span-3 ",
		},
		Type:   " Tool ",
		Name:   " Browser_Open ",
		Status: " OK ",
		Usage: &Usage{
			ModelID: " gpt-test ",
		},
		Attrs: map[string]any{
			" tool_call_id ": "call-7",
			"":               "dropped",
		},
	})

	if span.SchemaVersion != TraceSpanSchemaV1 {
		t.Fatalf("unexpected schema version %q", span.SchemaVersion)
	}
	if span.Kind != "trace.span" || span.RunID != "run-1" || span.TraceID != "trace-1" || span.SpanID != "span-7" {
		t.Fatalf("unexpected normalized envelope: %#v", span.Envelope)
	}
	if span.Type != "tool" || span.Name != "Browser_Open" || span.Status != "ok" {
		t.Fatalf("unexpected normalized span fields: %#v", span)
	}
	if len(span.Attrs) != 1 || span.Attrs["tool_call_id"] != "call-7" {
		t.Fatalf("expected trimmed attrs, got %#v", span.Attrs)
	}
	if err := ValidateTraceSpan(span); err != nil {
		t.Fatalf("validate trace span: %v", err)
	}

	payload, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("marshal trace span: %v", err)
	}
	text := string(payload)
	if !strings.Contains(text, `"schema_version":"agentx.trace_span.v1"`) ||
		!strings.Contains(text, `"trace_id":"trace-1"`) ||
		!strings.Contains(text, `"span_id":"span-7"`) {
		t.Fatalf("expected trace span identity in payload: %s", text)
	}
	if strings.Contains(text, `"arguments":`) {
		t.Fatalf("trace span schema must not expose raw arguments: %s", text)
	}
}

func TestValidateTraceSpanRequiresReservedIdentity(t *testing.T) {
	base := TraceSpan{
		Envelope: Envelope{
			Kind:    "trace.span",
			RunID:   "run-1",
			TraceID: "trace-1",
			SpanID:  "span-1",
		},
		Type: "tool",
	}

	cases := []struct {
		name   string
		mutate func(*TraceSpan)
		want   string
	}{
		{
			name: "kind",
			mutate: func(span *TraceSpan) {
				span.Kind = ""
			},
			want: "kind is required",
		},
		{
			name: "run_id",
			mutate: func(span *TraceSpan) {
				span.RunID = ""
			},
			want: "run_id is required",
		},
		{
			name: "trace_id",
			mutate: func(span *TraceSpan) {
				span.TraceID = ""
			},
			want: "trace_id is required",
		},
		{
			name: "span_id",
			mutate: func(span *TraceSpan) {
				span.SpanID = ""
			},
			want: "span_id is required",
		},
		{
			name: "type",
			mutate: func(span *TraceSpan) {
				span.Type = ""
			},
			want: "type is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			span := base
			tc.mutate(&span)
			if err := ValidateTraceSpan(span); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q error, got %v", tc.want, err)
			}
		})
	}
}

func TestValidateToolExecutionPlanAndJSONShape(t *testing.T) {
	plan := NormalizeToolExecutionPlan(ToolExecutionPlan{
		Envelope:              Envelope{RunID: " run-1 "},
		PlanID:                " plan-1 ",
		ExecutionContractID:   " contract-1 ",
		ExecutionContractDiff: []string{" tools.max_concurrency "},
		Calls: []ToolPlanCall{
			{
				ToolCallID:    " call-1 ",
				ToolName:      " Browser_Open ",
				ArgumentsHash: " sha256:abc ",
				Status:        StatusPlanned,
			},
		},
	})
	if plan.SchemaVersion != ToolExecutionPlanSchemaV1 || plan.Kind != KindToolPlanCreated {
		t.Fatalf("unexpected plan envelope: %#v", plan.Envelope)
	}
	if err := ValidateToolExecutionPlan(plan); err != nil {
		t.Fatalf("validate plan: %v", err)
	}

	payload, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, `"arguments":`) {
		t.Fatalf("tool execution plan must not expose raw arguments: %s", text)
	}
	if !strings.Contains(text, `"arguments_hash":"sha256:abc"`) {
		t.Fatalf("expected arguments hash in payload: %s", text)
	}
	if !strings.Contains(text, `"execution_contract_id":"contract-1"`) ||
		!strings.Contains(text, `"execution_contract_diff":["tools.max_concurrency"]`) {
		t.Fatalf("expected execution contract projection in payload: %s", text)
	}
}

func TestValidateHandoffRequiresSourceAndTarget(t *testing.T) {
	record := HandoffRecord{
		Envelope:    Envelope{Kind: KindHandoffAccepted, RunID: "run-parent"},
		HandoffID:   "handoff-1",
		HandoffKind: HandoffKindTaskChild,
		Source:      HandoffEndpoint{AgentID: "root"},
		Target:      HandoffEndpoint{TaskID: " task-child ", SessionID: " session-child "},
		Isolation:   HandoffIsolation{Scope: "task:call-1"},
	}
	if err := ValidateHandoffRecord(record); err != nil {
		t.Fatalf("validate handoff: %v", err)
	}
	normalized := NormalizeHandoffRecord(record)
	if normalized.HandoffKind != HandoffKindTaskChild ||
		normalized.Target.TaskID != "task-child" ||
		normalized.Target.SessionID != "session-child" {
		t.Fatalf("expected normalized handoff kind/task endpoint, got %#v", normalized)
	}
	if err := ValidateHandoffRecord(HandoffRecord{
		Envelope:  Envelope{Kind: KindHandoffAccepted, RunID: "run-parent"},
		HandoffID: "handoff-1",
	}); err == nil || !strings.Contains(err.Error(), "source is required") {
		t.Fatalf("expected missing source error, got %v", err)
	}
}

func TestValidateSandboxManifestRejectsEscapingEntryPath(t *testing.T) {
	err := ValidateSandboxManifest(SandboxManifest{
		Envelope:   Envelope{RunID: "run-1"},
		ManifestID: "sandbox-1",
		Root:       "/workspace",
		Entries: []SandboxEntry{
			{Path: "../secret", Kind: "mount"},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "must not escape root") {
		t.Fatalf("expected escape-root error, got %v", err)
	}
}

func TestArtifactVersionAndLinkValidation(t *testing.T) {
	version := NormalizeArtifactVersion(ArtifactVersion{
		Envelope:   Envelope{Kind: " PDF_PARSE_RESULT ", RunID: " run-1 "},
		ArtifactID: " artifact-1 ",
		Version:    1,
		Scope:      " Session ",
		MIMEType:   " Application/JSON ",
		Payload:    ArtifactPayload{Storage: " Blob ", BlobRef: " sha256:abc "},
	})
	if version.SchemaVersion != ArtifactVersionSchemaV1 || version.Kind != "pdf_parse_result" {
		t.Fatalf("unexpected artifact envelope: %#v", version.Envelope)
	}
	if version.Scope != "session" || version.MIMEType != "application/json" || version.Payload.Storage != "blob" {
		t.Fatalf("expected normalized artifact fields, got %#v", version)
	}
	if err := ValidateArtifactVersion(version); err != nil {
		t.Fatalf("validate artifact version: %v", err)
	}

	link := NormalizeArtifactLink(ArtifactLink{
		SourceArtifactID: " artifact-1 ",
		TargetArtifactID: " artifact-2 ",
		Relation:         " Derived_From ",
	})
	if link.SchemaVersion != ArtifactLinkSchemaV1 || link.Relation != "derived_from" {
		t.Fatalf("unexpected link normalization: %#v", link)
	}
	if err := ValidateArtifactLink(link); err != nil {
		t.Fatalf("validate artifact link: %v", err)
	}
}
