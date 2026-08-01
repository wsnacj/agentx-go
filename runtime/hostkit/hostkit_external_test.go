package hostkit_test

import (
	"context"
	"testing"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

type externalFactory struct{}

func (externalFactory) BuildRun(_ context.Context, request execution.Request) (hostkit.RunConfig, error) {
	round, err := hostkit.NewModelToolRoundAdapter(hostkit.ModelToolRoundConfig{
		RequestModel: func(_ context.Context, input toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
			if input.Round == 1 {
				return hostkit.ModelResult{Response: llm.ChatResponse{
					Calls: []llm.FunctionCall{{Name: "lookup", Arguments: `{"q":"agentx"}`}},
				}}, nil
			}
			return hostkit.ModelResult{Response: llm.ChatResponse{Content: "external reply"}}, nil
		},
		ExecuteTools: func(context.Context, hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
			return hostkit.ToolResult{NextChunks: []string{"lookup: agentx"}}, nil
		},
	})
	if err != nil {
		return hostkit.RunConfig{}, err
	}
	return hostkit.RunConfig{
		RunID:     "external-run",
		SessionID: request.SessionID,
		Assembly: toolloop.AssemblyConfig{
			MaxRounds:   2,
			Coordinator: toolloop.CoordinatorConfig{Executor: round},
		},
	}, nil
}

func (externalFactory) Shutdown(context.Context) error { return nil }

func (externalFactory) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
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
