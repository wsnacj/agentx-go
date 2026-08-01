package main

import (
	"context"
	"fmt"
	"sync"

	agentx "github.com/wsnacj/agentx-go"
	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

type consumerFactory struct {
	mu     sync.Mutex
	closed bool
}

func (factory *consumerFactory) BuildRun(_ context.Context, request execution.Request) (hostkit.RunConfig, error) {
	round, err := hostkit.NewModelToolRoundAdapter(hostkit.ModelToolRoundConfig{
		RequestModel: func(_ context.Context, input toolloop.RoundExecutionInput) (hostkit.ModelResult, error) {
			if input.Round == 1 {
				return hostkit.ModelResult{Response: llm.ChatResponse{
					Content: "tool requested",
					Calls:   []llm.FunctionCall{{Name: "lookup", Arguments: `{"topic":"agentx"}`}},
				}}, nil
			}
			return hostkit.ModelResult{Response: llm.ChatResponse{Content: "hostkit-conformance:2"}}, nil
		},
		ExecuteTools: func(context.Context, hostkit.ModelToolRoundExchange) (hostkit.ToolResult, error) {
			return hostkit.ToolResult{
				Runs:       []toolloop.RunObservation{{Name: "lookup", Output: "portable result"}},
				NextChunks: []string{"portable result"},
			}, nil
		},
	})
	if err != nil {
		return hostkit.RunConfig{}, err
	}
	return hostkit.RunConfig{
		RunID:     "hostkit-conformance",
		SessionID: request.SessionID,
		Assembly: toolloop.AssemblyConfig{
			MaxRounds: 3,
			Coordinator: toolloop.CoordinatorConfig{
				Executor: round,
			},
			Initial: toolloop.RoundState{Chunks: []string{request.Input}},
		},
	}, nil
}

func (factory *consumerFactory) Shutdown(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	factory.mu.Lock()
	defer factory.mu.Unlock()
	factory.closed = true
	return nil
}

func (*consumerFactory) ClassifyError(error) agentx.ErrorCode {
	return agentx.CodeExecutionFailed
}

func run(ctx context.Context) (string, error) {
	client, err := hostkit.New(hostkit.Config{Factory: &consumerFactory{}})
	if err != nil {
		return "", err
	}
	result, runErr := client.Run(ctx, agentx.RunRequest{
		Input:     "exercise portable host kit",
		SessionID: "hostkit-session",
	})
	shutdownErr := client.Shutdown(context.Background())
	if runErr != nil {
		return "", runErr
	}
	if shutdownErr != nil {
		return "", shutdownErr
	}
	return fmt.Sprintf("agentx-hostkit-ok:%s:%s", result.Status, result.Reply), nil
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
