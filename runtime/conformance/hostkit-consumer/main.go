package main

import (
	"context"
	"fmt"
	"sync"

	agentx "github.com/wsnacj/agentx-go"
	"github.com/wsnacj/agentx-go/runtime/execution"
	"github.com/wsnacj/agentx-go/runtime/hostkit"
	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

type consumerFactory struct {
	mu     sync.Mutex
	closed bool
}

func (factory *consumerFactory) BuildRun(_ context.Context, request execution.Request) (hostkit.RunConfig, error) {
	return hostkit.RunConfig{
		RunID:     "hostkit-conformance",
		SessionID: request.SessionID,
		Assembly: toolloop.AssemblyConfig{
			MaxRounds: 3,
			Coordinator: toolloop.CoordinatorConfig{
				Executor: &scriptedModelToolRound{},
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

type scriptedModelToolRound struct{}

func (*scriptedModelToolRound) ExecuteRound(_ context.Context, input toolloop.RoundExecutionInput) (toolloop.RoundExecutionResult, error) {
	if input.Round == 1 {
		return toolloop.RoundExecutionResult{
			Kind:  toolloop.OutcomeContinue,
			Reply: "tool requested",
			Continuation: &toolloop.RoundContinuation{
				Calls:      []toolloop.Call{{Name: "lookup", Arguments: `{"topic":"agentx"}`}},
				Runs:       []toolloop.RunObservation{{Name: "lookup", Output: "portable result"}},
				NextChunks: []string{"portable result"},
			},
		}, nil
	}
	return toolloop.RoundExecutionResult{
		Kind:  toolloop.OutcomeCompleted,
		Reply: "hostkit-conformance:2",
	}, nil
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
