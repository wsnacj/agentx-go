package agent

import (
	"os"
	"strings"
	"testing"
)

func TestChineseAPIReferenceCoversExportedContract(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, required := range []string{
		"Request", "Backend", "BackendFuncs", "Options", "Register",
		"Definitions", "NewTaskHandler", "NewSubagentsHandler", "NewAgentStepHandler",
		"TasksSpawnDefinition", "TasksWaitDefinition", "TasksRunDefinition",
		"TasksCancelDefinition", "TasksReplayDefinition", "TasksCollectDefinition",
		"TasksDeadletterListDefinition", "SubagentsDefinition", "AgentStepDefinition",
		"Experimental extension", "Host 责任", "context", "fanout",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
