package tools_test

import (
	"testing"

	tools "github.com/wsnacj/agentx-go/tools"
)

func TestToolMetadataPreservesUnspecifiedBooleanHints(t *testing.T) {
	metadata := tools.ToolMetadata{
		Type:   "browser",
		Source: tools.ToolSourceBuiltin,
	}
	if metadata.ReadOnly != nil || metadata.ConcurrencySafe != nil || metadata.Destructive != nil {
		t.Fatalf("zero metadata must leave boolean hints unspecified: %#v", metadata)
	}
}
