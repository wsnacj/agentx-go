package main

import (
	"context"
	"testing"
)

func TestRunDataPlaneConsumer(t *testing.T) {
	got, err := runDataPlaneConsumer(context.Background())
	if err != nil {
		t.Fatalf("runDataPlaneConsumer(): %v", err)
	}
	const want = "agentx-run-data-plane-ok:event-1:artifact-consumer-1:1:1"
	if got != want {
		t.Fatalf("result = %q, want %q", got, want)
	}
}
