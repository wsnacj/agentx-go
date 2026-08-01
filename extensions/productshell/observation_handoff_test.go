package productshell

import (
	"reflect"
	"strings"
	"testing"
)

func TestBuildHostProcessProgressObservationNormalizesAndClones(t *testing.T) {
	missing := []string{" missing ", ""}
	boundaries := []string{" no_tool_output_process_status_parsing "}
	got := BuildHostProcessProgressObservation(HostProcessProgressObservationInput{
		Status:                         " ready ",
		DisplayLine:                    " kind=progress ",
		ProcessRef:                     " host-process:demo ",
		ProcessCount:                   -1,
		ActiveCount:                    -2,
		ConsumesHostProcessSessionView: true,
		MissingInputs:                  missing,
		Boundaries:                     boundaries,
	})
	if got == nil {
		t.Fatal("expected observation")
	}
	missing[0] = "mutated"
	boundaries[0] = "mutated"
	if got.Source != "productshellruntime_host_process_progress_display" || got.Status != "ready" || got.ProcessRef != "host-process:demo" {
		t.Fatalf("unexpected normalization: %#v", got)
	}
	if got.ProcessCount != 0 || got.ActiveCount != 0 || !got.ConsumesHostProcessSessionView {
		t.Fatalf("unexpected typed values: %#v", got)
	}
	if !reflect.DeepEqual(got.MissingInputs, []string{"missing"}) || !reflect.DeepEqual(got.Boundaries, []string{"no_tool_output_process_status_parsing"}) {
		t.Fatalf("input slices were not normalized and cloned: %#v", got)
	}
}

func TestBuildHostDiagnosticOperatorLineObservationEmptyAndClone(t *testing.T) {
	if got := BuildHostDiagnosticOperatorLineObservation(HostDiagnosticOperatorLineObservationInput{}); got != nil {
		t.Fatalf("empty input = %#v, want nil", got)
	}
	blocked := []string{" blocked "}
	got := BuildHostDiagnosticOperatorLineObservation(HostDiagnosticOperatorLineObservationInput{
		Key:                 " diagnostic ",
		OperatorDisplayLine: " kind=ready ",
		BlockedReasons:      blocked,
	})
	blocked[0] = "mutated"
	if got == nil || got.Source != "host_diagnostic_operator_line" || got.Key != "diagnostic" || got.OperatorDisplayLine != "kind=ready" {
		t.Fatalf("unexpected observation: %#v", got)
	}
	if !reflect.DeepEqual(got.BlockedReasons, []string{"blocked"}) {
		t.Fatalf("blocked reasons alias caller memory: %#v", got.BlockedReasons)
	}
}

func TestBuildSessionObservationPreservesOrderAndTypedSummary(t *testing.T) {
	got := BuildSessionObservation(SessionObservationInput{
		SessionID: " session ",
		Events: []SessionEventObservationInput{
			{Role: "user", Content: "  run   tests  "},
			{Role: "assistant", Content: "running", ToolCallCount: 1},
			{Role: "tool", Content: "ok", ToolCallID: "call-1"},
		},
		Branches: []SessionBranchObservationInput{
			{BranchID: " branch ", NodeExecID: "node-1", Status: "running", StartedAt: 10},
			{BranchID: "branch", NodeExecID: "node-2", Status: "completed", FinishedAt: 20},
		},
		Compaction: SessionCompactionObservationInput{Passes: 1, CompactedToolOutputs: 1},
	})
	if got == nil || got.SessionID != "session" || got.EventCount != 3 || got.ToolCallMessageCount != 1 || got.ToolResultCount != 1 {
		t.Fatalf("unexpected session observation: %#v", got)
	}
	if got.LatestUserPreview != "run tests" || got.BranchCount != 1 || got.Branches[0].NodeCount != 2 || !got.Branches[0].Terminal {
		t.Fatalf("unexpected session summary: %#v", got)
	}
	wantLabels := []string{"current_session", "has_history", "has_tool_results", "has_tool_calls", "branched", "transcript_sanitized", "compacted"}
	if !reflect.DeepEqual(got.Labels, wantLabels) {
		t.Fatalf("labels = %#v, want %#v", got.Labels, wantLabels)
	}
}

