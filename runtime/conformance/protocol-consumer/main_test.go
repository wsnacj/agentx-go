package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	agentxbudget "github.com/wsnacj/agentx-go/runtime/budget"
	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	agentxpromptcontext "github.com/wsnacj/agentx-go/runtime/promptcontext"
	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
	agentxtoolerrors "github.com/wsnacj/agentx-go/runtime/toolerrors"
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
