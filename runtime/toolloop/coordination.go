package toolloop

import (
	"context"
	"fmt"
	"strings"
)

// RoundState is the portable state carried between host-owned rounds.
type RoundState struct {
	Chunks           []string
	ForceNoToolCalls bool
	FinalReply       string
}

// RoundExecutionInput describes the host-owned round to execute.
type RoundExecutionInput struct {
	Round     int
	MaxRounds int
	State     RoundState
}

// RoundContinuation contains the portable observations needed to continue.
type RoundContinuation struct {
	Calls            []Call
	Runs             []RunObservation
	Failures         []FailureObservation
	NextChunks       []string
	ForceNoToolCalls bool
}

// RoundExecutionResult is the portable result of one host-owned round.
type RoundExecutionResult struct {
	Kind         OutcomeKind
	Reply        string
	Continuation *RoundContinuation
}

// RoundExecutor executes the concrete model/tool work for one round.
type RoundExecutor interface {
	ExecuteRound(context.Context, RoundExecutionInput) (RoundExecutionResult, error)
}

// ContinuationObservation is supplied to an optional host policy between the
// canonical failure fuse and loop detector.
type ContinuationObservation struct {
	Round int
	Calls []Call
	Runs  []RunObservation
}

// ContinuationPolicy retains host-specific stop policy outside the portable
// coordinator. A non-empty code identifies the host policy decision.
type ContinuationPolicy interface {
	ObserveContinuation(context.Context, ContinuationObservation) (code string, stop bool, err error)
}

// TerminationKind identifies which coordination layer stopped continuation.
type TerminationKind string

const (
	TerminationFailureFuse TerminationKind = "failure_fuse"
	TerminationHostPolicy  TerminationKind = "host_policy"
	TerminationLoop        TerminationKind = "loop"
)

// TerminationSignal is a portable stop fact. User-visible projection remains
// host-owned.
type TerminationSignal struct {
	Kind    TerminationKind
	Code    string
	Failure FailureSignal
	Loop    LoopSignal
}

// CoordinatorConfig supplies the host executor and already-resolved policy
// mechanisms.
type CoordinatorConfig struct {
	Executor           RoundExecutor
	LoopDetector       *LoopDetector
	FailureFuse        *FailureFuse
	ContinuationPolicy ContinuationPolicy
}

// Coordinator owns portable round-result state application and stop ordering.
type Coordinator struct {
	config CoordinatorConfig
	state  RoundState
}

// NewCoordinator constructs a coordinator with a defensive copy of state.
func NewCoordinator(config CoordinatorConfig, initial RoundState) (*Coordinator, error) {
	if config.Executor == nil {
		return nil, fmt.Errorf("agentx tool loop: round executor is required")
	}
	return &Coordinator{
		config: config,
		state:  cloneRoundState(initial),
	}, nil
}

// Step executes and coordinates one host-owned round.
func (c *Coordinator) Step(ctx context.Context, input StepInput) (StepResult, error) {
	if c == nil || c.config.Executor == nil {
		return StepResult{}, fmt.Errorf("agentx tool loop: round coordinator is required")
	}
	result, err := c.config.Executor.ExecuteRound(ctx, RoundExecutionInput{
		Round:     input.Round,
		MaxRounds: input.MaxRounds,
		State:     cloneRoundState(c.state),
	})
	if err != nil {
		return StepResult{}, err
	}
	if strings.TrimSpace(result.Reply) != "" {
		c.state.FinalReply = result.Reply
	}
	switch result.Kind {
	case OutcomeCompleted, OutcomeTerminated:
		return StepResult{Kind: result.Kind}, nil
	case OutcomeContinue:
		// Continue below after applying the mandatory continuation.
	default:
		return StepResult{}, fmt.Errorf(
			"agentx tool loop: unsupported round outcome %q",
			result.Kind,
		)
	}
	continuation := result.Continuation
	if continuation == nil {
		return StepResult{}, fmt.Errorf("agentx: open tool loop round continuation is required")
	}
	c.state.Chunks = append([]string(nil), continuation.NextChunks...)
	c.state.ForceNoToolCalls = continuation.ForceNoToolCalls

	if c.config.FailureFuse != nil {
		if signal, stop := c.config.FailureFuse.Observe(input.Round, continuation.Failures); stop {
			return StepResult{
				Kind: OutcomeTerminated,
				Termination: &TerminationSignal{
					Kind:    TerminationFailureFuse,
					Failure: signal,
				},
			}, nil
		}
	}
	observation := ContinuationObservation{
		Round: input.Round,
		Calls: append([]Call(nil), continuation.Calls...),
		Runs:  append([]RunObservation(nil), continuation.Runs...),
	}
	if c.config.ContinuationPolicy != nil {
		code, stop, err := c.config.ContinuationPolicy.ObserveContinuation(ctx, observation)
		if err != nil {
			return StepResult{}, err
		}
		if stop {
			return StepResult{
				Kind: OutcomeTerminated,
				Termination: &TerminationSignal{
					Kind: TerminationHostPolicy,
					Code: strings.TrimSpace(code),
				},
			}, nil
		}
	}
	if c.config.LoopDetector != nil {
		if signal, stop := c.config.LoopDetector.Observe(input.Round, continuation.Calls, continuation.Runs); stop {
			return StepResult{
				Kind: OutcomeTerminated,
				Termination: &TerminationSignal{
					Kind: TerminationLoop,
					Loop: signal,
				},
			}, nil
		}
	}
	return StepResult{Kind: OutcomeContinue}, nil
}

// State returns a defensive copy of the current portable state.
func (c *Coordinator) State() RoundState {
	if c == nil {
		return RoundState{}
	}
	return cloneRoundState(c.state)
}

func cloneRoundState(state RoundState) RoundState {
	state.Chunks = append([]string(nil), state.Chunks...)
	return state
}
