package ark_test

import (
	"testing"

	"github.com/wsnacj/agentx-go/providers/ark"
)

func TestResponseModelCapabilitiesMatchesResponsesPath(t *testing.T) {
	got := ark.ResponseModelCapabilities()
	if !got.TextGeneration || !got.ToolCalling || !got.Streaming || !got.ReasoningControl || got.VisionInput {
		t.Fatalf("ResponseModelCapabilities() = %#v", got)
	}
}
