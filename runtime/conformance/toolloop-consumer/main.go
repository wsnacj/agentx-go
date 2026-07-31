package main

import (
	"context"
	"fmt"
	"os"

	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

func main() {
	result, err := run(context.Background(), 3)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("agentx-toolloop-ok:%d\n", result.Round)
}

func run(ctx context.Context, completeAt int) (toolloop.Result, error) {
	coordinator, err := toolloop.NewCoordinator(toolloop.CoordinatorConfig{
		Executor: conformanceRoundExecutor{completeAt: completeAt},
	}, toolloop.RoundState{Chunks: []string{"start"}})
	if err != nil {
		return toolloop.Result{}, err
	}
	runtime, err := toolloop.New(
		toolloop.Config{MaxRounds: 4},
		coordinator,
	)
	if err != nil {
		return toolloop.Result{}, err
	}
	return runtime.Run(ctx)
}

type conformanceRoundExecutor struct {
	completeAt int
}

func (executor conformanceRoundExecutor) ExecuteRound(
	_ context.Context,
	input toolloop.RoundExecutionInput,
) (toolloop.RoundExecutionResult, error) {
	if input.Round >= executor.completeAt {
		return toolloop.RoundExecutionResult{
			Kind:  toolloop.OutcomeCompleted,
			Reply: "done",
		}, nil
	}
	return toolloop.RoundExecutionResult{
		Kind:  toolloop.OutcomeContinue,
		Reply: "working",
		Continuation: &toolloop.RoundContinuation{
			NextChunks: []string{"continue"},
		},
	}, nil
}
