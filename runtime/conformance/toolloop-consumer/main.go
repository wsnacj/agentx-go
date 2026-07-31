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
	assembly, err := toolloop.NewAssembly(toolloop.AssemblyConfig{
		MaxRounds: 4,
		Coordinator: toolloop.CoordinatorConfig{
			Executor: conformanceRoundExecutor{completeAt: completeAt},
		},
		Initial: toolloop.RoundState{Chunks: []string{"start"}},
	})
	if err != nil {
		return toolloop.Result{}, err
	}
	result, err := assembly.Run(ctx)
	return result.Driver, err
}

type conformanceRoundExecutor struct {
	completeAt int
}

func (executor conformanceRoundExecutor) ExecuteRound(
	ctx context.Context,
	input toolloop.RoundExecutionInput,
) (toolloop.RoundExecutionResult, error) {
	phases, err := toolloop.NewRoundPhaseCoordinator(conformancePhaseExecutor{
		actionRequired: input.Round < executor.completeAt,
	})
	if err != nil {
		return toolloop.RoundExecutionResult{}, err
	}
	phaseResult, err := phases.Execute(ctx, toolloop.RoundPhaseInput{
		Round:     input.Round,
		MaxRounds: input.MaxRounds,
	})
	if err != nil {
		return toolloop.RoundExecutionResult{}, err
	}
	if input.Round >= executor.completeAt {
		return toolloop.RoundExecutionResult{
			Kind:  toolloop.OutcomeCompleted,
			Reply: phaseResult.Reply,
		}, nil
	}
	return toolloop.RoundExecutionResult{
		Kind:  toolloop.OutcomeContinue,
		Reply: phaseResult.Reply,
		Continuation: &toolloop.RoundContinuation{
			NextChunks: []string{"continue"},
		},
	}, nil
}

type conformancePhaseExecutor struct {
	actionRequired bool
}

func (executor conformancePhaseExecutor) Request(context.Context, toolloop.RoundPhaseInput) (toolloop.RoundRequestResult, error) {
	return toolloop.RoundRequestResult{ActionRequired: executor.actionRequired}, nil
}

func (executor conformancePhaseExecutor) Observe(context.Context, toolloop.RoundPhaseInput) (toolloop.RoundObserveResult, error) {
	if executor.actionRequired {
		return toolloop.RoundObserveResult{Reply: "working"}, nil
	}
	return toolloop.RoundObserveResult{Reply: "done"}, nil
}

func (conformancePhaseExecutor) BeforeAction(context.Context, toolloop.RoundPhaseInput) (bool, error) {
	return true, nil
}

func (conformancePhaseExecutor) Act(context.Context, toolloop.RoundPhaseInput) error {
	return nil
}
