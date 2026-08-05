package connector_test

import (
	"os"
	"strings"
	"testing"
)

func TestConnectorAPIReference(t *testing.T) {
	content, err := os.ReadFile("API.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"Spec", "Normalize", "Project", "credential-free", "不启动进程", "Experimental"} {
		if !strings.Contains(string(content), required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
