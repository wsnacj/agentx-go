package main

import (
	"context"
	"errors"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
)

func TestFixedVersionConsumerConstructsRunsAndShutsDown(t *testing.T) {
	client, err := newClient(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("newClient() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "run",
		SessionID: "fixed-version-session",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "completed" ||
		result.Reply != "agentx-runtime-construction-ok" ||
		result.SessionID != "fixed-version-session" {
		t.Fatalf("Run() result = %#v", result)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if _, err := client.Run(context.Background(), agentx.RunRequest{Input: "closed"}); !errors.Is(
		err,
		&agentx.Error{Code: agentx.CodeClientClosed},
	) {
		t.Fatalf("Run() after Shutdown error = %v", err)
	}
}
