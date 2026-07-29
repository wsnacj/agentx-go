package main

import (
	"encoding/json"
	"strings"
	"testing"

	agentxprotocol "github.com/wsnacj/agentx-go/runtime/protocol"
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
