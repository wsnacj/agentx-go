package gemini_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/providers/gemini"
)

func TestModelCapabilitiesReflectsExplicitHighLevelToolSupport(t *testing.T) {
	got := (gemini.ModelConfig{Capability: gemini.Capability{Vision: true, LocalFiles: true, Streaming: true, ToolCalling: true}}).ModelCapabilities()
	if !got.TextGeneration || !got.VisionInput || !got.LocalMediaInput || !got.Streaming || !got.ToolCalling {
		t.Fatalf("ModelCapabilities() = %#v", got)
	}
	if disabled := (gemini.ModelConfig{}).ModelCapabilities(); disabled.ToolCalling {
		t.Fatalf("default ModelCapabilities() = %#v", disabled)
	}
}
