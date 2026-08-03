package process

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
		"LocalOptions", "LocalAdapter", "NewLocalAdapter", "Command", "CommandResult",
		"Termination", "ListRequest", "ListResult", "Run", "List", "ErrorCode", "AsError",
		"Experimental extension", "显式 opt-in", "Host 责任", "非 sandbox",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("API.md missing %q", required)
		}
	}
}
