package main

import (
	"context"
	"runtime"
	"testing"
)

func TestFixedVersionVideoFramesConsumer(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stubs are unavailable on windows")
	}
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !value.Verified {
		t.Fatalf("result=%#v", value)
	}
}

func TestFixedVersionVideoFramesConsumerHonorsCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stubs are unavailable on windows")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := run(ctx); err == nil {
		t.Fatal("expected canceled consumer run")
	}
}