func TestBuildHostUIHandoffEnvelopeFromOperatorLinesIsDisplaySafe(t *testing.T) {
	lines := []HostDiagnosticOperatorLineObservation{
		{
			Source:              "hostwiring_reader",
			Key:                 "progress.ready",
			Available:           true,
			Status:              "ready",
			OperatorDisplayLine: "kind=progress;status=ready;tool_output=false",
			Boundaries:          []string{"no_channel_adapter_host_diagnostics_json_decode"},
			NextHostAction:      "render_operator_line",
		},
		{
			Source:              "hostwiring_reader",
			Key:                 "unsafe.diagnostic",
			OperatorDisplayLine: "url=https://secret.example/raw",
		},
	}
	envelope := BuildHostUIHandoffEnvelopeFromOperatorLines(lines, HostUIHandoffInput{Target: HostUIHandoffTargetLog, Source: "host_agent"})
	if envelope == nil || envelope.EntryCount != 2 || envelope.LatestEntry == nil {
		t.Fatalf("unexpected envelope: %#v", envelope)
	}
	if envelope.Entries[1].DisplayLine != "redacted" {
		t.Fatalf("unsafe line was not redacted: %#v", envelope.Entries[1])
	}
	if envelope.LatestEntry == &envelope.Entries[1] {
		t.Fatal("latest entry must be a copy, not a slice element pointer")
	}
	envelope.Entries[1].Key = "mutated"
	if envelope.LatestEntry.Key != "unsafe.diagnostic" {
		t.Fatalf("latest entry aliases entries slice: %#v", envelope.LatestEntry)
	}
	fields, ok := RenderHostUIHandoffLogFields(envelope.Entries[0])
	if !ok || !strings.Contains(fields, "line=kind=progress;status=ready;tool_output=false") {
		t.Fatalf("unexpected rendered fields: %q, %v", fields, ok)
	}
}

func TestNormalizeHostUIHandoffValues(t *testing.T) {
	if got := NormalizeHostUIHandoffToken(" host:ready "); got != "host:ready" {
		t.Fatalf("token = %q", got)
	}
	if got := NormalizeHostUIHandoffToken("https://secret.example"); got != "redacted" {
		t.Fatalf("unsafe token = %q", got)
	}
	if got := NormalizeHostUIHandoffDisplayLine(" kind=ready;status=ok "); got != "kind=ready;status=ok" {
		t.Fatalf("display line = %q", got)
	}
	if got := NormalizeHostUIHandoffDisplayLine("kind=ready\nstatus=ok"); got != "redacted" {
		t.Fatalf("multiline display line = %q", got)
	}
	if got := NormalizeHostUIHandoffDisplayLine("url=https://secret.example"); got != "redacted" {
		t.Fatalf("unsafe display line = %q", got)
	}
}

func TestBuildHostUIHandoffConsumerConformanceReport(t *testing.T) {
	envelope := testHostUIHandoffEnvelope()
	report := BuildHostUIHandoffConsumerConformanceReport(HostUIHandoffConsumerConformanceInput{
		Consumer:       "host_log_renderer",
		ExpectedTarget: HostUIHandoffTargetLog,
		ExpectedSource: "host_agent",
		Envelope:       envelope,
	})
	if !report.Passed || report.Status != "passed" || !report.EntriesDisplaySafe {
		t.Fatalf("unexpected conformance report: %#v", report)
	}
}

