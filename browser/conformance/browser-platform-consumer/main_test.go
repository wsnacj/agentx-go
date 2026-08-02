package main

import (
	"context"
	"errors"
	"testing"
)

func TestFixedVersionBrowserPlatformConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if value.ToolStatus != "opened" || value.ToolBackend != "memory" || !value.RepairApplied || value.BrowserdStatus != "ready" || !value.NoManagedProcess {
		t.Fatalf("result=%#v", value)
	}
}

func TestFixedVersionBrowserPlatformConsumerPropagatesCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v, want context.Canceled", err)
	}
}
