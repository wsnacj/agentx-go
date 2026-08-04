package main

import (
	"context"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	output, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(output, "agentx-session-hostkit-ok:") {
		t.Fatalf("unexpected session output: %s", output)
	}

	resumeOutput, err := runResume(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resumeOutput, "agentx-resume-hostkit-ok:") {
		t.Fatalf("unexpected resume output: %s", resumeOutput)
	}
	if !strings.Contains(resumeOutput, "cross-construction") {
		t.Fatalf("expected cross-construction resume proof, got %s", resumeOutput)
	}
}
