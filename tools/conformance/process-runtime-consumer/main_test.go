package main

import (
	"context"
	"testing"
)

func TestFixedVersionProcessRuntimeConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !value.Verified {
		t.Fatalf("result=%#v", value)
	}
}
