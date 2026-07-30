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

func TestProjectSemanticRunEventsMapsRunAndApprovalEvents(t *testing.T) {
	interrupted := ProjectSemanticRunEvents(Event{
		Component: "store",
		Name:      "checkpoint.upsert",
		SessionID: "sess-1",
		Attrs: map[string]any{
			"stage":                   "tool_pending",
			"round":                   3,
			"resume_envelope_schema":  "run_resume_envelope.v1",
			"resume_mode":             "replay",
			"interruption_kind":       "checkpoint",
			"pending_tool_call_count": 2,
			"has_pending_tool_calls":  true,
		},
	})
	if len(interrupted) != 1 || interrupted[0].Kind != SemanticRunEventKindRunInterrupted {
		t.Fatalf("expected checkpoint interruption event, got %#v", interrupted)
	}
	if interrupted[0].Stage != "tool_pending" ||
		interrupted[0].Round != 3 ||
		interrupted[0].Reason != "pending_tool_calls" ||
		interrupted[0].Checkpoint == nil ||
		interrupted[0].Checkpoint.PendingToolCallCount != 2 ||
		!interrupted[0].Checkpoint.HasPendingToolCalls {
		t.Fatalf("unexpected checkpoint interruption projection: %#v", interrupted[0])
	}

	resumed := ProjectSemanticRunEvents(Event{
		Component: "runner",
		Name:      "run.start",
		SessionID: "sess-1",
		Attrs: map[string]any{
			"resume": true,
		},
	})
	if len(resumed) != 1 ||
		resumed[0].Kind != SemanticRunEventKindRunResumed ||
		resumed[0].Reason != "resume_requested" {
		t.Fatalf("expected resumed projection, got %#v", resumed)
	}

	requested := ProjectSemanticRunEvents(Event{
		Component: "tool",
		Name:      "tool.runtime_decision",
		Tool:      "exec",
		Status:    "confirm_required",
		Attrs: map[string]any{
			"action":           "exec",
			"checked":          true,
			"requires_confirm": true,
			"reason":           "approval_required",
			"policy_source":    "tool_call_policy",
		},
	})
	if len(requested) != 1 ||
		requested[0].Kind != SemanticRunEventKindApprovalRequested ||
		requested[0].Approval == nil ||
		!requested[0].Approval.RequiresConfirm ||
		requested[0].Approval.Decision != "confirm_required" ||
		requested[0].Approval.PolicySource != "tool_call_policy" {
		t.Fatalf("expected approval requested projection, got %#v", requested)
	}

	resolved := ProjectSemanticRunEvents(Event{
		Component: "tool",
		Name:      "tool.approval",
		Tool:      "write",
		Status:    "denied",
		Attrs: map[string]any{
			"reason": "protected_metadata_write_denied",
		},
	})
	if len(resolved) != 1 ||
		resolved[0].Kind != SemanticRunEventKindApprovalResolved ||
		resolved[0].Approval == nil ||
		!resolved[0].Approval.Denied ||
		resolved[0].Approval.Decision != "denied" ||
		resolved[0].Reason != "protected_metadata_write_denied" {
		t.Fatalf("expected approval resolved projection, got %#v", resolved)
	}
}

func TestProjectSemanticRunEventsMapsGuardianReviewLifecycle(t *testing.T) {
	started := ProjectSemanticRunEvents(Event{
		Component: "tool",
		Name:      "tool.guardian_review.start",
		Tool:      "browser_act",
		Status:    "start",
		Attrs: map[string]any{
			"review_id":     "review-1",
			"stage":         "approval",
			"risk":          "high",
			"runtime_owner": "guardian",
		},
	})
	if len(started) != 1 ||
		started[0].Kind != SemanticRunEventKindApprovalRequested ||
		started[0].Approval == nil ||
		started[0].Approval.ReviewID != "review-1" ||
		started[0].Approval.Stage != "approval" {
		t.Fatalf("expected guardian review request projection, got %#v", started)
	}

	finished := ProjectSemanticRunEvents(Event{
		Component: "tool",
		Name:      "tool.guardian_review.finish",
		Tool:      "browser_act",
		Status:    "denied",
		Attrs: map[string]any{
			"review_id": "review-1",
			"denied":    true,
			"outcome":   "denied",
			"reviewer":  "policy",
			"rationale": "unsafe_target",
		},
	})
	if len(finished) != 1 ||
		finished[0].Kind != SemanticRunEventKindApprovalResolved ||
		finished[0].Approval == nil ||
		!finished[0].Approval.Denied ||
		finished[0].Approval.Decision != "denied" ||
		finished[0].Approval.Reviewer != "policy" ||
		finished[0].Reason != "unsafe_target" {
		t.Fatalf("expected guardian review resolved projection, got %#v", finished)
	}
}

