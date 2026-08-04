package anthropic_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/providers/anthropic"
)

func TestModelCapabilitiesMatchesMessagesAdapter(t *testing.T) {
	got := (anthropic.ModelConfig{}).ModelCapabilities()
	if !got.TextGeneration || !got.ToolCalling || got.Streaming || got.VisionInput {
		t.Fatalf("ModelCapabilities() = %#v", got)
	}
}
