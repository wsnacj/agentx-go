package main

import (
	"context"
	"testing"
)

func TestRun(t *testing.T) {
	output, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if output != "agentx-session-hostkit-ok:host_action_recorded:true:1:1" {
		t.Fatalf("unexpected output: %s", output)
	}
}
