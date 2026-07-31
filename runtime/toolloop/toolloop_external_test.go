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
