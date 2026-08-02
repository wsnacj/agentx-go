package main

import (
	"context"
	"reflect"
	"testing"
)

func TestFixedVersionGeneralToolsConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"apply_patch", "cron", "diffs", "edit", "http_request",
		"memory_get", "memory_search", "message", "read", "write",
	}
	if !value.Verified || !reflect.DeepEqual(value.Registered, want) || !reflect.DeepEqual(value.Executed, want) {
		t.Fatalf("result=%#v", value)
	}
}
