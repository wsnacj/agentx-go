package main

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {
	got, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"revenue":42,"status":"parsed"}` {
		t.Fatalf("output = %s", got)
	}
}

