package toolloop

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestAssemblyDrivesCoordinatorAndReturnsState(t *testing.T) {
	var rounds []int
	assembly, err := NewAssembly(AssemblyConfig{
		MaxRounds: 3,
		Coordinator: CoordinatorConfig{
			Executor: roundExecutorFunc(func(_ context.Context, input RoundExecutionInput) (RoundExecutionResult, error) {
				rounds = append(rounds, input.Round)
				if input.Round == 2 {
					return RoundExecutionResult{Kind: OutcomeCompleted, Reply: "done"}, nil
				}
				return RoundExecutionResult{
					Kind:  OutcomeContinue,
					Reply: "working",
					Continuation: &RoundContinuation{
						NextChunks:       []string{"next"},
						ForceNoToolCalls: true,
					},
				}, nil
			}),
		},
		Initial: RoundState{Chunks: []string{"initial"}},
	})
	if err != nil {
		t.Fatalf("NewAssembly(): %v", err)
	}
	result, err := assembly.Run(context.Background())
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.Driver.Kind != OutcomeCompleted || result.Driver.Round != 2 {
		t.Fatalf("driver result = %#v", result.Driver)
	}
	if !reflect.DeepEqual(rounds, []int{1, 2}) {
		t.Fatalf("rounds = %#v", rounds)
	}
	if result.State.FinalReply != "done" || !result.State.ForceNoToolCalls || !reflect.DeepEqual(result.State.Chunks, []string{"next"}) {
		t.Fatalf("state = %#v", result.State)
	}
	result.State.Chunks[0] = "mutated"
	if state := assembly.coordinator.State(); state.Chunks[0] != "next" {
		t.Fatalf("returned state aliases coordinator state: %#v", state)
	}
}

func TestAssemblyPreservesContextAndErrorIdentity(t *testing.T) {
	type contextKey string
	sentinel := errors.New("round failed")
	ctx := context.WithValue(context.Background(), contextKey("key"), "value")
	assembly, err := NewAssembly(AssemblyConfig{
		MaxRounds: 2,
		Coordinator: CoordinatorConfig{
			Executor: roundExecutorFunc(func(got context.Context, input RoundExecutionInput) (RoundExecutionResult, error) {
				if got != ctx || got.Value(contextKey("key")) != "value" || input.Round != 1 {
					t.Fatalf("context or input changed: ctx=%v input=%#v", got, input)
				}
				return RoundExecutionResult{}, sentinel
			}),
		},
	})
	if err != nil {
		t.Fatalf("NewAssembly(): %v", err)
	}
	result, err := assembly.Run(ctx)
	if !errors.Is(err, sentinel) || result.Driver.Round != 1 {
		t.Fatalf("Run() result = %#v error = %v", result, err)
	}
}

func TestAssemblyReturnsTerminationFact(t *testing.T) {
	assembly, err := NewAssembly(AssemblyConfig{
		MaxRounds: 2,
		Coordinator: CoordinatorConfig{
			Executor: roundExecutorFunc(func(context.Context, RoundExecutionInput) (RoundExecutionResult, error) {
				return RoundExecutionResult{
					Kind: OutcomeContinue,
					Continuation: &RoundContinuation{
						Failures: []FailureObservation{{Tool: "tool", Failed: true}},
					},
				}, nil
			}),
			FailureFuse: NewFailureFuse(FailureFuseConfig{Enabled: true, Threshold: 1}),
		},
	})
	if err != nil {
		t.Fatalf("NewAssembly(): %v", err)
	}
	result, err := assembly.Run(context.Background())
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.Driver.Kind != OutcomeTerminated || result.Termination == nil || result.Termination.Kind != TerminationFailureFuse {
		t.Fatalf("result = %#v", result)
	}
	result.Termination.Kind = TerminationHostPolicy
	if assembly.stepper.termination.Kind != TerminationFailureFuse {
		t.Fatalf("returned termination aliases assembly state: %#v", assembly.stepper.termination)
	}
}

func TestAssemblyValidationAndNilReceiver(t *testing.T) {
	_, err := NewAssembly(AssemblyConfig{MaxRounds: 1})
	if err == nil || err.Error() != "agentx tool loop: round executor is required" {
		t.Fatalf("missing executor error = %v", err)
	}
	_, err = NewAssembly(AssemblyConfig{
		Coordinator: CoordinatorConfig{Executor: roundExecutorFunc(func(context.Context, RoundExecutionInput) (RoundExecutionResult, error) {
			return RoundExecutionResult{}, nil
		})},
	})
	if err == nil || err.Error() != "agentx tool loop: max rounds must be positive" {
		t.Fatalf("invalid rounds error = %v", err)
	}
	var assembly *Assembly
	if _, err := assembly.Run(context.Background()); err == nil || err.Error() != "agentx tool loop: assembly runtime is required" {
		t.Fatalf("nil assembly error = %v", err)
	}
}
