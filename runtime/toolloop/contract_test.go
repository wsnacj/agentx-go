package toolloop

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type stepperFunc func(context.Context, StepInput) (StepResult, error)

func (fn stepperFunc) Step(ctx context.Context, input StepInput) (StepResult, error) {
	return fn(ctx, input)
}

func TestRuntimeDrivesRoundsToCompletion(t *testing.T) {
	var inputs []StepInput
	runtime, err := New(Config{MaxRounds: 4}, stepperFunc(func(_ context.Context, input StepInput) (StepResult, error) {
		inputs = append(inputs, input)
		if input.Round == 3 {
			return StepResult{Kind: OutcomeCompleted}, nil
		}
		return StepResult{Kind: OutcomeContinue}, nil
	}))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(context.Background())
	if err != nil {
		t.Fatalf("Run(): %v", err)
	}
	if result.Kind != OutcomeCompleted || result.Round != 3 {
		t.Fatalf("result = %#v", result)
	}
	want := []StepInput{
		{Round: 1, MaxRounds: 4},
		{Round: 2, MaxRounds: 4},
		{Round: 3, MaxRounds: 4},
	}
	if !reflect.DeepEqual(inputs, want) {
		t.Fatalf("inputs = %#v, want %#v", inputs, want)
	}
}

func TestRuntimeReportsMaxRoundsAndTermination(t *testing.T) {
	tests := map[string]struct {
		step StepResult
		want Result
	}{
		"max rounds": {
			step: StepResult{Kind: OutcomeContinue},
			want: Result{Kind: OutcomeMaxRounds, Round: 2},
		},
		"terminated": {
			step: StepResult{Kind: OutcomeTerminated},
			want: Result{Kind: OutcomeTerminated, Round: 1},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			runtime, err := New(Config{MaxRounds: 2}, stepperFunc(func(_ context.Context, _ StepInput) (StepResult, error) {
				return test.step, nil
			}))
			if err != nil {
				t.Fatalf("New(): %v", err)
			}
			result, err := runtime.Run(context.Background())
			if err != nil {
				t.Fatalf("Run(): %v", err)
			}
			if result != test.want {
				t.Fatalf("result = %#v, want %#v", result, test.want)
			}
		})
	}
}

func TestRuntimePreservesStepErrorAndContextIdentity(t *testing.T) {
	type contextKey string
	sentinel := errors.New("step failed")
	ctx := context.WithValue(context.Background(), contextKey("identity"), "kept")
	runtime, err := New(Config{MaxRounds: 3}, stepperFunc(func(got context.Context, input StepInput) (StepResult, error) {
		if got != ctx || got.Value(contextKey("identity")) != "kept" {
			t.Fatalf("context identity was not preserved")
		}
		if input.Round != 1 {
			t.Fatalf("round = %d", input.Round)
		}
		return StepResult{}, sentinel
	}))
	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	result, err := runtime.Run(ctx)
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run() error = %v", err)
	}
	if result.Round != 1 || result.Kind != "" {
		t.Fatalf("result = %#v", result)
	}
}

func TestLoopDetectorPriorityAndReplay(t *testing.T) {
	detector := NewLoopDetector(LoopDetectorConfig{
		Enabled:             true,
		RepeatThreshold:     3,
		PingPongThreshold:   4,
		NoProgressThreshold: 3,
	})
	calls := []Call{{Name: " lookup ", Arguments: `{"b":2,"a":1}`}}
	runs := []RunObservation{{Name: "lookup", Output: `{"status":"pending"}`}}
	for round := 1; round <= 2; round++ {
		if signal, ok := detector.Observe(round, calls, runs); ok {
			t.Fatalf("round %d signal = %#v", round, signal)
		}
	}
	signal, ok := detector.Observe(
		3,
		[]Call{{Name: "LOOKUP", Arguments: `{"a":1,"b":2}`}},
		runs,
	)
	if !ok || signal.Kind != LoopKindNoProgress || signal.Count != 3 {
		t.Fatalf("signal = %#v ok=%t", signal, ok)
	}
	replay, ok := detector.ShouldSuppressReplay(calls)
	if !ok || replay.Kind != LoopKindReplay || replay.Round != 3 {
		t.Fatalf("replay = %#v ok=%t", replay, ok)
	}
}

func TestLoopDetectorDetectsPingPong(t *testing.T) {
	detector := NewLoopDetector(LoopDetectorConfig{
		Enabled:             true,
		RepeatThreshold:     3,
		PingPongThreshold:   4,
		NoProgressThreshold: 3,
	})
	callA := []Call{{Name: "lookup", Arguments: `{"id":"a"}`}}
	callB := []Call{{Name: "lookup", Arguments: `{"id":"b"}`}}
	_, _ = detector.Observe(1, callA, []RunObservation{{Name: "lookup", Output: "a1"}})
	_, _ = detector.Observe(2, callB, []RunObservation{{Name: "lookup", Output: "b1"}})
	_, _ = detector.Observe(3, callA, []RunObservation{{Name: "lookup", Output: "a2"}})
	signal, ok := detector.Observe(4, callB, []RunObservation{{Name: "lookup", Output: "b2"}})
	if !ok || signal.Kind != LoopKindPingPong || signal.Count != 4 {
		t.Fatalf("signal = %#v ok=%t", signal, ok)
	}
}

func TestFailureFuseTracksResetAndInvalidArguments(t *testing.T) {
	fuse := NewFailureFuse(FailureFuseConfig{Enabled: true, Threshold: 3})
	failed := []FailureObservation{{
		Tool:       "llm_task",
		Failed:     true,
		ErrorClass: "invalid_args",
	}}
	if signal, ok := fuse.Observe(1, failed); ok {
		t.Fatalf("first signal = %#v", signal)
	}
	signal, ok := fuse.Observe(2, failed)
	if !ok ||
		signal.Tool != "llm_task" ||
		signal.Count != 2 ||
		signal.ErrorClass != "invalid_args" {
		t.Fatalf("signal = %#v ok=%t", signal, ok)
	}

	resetFuse := NewFailureFuse(FailureFuseConfig{Enabled: true, Threshold: 2})
	_, _ = resetFuse.Observe(1, []FailureObservation{{
		Tool: "echo", Failed: true, ErrorClass: "timeout",
	}})
	if signal, ok := resetFuse.Observe(2, []FailureObservation{{Tool: "echo"}}); ok {
		t.Fatalf("success signal = %#v", signal)
	}
	if signal, ok := resetFuse.Observe(3, []FailureObservation{{
		Tool: "echo", Failed: true, ErrorClass: "timeout",
	}}); ok {
		t.Fatalf("signal after reset = %#v", signal)
	}
}
