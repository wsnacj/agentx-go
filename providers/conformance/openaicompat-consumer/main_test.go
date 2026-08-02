package main

import (
	"context"
	"testing"
)

func TestFixedVersionConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if value.Content != "ready" || value.Tool != "lookup" || value.TotalTokens != 3 || value.Authorization != "Bearer fixture-token" {
		t.Fatalf("result = %#v", value)
	}
}
