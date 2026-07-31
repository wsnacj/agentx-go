package toolloop_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wsnacj/agentx-go/runtime/toolloop"
)

type externalStepper struct {
	stopAt int
	err    error
}

type externalRoundExecutor struct{}

type externalPhaseExecutor struct {
	steps  *[]string
	action bool
	gate   bool
}

func (executor externalPhaseExecutor) Request(context.Context, toolloop.RoundPhaseInput) (toolloop.RoundRequestResult, error) {
	*executor.steps = append(*executor.steps, "request")
	return toolloop.RoundRequestResult{ActionRequired: executor.action}, nil
}

func (executor externalPhaseExecutor) Observe(context.Context, toolloop.RoundPhaseInput) (toolloop.RoundObserveResult, error) {
	*executor.steps = append(*executor.steps, "observe")
	return toolloop.RoundObserveResult{Reply: "raw reply"}, nil
}

func (executor externalPhaseExecutor) BeforeAction(context.Context, toolloop.RoundPhaseInput) (bool, error) {
	*executor.steps = append(*executor.steps, "before_action")
	return executor.gate, nil
}

func (executor externalPhaseExecutor) Act(context.Context, toolloop.RoundPhaseInput) error {
	*executor.steps = append(*executor.steps, "act")
	return nil
}

func (externalRoundExecutor) ExecuteRound(_ context.Context, input toolloop.RoundExecutionInput) (toolloop.RoundExecutionResult, error) {
	if input.Round >= input.MaxRounds {
		return toolloop.RoundExecutionResult{Kind: toolloop.OutcomeCompleted, Reply: "done"}, nil
	}
	return toolloop.RoundExecutionResult{
		Kind:  toolloop.OutcomeContinue,
		Reply: "working",
		Continuation: &toolloop.RoundContinuation{
			NextChunks: []string{"continue"},
		},
	}, nil
}

func (stepper externalStepper) Step(_ context.Context, input toolloop.StepInput) (toolloop.StepResult, error) {
	if stepper.err != nil {
		return toolloop.StepResult{}, stepper.err
	}
	if input.Round >= stepper.stopAt {
		return toolloop.StepResult{Kind: toolloop.OutcomeCompleted}, nil
	}
	return toolloop.StepResult{Kind: toolloop.OutcomeContinue}, nil
}

func TestExternalConsumerCanDriveAndObserveToolLoop(t *testing.T) {
	runtime, err := toolloop.New(
		toolloop.Config{MaxRounds: 4},
		externalStepper{stopAt: 2},
	)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background())
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.Kind != toolloop.OutcomeCompleted || result.Round != 2 {
		t.Fatalf("result = %#v", result)
	}

	detector := toolloop.NewLoopDetector(toolloop.LoopDetectorConfig{
		Enabled:             true,
		RepeatThreshold:     2,
		PingPongThreshold:   4,
		NoProgressThreshold: 2,
	})
	calls := []toolloop.Call{{Name: "lookup", Arguments: `{"id":1}`}}
	runs := []toolloop.RunObservation{{Name: "lookup", Output: "same"}}
	if signal, ok := detector.Observe(1, calls, runs); ok {
		t.Fatalf("first signal = %#v", signal)
	}
	signal, ok := detector.Observe(2, calls, runs)
	if !ok || signal.Kind != toolloop.LoopKindNoProgress {
		t.Fatalf("signal = %#v ok=%t", signal, ok)
	}
}

func TestExternalConsumerRetainsStepErrorIdentity(t *testing.T) {
	sentinel := errors.New("external step failed")
	runtime, err := toolloop.New(
		toolloop.Config{MaxRounds: 2},
		externalStepper{err: sentinel},
	)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background())
	if !errors.Is(err, sentinel) || result.Round != 1 {
		t.Fatalf("result=%#v error=%v", result, err)
	}
}

func TestExternalConsumerCanUseRoundCoordinatorAsStepper(t *testing.T) {
	coordinator, err := toolloop.NewCoordinator(toolloop.CoordinatorConfig{
		Executor: externalRoundExecutor{},
	}, toolloop.RoundState{Chunks: []string{"start"}})
	if err != nil {
		t.Fatalf("NewCoordinator(): %v", err)
	}
	runtime, err := toolloop.New(toolloop.Config{MaxRounds: 2}, coordinator)
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background())
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.Kind != toolloop.OutcomeCompleted || result.Round != 2 {
		t.Fatalf("result = %#v", result)
	}
	state := coordinator.State()
	if state.FinalReply != "done" || len(state.Chunks) != 1 || state.Chunks[0] != "continue" {
		t.Fatalf("state = %#v", state)
	}
}

func TestExternalConsumerCanCoordinateRoundPhases(t *testing.T) {
	var steps []string
	coordinator, err := toolloop.NewRoundPhaseCoordinator(externalPhaseExecutor{
		steps:  &steps,
		action: true,
		gate:   true,
	})
	if err != nil {
		t.Fatalf("NewRoundPhaseCoordinator(): %v", err)
	}
	result, err := coordinator.Execute(context.Background(), toolloop.RoundPhaseInput{
		Round:     1,
		MaxRounds: 3,
	})
	if err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if result.Kind != toolloop.RoundPhaseActionCompleted || result.Reply != "raw reply" {
		t.Fatalf("result = %#v", result)
	}
	want := []string{"request", "observe", "before_action", "act"}
	if len(steps) != len(want) {
		t.Fatalf("steps = %v", steps)
	}
	for index := range want {
		if steps[index] != want[index] {
			t.Fatalf("steps = %v", steps)
		}
	}
}
