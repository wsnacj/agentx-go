package main

import (
	"context"
	"testing"
)

func TestFixedVersionDocumentToolsConsumer(t *testing.T) {
	result, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "success" || result.Answer != "canonical document tools [p1]" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
