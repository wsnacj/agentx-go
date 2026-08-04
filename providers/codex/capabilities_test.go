package codex_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/providers/codex"
)

func TestModelCapabilitiesProjectsReasoningConfiguration(t *testing.T) {
	got := (codex.ModelConfig{ReasoningDefault: "minimal"}).ModelCapabilities()
	if !got.TextGeneration || !got.ToolCalling || !got.ReasoningControl || got.Streaming {
		t.Fatalf("ModelCapabilities() = %#v", got)
	}
}
