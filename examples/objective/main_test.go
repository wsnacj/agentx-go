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
	if !strings.Contains(output, "agentx-objective-hostkit-ok:satisfied") {
		t.Fatalf("unexpected output: %s", output)
	}
}
