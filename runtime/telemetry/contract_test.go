package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestEventReflectAndJSONContract(t *testing.T) {
	want := []struct {
		name string
		typ  reflect.Type
		tag  string
	}{
		{name: "SchemaVersion", typ: reflect.TypeFor[string](), tag: "schema_version"},
		{name: "Timestamp", typ: reflect.TypeFor[time.Time](), tag: "timestamp"},
		{name: "Component", typ: reflect.TypeFor[string](), tag: "component"},
		{name: "Name", typ: reflect.TypeFor[string](), tag: "name"},
		{name: "Level", typ: reflect.TypeFor[Level](), tag: "level"},
		{name: "SessionID", typ: reflect.TypeFor[string](), tag: "session_id,omitempty"},
		{name: "Round", typ: reflect.TypeFor[int](), tag: "round,omitempty"},
		{name: "Tool", typ: reflect.TypeFor[string](), tag: "tool,omitempty"},
		{name: "Model", typ: reflect.TypeFor[string](), tag: "model,omitempty"},
		{name: "Status", typ: reflect.TypeFor[string](), tag: "status,omitempty"},
		{name: "Attrs", typ: reflect.TypeFor[map[string]any](), tag: "attrs,omitempty"},
	}

	typ := reflect.TypeFor[Event]()
	if typ.NumField() != len(want) {
		t.Fatalf("Event field count = %d, want %d", typ.NumField(), len(want))
	}
	for i, fieldContract := range want {
		field := typ.Field(i)
		if field.Name != fieldContract.name ||
			field.Type != fieldContract.typ ||
			field.Tag.Get("json") != fieldContract.tag {
			t.Fatalf(
				"Event field[%d] = %s %s json:%q, want %s %s json:%q",
				i,
				field.Name,
				field.Type,
				field.Tag.Get("json"),
				fieldContract.name,
				fieldContract.typ,
				fieldContract.tag,
			)
		}
	}

	now := time.Date(2026, 7, 30, 8, 0, 0, 0, time.UTC)
	payload, err := json.Marshal(Event{
		SchemaVersion: EventSchemaV1,
		Timestamp:     now,
		Component:     "runner",
		Name:          "run.start",
		Level:         LevelInfo,
	})
	if err != nil {
		t.Fatalf("Marshal(): %v", err)
	}
	if got, want := string(payload), `{"schema_version":"v1","timestamp":"2026-07-30T08:00:00Z","component":"runner","name":"run.start","level":"info"}`; got != want {
		t.Fatalf("Event JSON = %s, want %s", got, want)
	}
}

