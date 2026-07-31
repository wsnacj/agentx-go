package hostkit_test

import (
	"context"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

type externalFactory struct{}

func (externalFactory) BuildRun(_ context.Context, request execution.Request) (hostkit.RunConfig, error) {
	return hostkit.RunConfig{
		RunID:     "external-run",
		SessionID: request.SessionID,
		Assembly: toolloop.AssemblyConfig{
			MaxRounds:   1,
			Coordinator: toolloop.CoordinatorConfig{Executor: externalExecutor{}},
		},
	}, nil
}

func (externalFactory) Shutdown(context.Context) error { return nil }

func (externalFactory) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

type externalExecutor struct{}

func (externalExecutor) ExecuteRound(context.Context, toolloop.RoundExecutionInput) (toolloop.RoundExecutionResult, error) {
	return toolloop.RoundExecutionResult{Kind: toolloop.OutcomeCompleted, Reply: "external reply"}, nil
}

func TestExternalPackageBuildsRunnableClient(t *testing.T) {
	client, err := hostkit.New(hostkit.Config{Factory: externalFactory{}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	result, err := client.Run(context.Background(), agentx.RunRequest{
		Input:     "run",
		SessionID: "external-session",
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Status != "completed" || result.Reply != "external reply" || result.RunID != "external-run" {
		t.Fatalf("result = %#v", result)
	}
}
