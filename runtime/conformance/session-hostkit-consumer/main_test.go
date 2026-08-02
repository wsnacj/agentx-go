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

func TestRunResume(t *testing.T) {
	output, err := runResume(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if output != "agentx-resume-hostkit-ok:objective_runtime_scheduler_resume_daemon_service_completed:1:true" {
		t.Fatalf("unexpected output: %s", output)
	}
}
