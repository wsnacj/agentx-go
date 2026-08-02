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
	if output != "agentx-objective-hostkit-ok:satisfied:adapter:conformance:true" {
		t.Fatalf("unexpected output: %s", output)
	}
}

func TestRunCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := run(ctx)
	if err == nil || !strings.Contains(err.Error(), "objective blocked") {
		t.Fatalf("expected structured blocked result, got %v", err)
	}
}