func TestSemanticRunEventProjectableSourceEventsStayExplicit(t *testing.T) {
	want := []string{
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
	got := SemanticRunEventProjectableSourceEvents()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected projectable source events: got %#v want %#v", got, want)
	}
	got[0] = "mutated"
	if SemanticRunEventProjectableSourceEvents()[0] != SemanticRunSourceEventRunStart {
		t.Fatalf("projectable source events must be returned as a copy")
	}
	for _, sourceEvent := range want {
		if !IsSemanticRunEventProjectableSourceEvent(" " + sourceEvent + " ") {
			t.Fatalf("expected %q to be recognized as projectable", sourceEvent)
		}
	}
	if IsSemanticRunEventProjectableSourceEvent("tool.finish") {
		t.Fatalf("tool.finish must not be a semantic run event projection source")
	}
}

func TestSemanticRunEventProjectableSourceEventsStillProjectWithReferenceFixtures(t *testing.T) {
	fixtures := map[string]Event{
		SemanticRunSourceEventRunStart: {
			Name: SemanticRunSourceEventRunStart,
			Attrs: map[string]any{
				"resume": true,
			},
		},
		SemanticRunSourceEventRunFinish: {
			Name: SemanticRunSourceEventRunFinish,
			Attrs: map[string]any{
				"final_status": "incomplete",
			},
		},
		SemanticRunSourceEventCheckpointUpsert: {
			Name: SemanticRunSourceEventCheckpointUpsert,
			Attrs: map[string]any{
				"stage":                   "tool_pending",
				"pending_tool_call_count": 1,
			},
		},
		SemanticRunSourceEventHookPermissionRequest: {
			Name: SemanticRunSourceEventHookPermissionRequest,
			Tool: "exec",
			Attrs: map[string]any{
				"reason":           "approval_required",
				"requires_confirm": true,
			},
		},
		SemanticRunSourceEventToolGuardianReviewStart: {
			Name: SemanticRunSourceEventToolGuardianReviewStart,
			Tool: "browser_act",
			Attrs: map[string]any{
				"review_id": "review-1",
				"risk":      "high",
			},
		},
		SemanticRunSourceEventToolApproval: {
			Name:   SemanticRunSourceEventToolApproval,
			Tool:   "write",
			Status: "denied",
			Attrs: map[string]any{
				"reason": "protected_metadata_write_denied",
			},
		},
		SemanticRunSourceEventToolGuardianReview: {
			Name:   SemanticRunSourceEventToolGuardianReview,
			Tool:   "browser_act",
			Status: "denied",
			Attrs: map[string]any{
				"denied":    true,
				"review_id": "review-1",
			},
		},
		SemanticRunSourceEventToolGuardianReviewFinish: {
			Name:   SemanticRunSourceEventToolGuardianReviewFinish,
			Tool:   "browser_act",
			Status: "denied",
			Attrs: map[string]any{
				"denied":    true,
				"review_id": "review-1",
			},
		},
		SemanticRunSourceEventToolRuntimeDecision: {
			Name:   SemanticRunSourceEventToolRuntimeDecision,
			Tool:   "exec",
			Status: "confirm_required",
			Attrs: map[string]any{
				"requires_confirm": true,
				"reason":           "approval_required",
			},
		},
	}
	for _, sourceEvent := range SemanticRunEventProjectableSourceEvents() {
		event, ok := fixtures[sourceEvent]
		if !ok {
			t.Fatalf("missing reference fixture for %q", sourceEvent)
		}
		if projected := ProjectSemanticRunEvents(event); len(projected) == 0 {
			t.Fatalf("expected reference fixture for %q to project at least one semantic run event", sourceEvent)
		}
	}
	if len(fixtures) != len(SemanticRunEventProjectableSourceEvents()) {
		t.Fatalf("fixture set drifted: fixtures=%d sources=%d", len(fixtures), len(SemanticRunEventProjectableSourceEvents()))
	}
}

func TestSemanticRunEventJSONLSinkFiltersAndProjectsSemanticEvents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "semantic-run-events.jsonl")
	sink, err := NewSemanticRunEventJSONLSink(path)
	if err != nil {
		t.Fatalf("new semantic run event jsonl sink: %v", err)
	}
	if err := sink.Emit(context.Background(), Event{Component: "tool", Name: "tool.finish"}); err != nil {
		t.Fatalf("emit non-semantic event: %v", err)
	}
	if err := sink.Emit(context.Background(), Event{
		Component: "runner",
		Name:      "run.start",
		Attrs: map[string]any{
			"resume": true,
		},
	}); err != nil {
		t.Fatalf("emit semantic event: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read semantic run events jsonl: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected one projected semantic event, got %d: %s", len(lines), string(raw))
	}
	var event SemanticRunEvent
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatalf("decode projected semantic run event: %v", err)
	}
	if event.SchemaVersion != SemanticRunEventSchemaV1 ||
		event.Kind != SemanticRunEventKindRunResumed ||
		event.Reason != "resume_requested" {
		t.Fatalf("unexpected projected semantic run event: %#v", event)
	}
}

