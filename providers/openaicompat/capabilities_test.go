package openaicompat_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/providers/openaicompat"
)

func TestModelCapabilitiesProjectsExplicitConfiguration(t *testing.T) {
	enabled := true
	got := (openaicompat.ModelConfig{Capability: openaicompat.Capability{
		Vision: true, LocalFiles: true, Streaming: true, Bots: true,
		ReasoningEffort: &enabled, ParallelToolCalls: &enabled,
	}}).ModelCapabilities()
	if !got.TextGeneration || !got.ToolCalling || !got.VisionInput || !got.Streaming ||
		!got.LocalMediaInput || !got.ReasoningControl || !got.ParallelTools || !got.BotCompletion {
		t.Fatalf("ModelCapabilities() = %#v", got)
	}
}
