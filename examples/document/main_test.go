package main

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunUsesExplicitMemoryPorts(t *testing.T) {
	result, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "success" || !strings.Contains(result.Answer, "Revenue is 42") {
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