func TestReplaySemanticRunEventsFromStoredRecordsProjectsTrace(t *testing.T) {
	checkpointPayload, err := json.Marshal(Event{
		Component: "store",
		Name:      "checkpoint.upsert",
		Attrs: map[string]any{
			"stage":                   "tool_pending",
			"pending_tool_call_count": 1,
			"has_pending_tool_calls":  true,
		},
	})
	if err != nil {
		t.Fatalf("marshal checkpoint payload: %v", err)
	}
	runPayload, err := json.Marshal(Event{Component: "runner", Name: "run.start"})
	if err != nil {
		t.Fatalf("marshal run payload: %v", err)
	}
	events, trace := ReplaySemanticRunEventsFromStoredRecords([]StoredRawEventRecord{
		{
			EventID:     "event-checkpoint",
			RunID:       "run-1",
			BranchID:    "branch-1",
			NodeExecID:  "node-1",
			Name:        "checkpoint.upsert",
			PayloadJSON: string(checkpointPayload),
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
			Name:        "checkpoint.upsert",
			PayloadJSON: `{"component":`,
			CreatedAt:   1710000000789,
		},
	})

	if len(events) != 1 {
		t.Fatalf("expected one semantic projection, got %#v", events)
	}
	if events[0].SourceEventID != "event-checkpoint" ||
		events[0].RunID != "run-1" ||
		events[0].BranchID != "branch-1" ||
		events[0].NodeExecID != "node-1" ||
		!events[0].Timestamp.Equal(time.UnixMilli(1710000000123).UTC()) {
		t.Fatalf("expected runstore metadata on projected event, got %#v", events[0])
	}
	if trace.SchemaVersion != SemanticRunEventProjectionTraceSchemaV1 ||
		trace.Source != SemanticRunEventProjectionSourceRunstore ||
		trace.SourceEventCount != 3 ||
		trace.DecodedEventCount != 2 ||
		trace.ProjectedEventCount != 1 ||
		trace.InvalidEventCount != 1 ||
		trace.Summary.RunInterrupted != 1 ||
		trace.Summary.PendingToolCallInterrupts != 1 {
		t.Fatalf("unexpected projection trace: %#v", trace)
	}
	if len(trace.Sources) != 3 ||
		trace.Sources[0].EventID != "event-checkpoint" ||
		!reflect.DeepEqual(trace.Sources[0].ProjectedKinds, []string{SemanticRunEventKindRunInterrupted}) ||
		trace.Sources[1].ProjectedCount != 0 ||
		trace.Sources[2].Error != "invalid_payload_json" {
		t.Fatalf("unexpected source trace: %#v", trace.Sources)
	}
}

func TestSummarizeSemanticRunEvents(t *testing.T) {
	events := []SemanticRunEvent{
		{
			Kind:   SemanticRunEventKindRunInterrupted,
			Stage:  "tool_pending",
			Reason: "pending_tool_calls",
			Checkpoint: &SemanticRunCheckpointProjection{
				Stage:               "tool_pending",
				HasPendingToolCalls: true,
			},
		},
		{Kind: SemanticRunEventKindRunResumed},
		{
			Kind: SemanticRunEventKindApprovalRequested,
			Tool: "exec",
			Approval: &SemanticRunApprovalProjection{
				Tool:            "exec",
				Decision:        "confirm_required",
				RequiresConfirm: true,
				Reason:          "approval_required",
			},
		},
		{
			Kind: SemanticRunEventKindApprovalResolved,
			Tool: "write",
			Approval: &SemanticRunApprovalProjection{
				Tool:     "write",
				Decision: "denied",
				Denied:   true,
				Reason:   "protected_metadata_write_denied",
			},
		},
	}
	summary := SummarizeSemanticRunEvents(events)
	if summary.TotalEvents != 4 ||
		summary.RunInterrupted != 1 ||
		summary.RunResumed != 1 ||
		summary.ApprovalRequested != 1 ||
		summary.ApprovalResolved != 1 ||
		summary.ApprovalDenied != 1 ||
		summary.ApprovalRequiresConfirm != 1 ||
		summary.CheckpointInterruptions != 1 ||
		summary.PendingToolCallInterrupts != 1 {
		t.Fatalf("unexpected semantic run event summary: %#v", summary)
	}
	if summary.ByKind[SemanticRunEventKindRunInterrupted] != 1 ||
		summary.ByStage["tool_pending"] != 1 ||
		summary.ByTool["exec"] != 1 ||
		summary.ByTool["write"] != 1 ||
		summary.ByApprovalDecision["denied"] != 1 ||
		summary.ByApprovalReason["protected_metadata_write_denied"] != 1 ||
		summary.ByInterruptionReason["pending_tool_calls"] != 1 {
		t.Fatalf("unexpected summary maps: %#v", summary)
	}
}
