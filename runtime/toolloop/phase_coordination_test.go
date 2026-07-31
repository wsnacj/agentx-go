package toolloop

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type roundPhaseExecutorStub struct {
	steps       []string
	request     RoundRequestResult
	requestErr  error
	observe     RoundObserveResult
	observeErr  error
	proceed     bool
	beforeErr   error
	actErr      error
	seenContext context.Context
	seenInput   RoundPhaseInput
}

func (e *roundPhaseExecutorStub) Request(ctx context.Context, input RoundPhaseInput) (RoundRequestResult, error) {
	e.steps = append(e.steps, "request")
	e.seenContext, e.seenInput = ctx, input
	return e.request, e.requestErr
}

func (e *roundPhaseExecutorStub) Observe(context.Context, RoundPhaseInput) (RoundObserveResult, error) {
	e.steps = append(e.steps, "observe")
	return e.observe, e.observeErr
}

func (e *roundPhaseExecutorStub) BeforeAction(context.Context, RoundPhaseInput) (bool, error) {
	e.steps = append(e.steps, "before_action")
	return e.proceed, e.beforeErr
}

func (e *roundPhaseExecutorStub) Act(context.Context, RoundPhaseInput) error {
	e.steps = append(e.steps, "act")
	return e.actErr
}

func TestRoundPhaseCoordinatorNoActionStopsAfterObserveAndPreservesRawReply(t *testing.T) {
	executor := &roundPhaseExecutorStub{observe: RoundObserveResult{Reply: " raw reply "}}
	coordinator, err := NewRoundPhaseCoordinator(executor)
	if err != nil {
		t.Fatalf("NewRoundPhaseCoordinator(): %v", err)
	}
	ctx := context.WithValue(context.Background(), struct{}{}, "identity")
	input := RoundPhaseInput{Round: 2, MaxRounds: 4}
	result, err := coordinator.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if result.Kind != RoundPhaseNoAction || result.LastPhase != RoundPhaseObserve || result.Reply != " raw reply " {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(executor.steps, []string{"request", "observe"}) || executor.seenContext != ctx || executor.seenInput != input {
		t.Fatalf("steps=%v input=%#v", executor.steps, executor.seenInput)
	}
}

func TestRoundPhaseCoordinatorActionPathUsesGateBeforeAct(t *testing.T) {
	executor := &roundPhaseExecutorStub{
		request: RoundRequestResult{ActionRequired: true},
		observe: RoundObserveResult{Reply: "reply"},
		proceed: true,
	}
	coordinator, err := NewRoundPhaseCoordinator(executor)
	if err != nil {
		t.Fatalf("NewRoundPhaseCoordinator(): %v", err)
	}
	result, err := coordinator.Execute(context.Background(), RoundPhaseInput{Round: 1, MaxRounds: 3})
	if err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if result.Kind != RoundPhaseActionCompleted || result.LastPhase != RoundPhaseAct || result.Reply != "reply" {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(executor.steps, []string{"request", "observe", "before_action", "act"}) {
		t.Fatalf("order = %v", executor.steps)
	}
}

func TestRoundPhaseCoordinatorHostStopSkipsAct(t *testing.T) {
	executor := &roundPhaseExecutorStub{
		request: RoundRequestResult{ActionRequired: true},
		observe: RoundObserveResult{Reply: "reply"},
	}
	coordinator, err := NewRoundPhaseCoordinator(executor)
	if err != nil {
		t.Fatalf("NewRoundPhaseCoordinator(): %v", err)
	}
	result, err := coordinator.Execute(context.Background(), RoundPhaseInput{})
	if err != nil {
		t.Fatalf("Execute(): %v", err)
	}
	if result.Kind != RoundPhaseHostStopped || result.LastPhase != RoundPhaseBeforeAction {
		t.Fatalf("result = %#v", result)
	}
	if !reflect.DeepEqual(executor.steps, []string{"request", "observe", "before_action"}) {
		t.Fatalf("order = %v", executor.steps)
	}
}

func TestRoundPhaseCoordinatorPreservesErrorIdentityAndPhase(t *testing.T) {
	tests := []struct {
		name      string
		executor  *roundPhaseExecutorStub
		want      error
		wantPhase RoundPhase
	}{
		{name: "request", executor: &roundPhaseExecutorStub{requestErr: errors.New("request")}, wantPhase: RoundPhaseRequest},
		{name: "observe", executor: &roundPhaseExecutorStub{observeErr: errors.New("observe")}, wantPhase: RoundPhaseObserve},
		{name: "before action", executor: &roundPhaseExecutorStub{request: RoundRequestResult{ActionRequired: true}, beforeErr: errors.New("before")}, wantPhase: RoundPhaseBeforeAction},
		{name: "act", executor: &roundPhaseExecutorStub{request: RoundRequestResult{ActionRequired: true}, proceed: true, actErr: errors.New("act")}, wantPhase: RoundPhaseAct},
	}
	for index := range tests {
		test := &tests[index]
		switch test.wantPhase {
		case RoundPhaseRequest:
			test.want = test.executor.requestErr
		case RoundPhaseObserve:
			test.want = test.executor.observeErr
		case RoundPhaseBeforeAction:
			test.want = test.executor.beforeErr
		case RoundPhaseAct:
			test.want = test.executor.actErr
		}
		t.Run(test.name, func(t *testing.T) {
			coordinator, err := NewRoundPhaseCoordinator(test.executor)
			if err != nil {
				t.Fatalf("NewRoundPhaseCoordinator(): %v", err)
			}
			result, got := coordinator.Execute(context.Background(), RoundPhaseInput{})
			if !errors.Is(got, test.want) || result.LastPhase != test.wantPhase {
				t.Fatalf("result=%#v error=%v, want phase=%s error=%v", result, got, test.wantPhase, test.want)
			}
		})
	}
}
