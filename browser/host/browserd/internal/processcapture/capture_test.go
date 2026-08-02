package processcapture

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestRunBoundsProcessOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	cmd := exec.Command("sh", "-c", "printf '0123456789abcdef'")
	result, err := Run(cmd, Limits{StdoutBytes: 8, StderrBytes: 8})
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("Run error = %v, want limit error", err)
	}
	if string(result.Stdout) != "01234567" || result.StdoutObservedBytes != 16 || !result.StdoutLimitExceeded {
		t.Fatalf("Run result = %#v", result)
	}
	if strings.Contains(result.Summary(), "0123456789abcdef") {
		t.Fatalf("summary leaked captured output: %q", result.Summary())
	}
}
