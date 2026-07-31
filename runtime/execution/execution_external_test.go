package execution_test

import (
	"context"
	"errors"
	"testing"
	"time"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
)

func TestExternalHostProvidesUnifiedClientRunAndShutdown(t *testing.T) {
	host := &externalHost{}
	runtime, err := execution.New(host)
	if err != nil {
		t.Fatalf("execution.New(): %v", err)
	}
	client, err := agentx.New(agentx.Config{Adapter: runtime})
	if err != nil {
		t.Fatalf("agentx.New(): %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "external run",
		SessionID: " external-session ",
	})
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.RunID != "external-run" ||
		result.SessionID != "external-session" ||
		result.Status != "completed" ||
		result.Reply != "agentx-execution-ok" {
		t.Fatalf("Run() result = %#v", result)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := client.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown(): %v", err)
	}
	if host.shutdownCalls != 1 {
		t.Fatalf("shutdown calls = %d", host.shutdownCalls)
	}
	if _, err := client.Run(context.Background(), agentx.RunRequest{Input: "closed"}); !errors.Is(
		err,
		&agentx.Error{Code: agentx.CodeClientClosed},
	) {
		t.Fatalf("Run() after Shutdown error = %v", err)
	}
}

type externalHost struct {
	shutdownCalls int
}

func (*externalHost) Run(_ context.Context, request execution.Request) (*execution.Result, error) {
	return &execution.Result{
		RunID:     "external-run",
		SessionID: request.SessionID,
		Status:    "completed",
		Reply:     "agentx-execution-ok",
	}, nil
}

func (host *externalHost) Shutdown(context.Context) error {
	host.shutdownCalls++
	return nil
}

func (*externalHost) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}
