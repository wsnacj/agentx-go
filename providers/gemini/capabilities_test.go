package gemini_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/providers/gemini"
)

func TestModelCapabilitiesKeepsHighLevelToolPathFailClosed(t *testing.T) {
	got := (gemini.ModelConfig{Capability: gemini.Capability{Vision: true, LocalFiles: true, Streaming: true}}).ModelCapabilities()
	if !got.TextGeneration || !got.VisionInput || !got.LocalMediaInput || !got.Streaming || got.ToolCalling {
		t.Fatalf("ModelCapabilities() = %#v", got)
	}
}
