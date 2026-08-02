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
	if result.Status != "parsed" || result.Text != "canonical pdf" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
