package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	agentxmedia "github.com/wsnacj/agentx-go/runtime/mediaartifact"
	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
	agentxsafeerror "github.com/wsnacj/agentx-go/runtime/telemetry/safeerror"
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
