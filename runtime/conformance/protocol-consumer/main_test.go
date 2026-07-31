package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	agentxbudget "github.com/wsnacj/agentx-go/runtime/budget"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	agentxpromptcontext "github.com/wsnacj/agentx-go/runtime/promptcontext"
	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
	agentxtelemetry "github.com/wsnacj/agentx-go/runtime/telemetry"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestCanonicalProtocolConsumer(t *testing.T) {
	payload, err := canonicalEventJSON()
	if err != nil {
		t.Fatalf("canonicalEventJSON(): %v", err)
	}

	var event agentxprotocol.RunEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if event.SchemaVersion != agentxprotocol.RunEventSchemaV1 {
		t.Fatalf("schema_version = %q, want %q", event.SchemaVersion, agentxprotocol.RunEventSchemaV1)
	}
	if event.Kind != "tool.call.completed" ||
		event.RunID != "run-consumer-1" ||
		event.ToolName != "browser_open" ||
		event.Status != "completed" {
		t.Fatalf("normalized event = %#v", event)
	}
}

func TestCanonicalWorkflowBindingStateConsumer(t *testing.T) {
	state, err := canonicalWorkflowBindingState()
	if err != nil {
		t.Fatalf("canonicalWorkflowBindingState(): %v", err)
	}
	if got := state["report"]; got != "ready" {
		t.Fatalf("state report = %#v, want ready", got)
	}
}

func TestCanonicalWorkflowTransitionConsumer(t *testing.T) {
	visited, err := canonicalWorkflowTransition()
	if err != nil {
		t.Fatalf("canonicalWorkflowTransition(): %v", err)
	}
	if !reflect.DeepEqual(visited, []string{"collect", "report"}) {
		t.Fatalf("visited = %#v, want collect/report", visited)
	}
}

func TestCanonicalWorkflowJournalConsumer(t *testing.T) {
	operations, err := canonicalWorkflowJournal()
	if err != nil {
		t.Fatalf("canonicalWorkflowJournal(): %v", err)
	}
	want := []string{
		"load:run-consumer-1",
		"create:run-consumer-1",
		"event:workflow.start",
		"event:workflow.node.state.input",
		"node:running",
		"event:workflow.node.start",
		"event:workflow.node.state.output",
		"node:completed",
		"event:workflow.node.finish",
		"load:run-consumer-1",
		"update:run-consumer-1",
		"event:workflow.finish",
	}
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("operations = %#v, want %#v", operations, want)
	}
}

