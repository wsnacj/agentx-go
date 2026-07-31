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
	runtime, err := toolloop.New(
		toolloop.Config{MaxRounds: 4},
		conformanceStepper{completeAt: completeAt},
	)
	if err != nil {
		return toolloop.Result{}, err
	}
	return runtime.Run(ctx)
}

type conformanceStepper struct {
	completeAt int
}

func (stepper conformanceStepper) Step(
	_ context.Context,
	input toolloop.StepInput,
) (toolloop.StepResult, error) {
	if input.Round >= stepper.completeAt {
		return toolloop.StepResult{Kind: toolloop.OutcomeCompleted}, nil
	}
	return toolloop.StepResult{Kind: toolloop.OutcomeContinue}, nil
}