func TestHostUIHandoffNegativeContractsRemainOrderedAndDisplaySafe(t *testing.T) {
	unsafeEnvelope := &HostUIHandoffEnvelope{
		Schema:     HostUIHandoffSchemaV1,
		Surface:    HostUIHandoffSurface,
		Target:     HostUIHandoffTargetLog,
		Source:     "host_agent",
		EntryCount: 1,
		Entries: []HostUIHandoffEntry{{
			Target:      HostUIHandoffTargetLog,
			Source:      "host_reader",
			Kind:        HostUIHandoffKindHostDiagnosticOperatorLine,
			Key:         "unsafe.diagnostic",
			Available:   true,
			Status:      "ready",
			DisplayLine: "url=https://secret.example/raw",
			Boundaries:  []string{"display_safe_operator_line_only"},
		}},
		Boundaries: []string{"productshell_host_ui_handoff"},
	}
	latest := unsafeEnvelope.Entries[0]
	unsafeEnvelope.LatestEntry = &latest
	conformance := BuildHostUIHandoffConsumerConformanceReport(HostUIHandoffConsumerConformanceInput{
		Consumer:       "host_log_renderer",
		ExpectedTarget: HostUIHandoffTargetLog,
		ExpectedSource: "host_agent",
		Envelope:       unsafeEnvelope,
	})
	wantBlocked := []string{
		"host_ui_handoff_entry_not_display_safe",
		"host_ui_handoff_delivery_boundary_missing",
		"host_ui_handoff_no_host_diagnostics_decode_boundary_missing",
		"host_ui_handoff_display_safe_boundary_missing",
	}
	if conformance.Passed || !reflect.DeepEqual(conformance.BlockedReasons, wantBlocked) {
		t.Fatalf("unsafe conformance = %#v, want blocked reasons %#v", conformance, wantBlocked)
	}

	fields, ok := RenderHostUIHandoffLogFields(HostUIHandoffEntry{
		Target:         HostUIHandoffTargetLog,
		Source:         "host_reader",
		Kind:           HostUIHandoffKindHostDiagnosticOperatorLine,
		Key:            "unsafe.diagnostic",
		Available:      true,
		Status:         "ready",
		DisplayLine:    "url=https://secret.example/raw",
		BlockedReasons: []string{"unsafe{json}"},
	})
	if !ok || !strings.Contains(fields, "line=redacted") || !strings.Contains(fields, "blocked=redacted") {
		t.Fatalf("unsafe log fields = %q, %v", fields, ok)
	}
	for _, forbidden := range []string{"https://secret.example", "{", "}"} {
		if strings.Contains(fields, forbidden) {
			t.Fatalf("log fields leaked %q in %q", forbidden, fields)
		}
	}
}

func TestBuildHostUIHandoffRuntimeUseReportUsesPortableInput(t *testing.T) {
	envelope := testHostUIHandoffEnvelope()
	report := BuildHostUIHandoffRuntimeUseReport(HostUIHandoffRuntimeUseInput{
		Consumer:           "host_log_renderer",
		HostAdapter:        "host_agent",
		Target:             HostUIHandoffTargetLog,
		Source:             "host_agent",
		ConformanceRefs:    []string{"test:runtime_use"},
		Envelope:           envelope,
		ConsumedEntryCount: len(envelope.Entries),
		Boundaries:         []string{"host_agent_owns_log_delivery"},
	})
	if !report.Ready || report.Status != "ready" || !report.ConsumesHandoffEnvelope || report.ConsumerConformance == nil || !report.ConsumerConformance.Passed {
		t.Fatalf("unexpected runtime-use report: %#v", report)
	}
	blocked := BuildHostUIHandoffRuntimeUseReport(HostUIHandoffRuntimeUseInput{
		Consumer:               "host_log_renderer",
		HostAdapter:            "host_agent",
		Target:                 HostUIHandoffTargetLog,
		Source:                 "host_agent",
		Envelope:               envelope,
		ConsumedEntryCount:     len(envelope.Entries),
		DecodedHostDiagnostics: true,
	})
	if blocked.Ready || !reflect.DeepEqual(blocked.BlockedReasons, []string{"host_ui_handoff_runtime_use_raw_host_diagnostics_decode"}) {
		t.Fatalf("unexpected blocked reason order: %#v", blocked)
	}
	mismatch := BuildHostUIHandoffRuntimeUseReport(HostUIHandoffRuntimeUseInput{
		Consumer:           "host_log_renderer",
		HostAdapter:        "host_agent",
		Target:             HostUIHandoffTargetLog,
		Source:             "host_agent",
		Envelope:           envelope,
		ConsumedEntryCount: 0,
	})
	if mismatch.Ready || mismatch.ConsumedAllEntries ||
		!reflect.DeepEqual(mismatch.MissingInputs, []string{"host_ui_handoff:consumer_entry_use"}) ||
		!reflect.DeepEqual(mismatch.BlockedReasons, []string{
			"host_ui_handoff_runtime_use_consumer_entry_missing",
			"host_ui_handoff_runtime_use_entry_count_mismatch",
		}) {
		t.Fatalf("entry-count mismatch report = %#v", mismatch)
	}
}

func testHostUIHandoffEnvelope() *HostUIHandoffEnvelope {
	return BuildHostUIHandoffEnvelopeFromOperatorLines([]HostDiagnosticOperatorLineObservation{{
		Source:              "hostwiring_reader",
		Key:                 "progress.ready",
		Status:              "ready",
		OperatorDisplayLine: "kind=progress;status=ready",
	}}, HostUIHandoffInput{Target: HostUIHandoffTargetLog, Source: "host_agent"})
}
