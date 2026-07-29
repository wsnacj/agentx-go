package agentx_test

import (
	"os"
	"strings"
	"testing"
)

func TestReferenceCoversW1Contract(t *testing.T) {
	content, err := os.ReadFile("docs/reference/agentx.md")
	if err != nil {
		t.Fatal(err)
	}
	reference := string(content)
	required := []string{
		"AdapterRunRequest",
		"AdapterRunResult",
		"Client",
		"CodeCanceled",
		"CodeClientClosed",
		"CodeDeadlineExceeded",
		"CodeExecutionFailed",
		"CodeInvalidArgument",
		"CodeShutdownFailed",
		"CodeUnsupportedProfile",
		"Config",
		"Error",
		"ErrorCode",
		"ExecutionAdapter",
		"ExecutionProfile",
		"New",
		"RunRequest",
		"RunResult",
		"Client.Run",
		"Client.Shutdown",
		"Error.Error",
		"Error.Is",
		"Error.Unwrap",
	}
	for _, name := range required {
		marker := "<!-- api:" + name + " -->"
		if count := strings.Count(reference, marker); count != 1 {
			t.Errorf("reference marker %q count = %d, want 1", marker, count)
		}
	}
}