func TestCanonicalProtocolValidationErrorContract(t *testing.T) {
	err := agentxprotocol.ValidateRunEvent(agentxprotocol.RunEvent{})
	if err == nil {
		t.Fatal("ValidateRunEvent() error = nil")
	}
	if got, want := err.Error(), `agentx/runtime/protocol: kind is required`; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestCanonicalProtocolSchemaIdentities(t *testing.T) {
	values := []string{
		agentxprotocol.RuntimeSchemaV1,
		agentxprotocol.RunEventSchemaV1,
		agentxprotocol.TraceSpanSchemaV1,
		agentxprotocol.ToolExecutionPlanSchemaV1,
		agentxprotocol.HandoffSchemaV1,
		agentxprotocol.SandboxManifestSchemaV1,
		agentxprotocol.ArtifactVersionSchemaV1,
		agentxprotocol.ArtifactLinkSchemaV1,
	}
	if len(values) != 8 {
		t.Fatalf("schema identity count = %d, want 8", len(values))
	}
	for _, value := range values {
		if !strings.HasPrefix(value, "agentx.") || !strings.HasSuffix(value, ".v1") {
			t.Fatalf("unexpected schema identity %q", value)
		}
	}
}

func TestCanonicalSafeErrorConsumer(t *testing.T) {
	payload, err := canonicalSafeErrorJSON()
	if err != nil {
		t.Fatalf("canonicalSafeErrorJSON(): %v", err)
	}
	if got, want := string(payload), `{"class":"runtime_error","code":"upstream_failed","identity":"consumer-error-1"}`; got != want {
		t.Fatalf("safeerror JSON = %s, want %s", got, want)
	}
	if strings.Contains(string(payload), "private consumer sentinel") {
		t.Fatalf("safeerror JSON leaked raw cause: %s", payload)
	}

	attrs := agentxsafeerror.AppendAttrs(nil, "tool_", agentxsafeerror.Projection{
		Class:    "runtime_error",
		Code:     "upstream_failed",
		Identity: "consumer-error-1",
	})
	if attrs["tool_error_identity"] != "consumer-error-1" {
		t.Fatalf("safeerror attrs = %#v", attrs)
	}
	if got := agentxsafeerror.Identity(" material "); got != "40b30b4e8f0d137056ac497e859ea198c1a00db4267d1ade9c458d04024e2981" {
		t.Fatalf("safeerror identity = %q", got)
	}
}

func TestCanonicalMediaArtifactConsumer(t *testing.T) {
	payload, err := canonicalMediaArtifactJSON()
	if err != nil {
		t.Fatalf("canonicalMediaArtifactJSON(): %v", err)
	}
	if got, want := string(payload), `{"source":"nodes","kind":"video","path":".agentx/nodes/capture.mp4","mime_type":"video/mp4","format":"mp4","bytes":4096,"duration_ms":2500,"fps":30,"has_audio":false}`; got != want {
		t.Fatalf("media artifact JSON = %s, want %s", got, want)
	}

	var decoded agentxmedia.Descriptor
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if decoded.HasAudio == nil || *decoded.HasAudio ||
		decoded.Source != "nodes" ||
		decoded.Kind != "video" ||
		decoded.Bytes != 4096 {
		t.Fatalf("decoded descriptor = %#v", decoded)
	}

	typ := reflect.TypeFor[agentxmedia.Descriptor]()
	if typ.NumField() != 20 {
		t.Fatalf("Descriptor field count = %d, want 20", typ.NumField())
	}
	if field := typ.Field(0); field.Name != "Source" || field.Tag.Get("json") != "source" {
		t.Fatalf("first field = %s json:%q", field.Name, field.Tag.Get("json"))
	}
	if field := typ.Field(19); field.Name != "CreatedAt" || field.Tag.Get("json") != "created_at,omitempty" {
		t.Fatalf("last field = %s json:%q", field.Name, field.Tag.Get("json"))
	}

	var compileShape agentxmedia.Descriptor
	compileShape.Source = "consumer"
	if compileShape.Source != "consumer" {
		t.Fatalf("compile shape = %#v", compileShape)
	}
}

func TestCanonicalToolArgumentErrorConsumer(t *testing.T) {
	typed, err := canonicalToolArgumentError()
	if err != nil {
		t.Fatalf("canonicalToolArgumentError(): %v", err)
	}
	if typed.Tool != "browser" ||
		typed.Code != agentxtoolerrors.ToolArgumentErrorCodeInvalidArgumentObject ||
		typed.Detail != "decode: top-level JSON object is required" ||
		!typed.Repairable ||
		typed.SafeAutorepair ||
		len(typed.AllowedRepairs) != 1 ||
		typed.AllowedRepairs[0].Kind != agentxtoolerrors.ToolArgumentRepairReturnValidJSONObject {
		t.Fatalf("typed tool argument error = %#v", typed)
	}

	fields := []string{" url ", "", "method", "url"}
	invalidErr := agentxtoolerrors.NewInvalidToolArgumentError(" http_request ", fields, "")
	fields[0] = "changed"
	invalid, ok := agentxtoolerrors.AsToolArgumentError(invalidErr)
	if !ok {
		t.Fatalf("AsToolArgumentError(%T) = false", invalidErr)
	}
	if invalid.Tool != "http_request" ||
		invalid.Error() != "http_request: invalid arguments" ||
		!reflect.DeepEqual(invalid.InvalidFields, []string{"url", "method"}) ||
		!reflect.DeepEqual(invalid.AllowedRepairs, []agentxtoolerrors.ToolArgumentRepair{{
			Kind: agentxtoolerrors.ToolArgumentRepairFixInvalidField,
			To:   "url,method",
		}}) {
		t.Fatalf("invalid tool argument error = %#v", invalid)
	}

	input := []string{"query"}
	genericErr := agentxtoolerrors.NewToolArgumentError(" search ", agentxtoolerrors.ToolArgumentErrorOptions{
		Code:          " custom ",
		MissingFields: input,
	})
	input[0] = "changed"
	generic, ok := agentxtoolerrors.AsToolArgumentError(genericErr)
	if !ok || !reflect.DeepEqual(generic.MissingFields, []string{"query"}) {
		t.Fatalf("generic tool argument error = %#v", generic)
	}
}

func TestCanonicalBudgetConsumer(t *testing.T) {
	controller := agentxbudget.NewController()

	tests := []struct {
		name string
		snap agentxbudget.Snapshot
		want agentxbudget.Verdict
	}{
		{
			name: "ok",
			snap: agentxbudget.Snapshot{ToolCalls: 7, DurationMs: 799},
			want: agentxbudget.Verdict{Allowed: true, Stage: agentxbudget.StageOK},
		},
		{
			name: "warn",
			snap: agentxbudget.Snapshot{ToolCalls: 8, DurationMs: 800},
			want: agentxbudget.Verdict{
				Allowed: true,
				Stage:   agentxbudget.StageWarn,
				Warnings: []string{
					"budget near limit (max_tool_calls): 8/10",
					"budget near limit (max_duration_ms): 800/1000ms",
				},
			},
		},
		{
			name: "soft stop",
			snap: agentxbudget.Snapshot{ToolCalls: 11, DurationMs: 1200},
			want: agentxbudget.Verdict{
				Allowed: false,
				Stage:   agentxbudget.StageSoftStop,
				Reason:  agentxbudget.ReasonMaxToolCalls,
			},
		},
		{
			name: "hard stop",
			snap: agentxbudget.Snapshot{DurationMs: 1001},
			want: agentxbudget.Verdict{
				Allowed: false,
				Stage:   agentxbudget.StageHardStop,
				Reason:  agentxbudget.ReasonMaxDurationMs,
			},
		},
	}

	limit := agentxbudget.Limit{MaxToolCalls: 10, MaxDurationMs: 1000}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := controller.Check(limit, tc.snap); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Check() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestCanonicalPromptContextConsumer(t *testing.T) {
	now := time.Date(2026, 7, 30, 1, 2, 3, 0, time.UTC)
	got := agentxpromptcontext.Build(agentxpromptcontext.BuildInput{
		Now:       now,
		Timezone:  "Asia/Shanghai",
		SessionID: "session-1",
		Model:     "model-1",
	})
	if got.Now != now ||
		got.Timezone != "Asia/Shanghai" ||
		got.SessionID != "session-1" ||
		got.Model != "model-1" {
		t.Fatalf("prompt context = %#v", got)
	}
	if text := got.TimestampText(); text != "2026-07-30T09:02:03+08:00" {
		t.Fatalf("TimestampText() = %q", text)
	}

	got.Timezone = "invalid/timezone"
	if text := got.TimestampText(); text != now.Format(time.RFC3339) {
		t.Fatalf("invalid timezone TimestampText() = %q", text)
	}

	var compileShape agentxpromptcontext.Context
	compileShape = agentxpromptcontext.Context(agentxpromptcontext.BuildInput{})
	if !compileShape.Now.IsZero() {
		t.Fatalf("compile shape = %#v", compileShape)
	}
}

func TestCanonicalTelemetryConsumer(t *testing.T) {
	payload, err := canonicalTelemetryJSON()
	if err != nil {
		t.Fatalf("canonicalTelemetryJSON(): %v", err)
	}
	var toolEvent agentxtelemetry.ToolEvent
	if err := json.Unmarshal(payload, &toolEvent); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if toolEvent.SchemaVersion != agentxtelemetry.ToolEventSchemaV1 ||
		toolEvent.Kind != agentxtelemetry.ToolEventKindCompleted ||
		toolEvent.Tool != "browser_open" ||
		toolEvent.DurationMs != 42 {
		t.Fatalf("tool event = %#v", toolEvent)
	}

	semantic := agentxtelemetry.ProjectSemanticRunEvents(agentxtelemetry.Event{
		Name: agentxtelemetry.SemanticRunSourceEventRunStart,
		Attrs: map[string]any{
			"resume": true,
		},
	})
	if len(semantic) != 1 ||
		semantic[0].Kind != agentxtelemetry.SemanticRunEventKindRunResumed {
		t.Fatalf("semantic projection = %#v", semantic)
	}
	summary := agentxtelemetry.SummarizeSemanticRunEvents(semantic)
	if summary.TotalEvents != 1 || summary.RunResumed != 1 {
		t.Fatalf("semantic summary = %#v", summary)
	}

	recordPayload, err := json.Marshal(agentxtelemetry.Event{
		Name: agentxtelemetry.SemanticRunSourceEventRunStart,
		Attrs: map[string]any{
			"resume": true,
		},
	})
	if err != nil {
		t.Fatalf("Marshal(record): %v", err)
	}
	replayed, trace := agentxtelemetry.ReplaySemanticRunEventsFromStoredRecords(
		[]agentxtelemetry.StoredRawEventRecord{{
			EventID:     "event-1",
			RunID:       "run-1",
			PayloadJSON: string(recordPayload),
			CreatedAt:   1710000000123,
		}},
	)
	if len(replayed) != 1 ||
		replayed[0].SourceEventID != "event-1" ||
		trace.ProjectedEventCount != 1 ||
		trace.InvalidEventCount != 0 {
		t.Fatalf("replay = %#v trace = %#v", replayed, trace)
	}
}

func TestCanonicalTelemetryPrivateJSONLConsumer(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "observations")
	path := filepath.Join(directory, "events.jsonl")
	sink, err := agentxtelemetry.NewJSONLSink(path)
	if err != nil {
		t.Fatalf("NewJSONLSink(): %v", err)
	}
	if err := sink.Emit(context.Background(), agentxtelemetry.Event{
		Component: "runner",
		Name:      "run.start",
	}); err != nil {
		t.Fatalf("Emit(): %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(): %v", err)
	}
	if !strings.Contains(string(raw), `"schema_version":"v1"`) ||
		!strings.Contains(string(raw), `"name":"run.start"`) {
		t.Fatalf("JSONL = %s", raw)
	}
	if runtime.GOOS != "windows" {
		directoryInfo, err := os.Stat(directory)
		if err != nil {
			t.Fatalf("Stat(directory): %v", err)
		}
		fileInfo, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat(file): %v", err)
		}
		if got, want := directoryInfo.Mode().Perm(), os.FileMode(0o700); got != want {
			t.Fatalf("directory mode = %#o, want %#o", got, want)
		}
		if got, want := fileInfo.Mode().Perm(), os.FileMode(0o600); got != want {
			t.Fatalf("file mode = %#o, want %#o", got, want)
		}
	}
}

