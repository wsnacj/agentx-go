package main

import (
	"context"
	"errors"
	"testing"
)

func TestRunUsesExplicitMemoryBackend(t *testing.T) {
	result, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "opened" || result.Backend != "memory" || result.Note != "https://93.184.216.34/agentx" {
		t.Fatalf("run() = %#v", result)
	}
}

func TestRunPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("run() error = %v", err)
	}
}
