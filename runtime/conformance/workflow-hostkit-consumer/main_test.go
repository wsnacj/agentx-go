package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumerRunsWithoutHSOrReplace(t *testing.T) {
	output, err := run(context.Background())
	if err != nil {
		t.Fatalf("run(): %v", err)
	}
	const want = "agentx-workflow-hostkit-ok:run-workflow-conformance:completed:workflow-hostkit-conformance"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}
