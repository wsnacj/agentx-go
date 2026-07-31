package main

import (
	"context"
	"errors"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

func TestFixedVersionConsumerRunsAndShutsDownThroughCanonicalExecution(t *testing.T) {
	client, err := newClient()
	if err != nil {
		t.Fatalf("newClient(): %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "run",
		SessionID: " fixed-version-session ",
	})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.RunID != "execution-conformance-run" ||
		result.SessionID != "fixed-version-session" ||
		result.Status != "completed" ||
		result.Reply != "agentx-execution-ok" {
		t.Fatalf("Run() result = %#v", result)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown(): %v", err)
	}
	if _, err := client.Run(context.Background(), agentx.RunRequest{Input: "closed"}); !errors.Is(
		err,
		&agentx.Error{Code: agentx.CodeClientClosed},
	) {
		t.Fatalf("Run() after Shutdown error = %v", err)
	}
}