func TestCanonicalWorkflowConsumer(t *testing.T) {
	payload, err := canonicalWorkflowJSON()
	if err != nil {
		t.Fatalf("canonicalWorkflowJSON(): %v", err)
	}
	var spec agentxworkflow.Spec
	if err := json.Unmarshal(payload, &spec); err != nil {
		t.Fatalf("Unmarshal(): %v", err)
	}
	if spec.ID != "consumer-workflow" ||
		spec.PlanningMode != agentxworkflow.PlanningBounded ||
		spec.EntryNode != "collect" ||
		len(spec.Nodes) != 1 ||
		spec.Nodes[0].Kind != agentxworkflow.NodeCollect ||
		spec.Nodes[0].ExecutionMode != agentxworkflow.ExecInline ||
		spec.Nodes[0].Retry.MaxAttempts != 2 ||
		len(spec.Edges) != 1 ||
		len(spec.StateSchema) != 1 ||
		len(spec.ArtifactSchema) != 1 ||
		len(spec.EvaluatorSchema) != 1 {
		t.Fatalf("workflow spec = %#v", spec)
	}
	if got := spec.Nodes[0].Config["format"]; got != "markdown" {
		t.Fatalf("workflow config format = %#v", got)
	}

	typ := reflect.TypeFor[agentxworkflow.Spec]()
	if typ.NumField() != 15 {
		t.Fatalf("Spec field count = %d, want 15", typ.NumField())
	}
	if field := typ.Field(0); field.Name != "ID" || field.Tag.Get("json") != "id,omitempty" {
		t.Fatalf("first field = %s json:%q", field.Name, field.Tag.Get("json"))
	}
	if field := typ.Field(14); field.Name != "DefaultContract" || field.Tag.Get("json") != "default_contract,omitempty" {
		t.Fatalf("last field = %s json:%q", field.Name, field.Tag.Get("json"))
	}
}
