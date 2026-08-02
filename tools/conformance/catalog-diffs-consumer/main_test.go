package main

import (
	"context"
	"testing"
)

func TestFixedVersionCatalogDiffsConsumer(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Registered) != 1 || value.Registered[0] != "diffs" || !value.Repaired || value.Additions != 1 || value.Deletions != 1 || value.Path != "sample.txt" {
		t.Fatalf("result=%#v", value)
	}
}
