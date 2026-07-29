package main

import (
	"encoding/json"
	"strings"
	"testing"

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
