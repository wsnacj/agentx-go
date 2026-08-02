package main

import (
	"context"
	"testing"
)

func TestNoHSRunnerConsumer(t *testing.T) {
	output, err := run(context.Background())
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	const want = "agentx-hostkit-ok:completed:hostkit-conformance:direct:hostkit-conformance:chat"
	if output != want {
		t.Fatalf("output = %q, want %q", output, want)
	}
}