func TestExportedSurfaceCompileShape(t *testing.T) {
	constants := []string{
		string(LevelDebug),
		string(LevelInfo),
		string(LevelWarn),
		string(LevelError),
		EventSchemaV1,
		ToolEventSchemaV1,
		ToolEventKindStarted,
		ToolEventKindArgumentsRepaired,
		ToolEventKindAuthorized,
		ToolEventKindExecuting,
		ToolEventKindCompleted,
		ToolEventKindFailed,
		ToolEventKindRetried,
		ToolEventKindResultMiddlewareObserved,
		ToolEventKindResultMiddlewareApplied,
		ToolEventKindProviderFallback,
		ToolEventSurfaceRetrieval,
		ToolEventSurfaceBrowser,
		ToolEventSurfacePDF,
		ToolEventSurfaceExec,
		ToolEventSurfaceOther,
		ToolEventProjectionTraceSchemaV1,
		ToolEventProjectionSourceRunstore,
		ToolEventProjectionSourceTelemetryJSONL,
		SemanticRunEventSchemaV1,
		SemanticRunEventKindRunInterrupted,
		SemanticRunEventKindRunResumed,
		SemanticRunEventKindApprovalRequested,
		SemanticRunEventKindApprovalResolved,
		SemanticRunSourceEventRunStart,
		SemanticRunSourceEventRunFinish,
		SemanticRunSourceEventCheckpointUpsert,
		SemanticRunSourceEventHookPermissionRequest,
		SemanticRunSourceEventToolGuardianReviewStart,
		SemanticRunSourceEventToolApproval,
		SemanticRunSourceEventToolGuardianReview,
		SemanticRunSourceEventToolGuardianReviewFinish,
		SemanticRunSourceEventToolRuntimeDecision,
		SemanticRunEventProjectionTraceSchemaV1,
		SemanticRunEventProjectionSourceRunstore,
		SemanticRunEventProjectionSourceTelemetryJSONL,
	}
	if len(constants) != 41 {
		t.Fatalf("exported constant count = %d, want 41", len(constants))
	}

	values := []any{
		Event{},
		JSONLSink{},
		Level(""),
		MultiSink{},
		SemanticRunApprovalProjection{},
		SemanticRunCheckpointProjection{},
		SemanticRunEvent{},
		SemanticRunEventJSONLSink{},
		SemanticRunEventProjectionSourceTrace{},
		SemanticRunEventProjectionTrace{},
		SemanticRunEventSummary{},
		SemanticRunTerminationProjection{},
		StoredRawEventRecord{},
		ToolEvent{},
		ToolEventJSONLSink{},
		ToolEventProjectionSourceTrace{},
		ToolEventProjectionTrace{},
		ToolEventSummary{},
		ToolProviderFallbackProjection{},
		ToolRepairProjection{},
		ToolResultMiddlewareProjection{},
		ToolResultOutputSchemaProjection{},
		ToolRuntimeDecisionProjection{},
		ToolSoftRejectionProjection{},
	}
	if len(values) != 24 {
		t.Fatalf("exported concrete type count = %d, want 24", len(values))
	}

	functions := []any{
		NewJSONLSink,
		NewToolEventJSONLSink,
		NewSemanticRunEventJSONLSink,
		ProjectToolEvents,
		NormalizeToolEventKind,
		ReplayToolEventsFromStoredRecords,
		SummarizeToolEvents,
		ToolEventSurfaceForTool,
		ProjectSemanticRunEvents,
		NormalizeSemanticRunEventKind,
		SemanticRunEventProjectableSourceEvents,
		IsSemanticRunEventProjectableSourceEvent,
		ReplaySemanticRunEventsFromStoredRecords,
		SummarizeSemanticRunEvents,
	}
	if len(functions) != 14 {
		t.Fatalf("exported function count = %d, want 14", len(functions))
	}

	var _ Sink = MultiSink{}
	var _ Sink = (*JSONLSink)(nil)
	var _ Sink = (*ToolEventJSONLSink)(nil)
	var _ Sink = (*SemanticRunEventJSONLSink)(nil)
}

func TestSinkNilAndErrorAggregationContract(t *testing.T) {
	var raw *JSONLSink
	var tool *ToolEventJSONLSink
	var semantic *SemanticRunEventJSONLSink
	for name, sink := range map[string]Sink{
		"raw":      raw,
		"tool":     tool,
		"semantic": semantic,
	} {
		if err := sink.Emit(context.Background(), Event{Name: "noop"}); err != nil {
			t.Fatalf("%s nil sink Emit() error = %v", name, err)
		}
	}

	err := (MultiSink{Sinks: []Sink{
		contractErrorSink("first"),
		nil,
		contractErrorSink("second"),
	}}).Emit(context.Background(), Event{Name: "run.start"})
	if err == nil || err.Error() != "first; second" {
		t.Fatalf("MultiSink error = %v, want %q", err, "first; second")
	}
	if !errors.Is(err, contractError("first")) {
		// MultiSink intentionally preserves text only; document that errors.Is is false.
		if !strings.Contains(err.Error(), "first") {
			t.Fatalf("MultiSink error lost first message: %v", err)
		}
	}
}

type contractError string

func (e contractError) Error() string { return string(e) }

type contractErrorSink string

func (s contractErrorSink) Emit(context.Context, Event) error {
	return contractError(s)
}
