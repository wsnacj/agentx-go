// Package toolloop owns the portable Open Tool Loop mechanism.
package toolloop

import (
	"context"
	"fmt"
)

// OutcomeKind identifies why a tool-loop driver stopped or continued.
type OutcomeKind string

const (
	OutcomeContinue   OutcomeKind = "continue"
	OutcomeCompleted  OutcomeKind = "completed"
	OutcomeTerminated OutcomeKind = "terminated"
	OutcomeMaxRounds  OutcomeKind = "max_rounds"
)

// Config controls the portable round driver.
type Config struct {
	MaxRounds int
}

// StepInput describes the current round.
type StepInput struct {
	Round     int
	MaxRounds int
}

// StepResult reports the portable outcome of one host-owned round.
type StepResult struct {
	Kind        OutcomeKind
	Termination *TerminationSignal
}

// Result reports the round and outcome that stopped the driver.
type Result struct {
	Kind  OutcomeKind
	Round int
}

// Stepper executes one host-owned model/tool round.
type Stepper interface {
	Step(context.Context, StepInput) (StepResult, error)
}

// Runtime drives host-owned rounds without owning model, tool, or policy
// implementations.
type Runtime struct {
	maxRounds int
	stepper   Stepper
}

// New validates and constructs a portable tool-loop Runtime.
func New(config Config, stepper Stepper) (*Runtime, error) {
	if config.MaxRounds <= 0 {
		return nil, fmt.Errorf("agentx tool loop: max rounds must be positive")
	}
	if stepper == nil {
		return nil, fmt.Errorf("agentx tool loop: stepper is required")
	}
	return &Runtime{
		maxRounds: config.MaxRounds,
		stepper:   stepper,
	}, nil
}

// Run executes at most MaxRounds host-owned steps. The supplied context is
// forwarded unchanged; cancellation policy remains with the host step.
func (rt *Runtime) Run(ctx context.Context) (Result, error) {
	if rt == nil || rt.stepper == nil || rt.maxRounds <= 0 {
		return Result{}, fmt.Errorf("agentx tool loop: runtime is required")
	}
	for round := 1; round <= rt.maxRounds; round++ {
		step, err := rt.stepper.Step(ctx, StepInput{
			Round:     round,
			MaxRounds: rt.maxRounds,
		})
		if err != nil {
			return Result{Round: round}, err
		}
		switch step.Kind {
		case OutcomeContinue:
			continue
		case OutcomeCompleted, OutcomeTerminated:
			return Result{
				Kind:  step.Kind,
				Round: round,
			}, nil
		default:
			return Result{Round: round}, fmt.Errorf(
				"agentx tool loop: unsupported step outcome %q",
				step.Kind,
			)
		}
	}
	return Result{
		Kind:  OutcomeMaxRounds,
		Round: rt.maxRounds,
	}, nil
}
