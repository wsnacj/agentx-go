package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !value.Verified {
		t.Fatalf("result: %#v", value)
	}
}
