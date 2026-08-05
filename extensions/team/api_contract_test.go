package team_test

import (
	"os"
	"strings"
	"testing"
)

func TestTeamAPIReference(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"Spec", "Member.DependsOn", "BuildPlan", "topological", "Session/Subagent", "scheduler", "Experimental"} {
		if !strings.Contains(string(content), required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
