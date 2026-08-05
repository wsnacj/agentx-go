package mcp_test

import (
	"os"
	"strings"
	"testing"
)

func TestMCPAPIReference(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"2025-11-25", "Transport", "Initialize", "DiscoverTools", "CallTool", "ToolSet",
		"DefinitionProvider", "Executor", "isError=true", "experimental MCP Tasks", "Experimental",
	} {
		if !strings.Contains(string(content), required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
