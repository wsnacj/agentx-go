package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumer(t *testing.T) {
	result, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "recognized" || result.Text != "canonical document" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
