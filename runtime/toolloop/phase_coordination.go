package toolloop

import (
	"context"
	"fmt"
)

// RoundPhase identifies a portable round execution phase.
type RoundPhase string

const (
	// RoundPhaseRequest identifies the host model-request phase.
	RoundPhaseRequest RoundPhase = "request"
	// RoundPhaseObserve identifies the host response-observation phase.
	RoundPhaseObserve RoundPhase = "observe"
	// RoundPhaseBeforeAction identifies the optional host action-admission phase.
	RoundPhaseBeforeAction RoundPhase = "before_action"
	// RoundPhaseAct identifies the concrete host action phase.
	RoundPhaseAct RoundPhase = "act"
)

// RoundPhaseInput identifies the host-owned round being coordinated.
type RoundPhaseInput struct {
	Round     int
	MaxRounds int
}

// RoundRequestResult reports whether the model response requires host action.
type RoundRequestResult struct {
	ActionRequired bool
}

// RoundObserveResult carries the raw assistant reply observed by the host.
type RoundObserveResult struct {
	Reply string
}

// RoundPhaseExecutor owns every concrete phase operation and policy decision.
// Implementations may retain request state between methods and must not be
// shared across concurrent round executions unless they provide synchronization.
type RoundPhaseExecutor interface {
	Request(context.Context, RoundPhaseInput) (RoundRequestResult, error)
	Observe(context.Context, RoundPhaseInput) (RoundObserveResult, error)
	BeforeAction(context.Context, RoundPhaseInput) (proceed bool, err error)
	Act(context.Context, RoundPhaseInput) error
}

// RoundPhaseResultKind identifies how portable phase orchestration completed.
type RoundPhaseResultKind string

const (
	// RoundPhaseNoAction means observation completed without requiring action.
	RoundPhaseNoAction RoundPhaseResultKind = "no_action"
	// RoundPhaseHostStopped means the host gate stopped before concrete action.
	RoundPhaseHostStopped RoundPhaseResultKind = "host_stopped"
	// RoundPhaseActionCompleted means the concrete host action completed.
	RoundPhaseActionCompleted RoundPhaseResultKind = "action_completed"
)

// RoundPhaseResult reports the portable outcome and last attempted phase.
type RoundPhaseResult struct {
	Kind      RoundPhaseResultKind
	LastPhase RoundPhase
	Reply     string
}

// RoundPhaseCoordinator owns portable request/observe/gate/act ordering.
type RoundPhaseCoordinator struct {
	executor RoundPhaseExecutor
}

// NewRoundPhaseCoordinator constructs a round phase coordinator.
func NewRoundPhaseCoordinator(executor RoundPhaseExecutor) (*RoundPhaseCoordinator, error) {
	if executor == nil {
		return nil, fmt.Errorf("agentx tool loop: round phase executor is required")
	}
	return &RoundPhaseCoordinator{executor: executor}, nil
}

// Execute coordinates one host-owned round without wrapping phase errors.
func (c *RoundPhaseCoordinator) Execute(ctx context.Context, input RoundPhaseInput) (RoundPhaseResult, error) {
	if c == nil || c.executor == nil {
		return RoundPhaseResult{}, fmt.Errorf("agentx tool loop: round phase coordinator is required")
	}
	requested, err := c.executor.Request(ctx, input)
	if err != nil {
		return RoundPhaseResult{LastPhase: RoundPhaseRequest}, err
	}
	observed, err := c.executor.Observe(ctx, input)
	if err != nil {
		return RoundPhaseResult{LastPhase: RoundPhaseObserve}, err
	}
	if !requested.ActionRequired {
		return RoundPhaseResult{
			Kind:      RoundPhaseNoAction,
			LastPhase: RoundPhaseObserve,
			Reply:     observed.Reply,
		}, nil
	}
	proceed, err := c.executor.BeforeAction(ctx, input)
	if err != nil {
		return RoundPhaseResult{LastPhase: RoundPhaseBeforeAction, Reply: observed.Reply}, err
	}
	if !proceed {
		return RoundPhaseResult{
			Kind:      RoundPhaseHostStopped,
			LastPhase: RoundPhaseBeforeAction,
			Reply:     observed.Reply,
		}, nil
	}
	if err := c.executor.Act(ctx, input); err != nil {
		return RoundPhaseResult{LastPhase: RoundPhaseAct, Reply: observed.Reply}, err
	}
	return RoundPhaseResult{
		Kind:      RoundPhaseActionCompleted,
		LastPhase: RoundPhaseAct,
		Reply:     observed.Reply,
	}, nil
}
