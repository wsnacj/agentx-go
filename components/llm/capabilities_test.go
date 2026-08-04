package llm

import "testing"

func TestModelCapabilitiesZeroValueIsFailClosed(t *testing.T) {
	got := ModelCapabilities{}
	if got.TextGeneration || got.ToolCalling || got.VisionInput || got.Streaming ||
		got.LocalMediaInput || got.ReasoningControl || got.ParallelTools || got.BotCompletion {
		t.Fatalf("zero ModelCapabilities must not declare support: %#v", got)
	}
}
