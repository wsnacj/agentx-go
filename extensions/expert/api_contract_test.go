package expert_test

import (
	"os"
	"strings"
	"testing"
)

func TestExpertAPIReference(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"Spec", "Instructions", "Requirement", "Parse", "Normalize", "Project", "errors.Is/As", "Session/Subagent", "Experimental"} {
		if !strings.Contains(string(content), required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
