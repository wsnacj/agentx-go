package main

import (
	"context"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := run(ctx)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	const want = "agentx-channel-ok:sent:channel-conformance:session_channel_ready"
	if got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
