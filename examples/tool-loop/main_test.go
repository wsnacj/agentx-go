package main

import (
	"context"
	"strings"
	"testing"
)

func TestToolLoopDirectAnswerExample(t *testing.T) {
	result, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "completed" || result.RunID != "run-tool-loop-example" || !strings.Contains(result.Reply, `"tool":"diffs"`) {
		t.Fatalf("result=%#v", result)
	}
}
