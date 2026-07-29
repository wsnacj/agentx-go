package main

import (
	"context"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
)

var _ agentx.ExecutionAdapter = (*conformanceAdapter)(nil)

func TestExternalStyleCanonicalConsumer(t *testing.T) {
	output, err := runProbe()
	if err != nil {
		t.Fatalf("runProbe() error = %v", err)
	}
	if output != "completed conformance-session" {
		t.Fatalf("runProbe() output = %q", output)
	}
}

func TestCanonicalContractCompileShape(t *testing.T) {
	var newClient func(agentx.Config) (*agentx.Client, error) = agentx.New
	client, err := newClient(agentx.Config{Adapter: &conformanceAdapter{}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.Run(context.Background(), agentx.RunRequest{Input: "compile shape"}); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if err := client.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
