package toolloop

import (
	"context"
	"fmt"
)

// AssemblyConfig combines the portable driver and round coordinator inputs.
// Concrete model, tool, persistence, and product policy remain in Coordinator
// ports supplied by the host.
type AssemblyConfig struct {
	MaxRounds   int
	Coordinator CoordinatorConfig
	Initial     RoundState
}

// AssemblyResult reports the driver outcome, final portable state, and any
// termination fact observed by the coordinator.
type AssemblyResult struct {
	Driver      Result
	State       RoundState
	Termination *TerminationSignal
}

// Assembly is a single-run composition of Runtime and Coordinator. It is
// stateful, intended for one logical Run, and is not safe for concurrent use.
type Assembly struct {
	driver      *Runtime
	coordinator *Coordinator
	stepper     *terminationCapturingStepper
}

// NewAssembly validates and assembles the portable driver and coordinator.
func NewAssembly(config AssemblyConfig) (*Assembly, error) {
	coordinator, err := NewCoordinator(config.Coordinator, config.Initial)
	if err != nil {
		return nil, err
	}
	stepper := &terminationCapturingStepper{coordinator: coordinator}
	driver, err := New(Config{MaxRounds: config.MaxRounds}, stepper)
	if err != nil {
		return nil, err
	}
	return &Assembly{
		driver:      driver,
		coordinator: coordinator,
		stepper:     stepper,
	}, nil
}

// Run executes the assembled driver. The caller context and round errors are
// forwarded without wrapping. Returned state and termination are defensive
// copies of the assembly's portable result.
func (assembly *Assembly) Run(ctx context.Context) (AssemblyResult, error) {
	if assembly == nil || assembly.driver == nil || assembly.coordinator == nil || assembly.stepper == nil {
		return AssemblyResult{}, fmt.Errorf("agentx tool loop: assembly runtime is required")
	}
	assembly.stepper.termination = nil
	driverResult, err := assembly.driver.Run(ctx)
	return AssemblyResult{
		Driver:      driverResult,
		State:       assembly.coordinator.State(),
		Termination: cloneTerminationSignal(assembly.stepper.termination),
	}, err
}

type terminationCapturingStepper struct {
	coordinator *Coordinator
	termination *TerminationSignal
}

func (stepper *terminationCapturingStepper) Step(ctx context.Context, input StepInput) (StepResult, error) {
	result, err := stepper.coordinator.Step(ctx, input)
	stepper.termination = cloneTerminationSignal(result.Termination)
	return result, err
}

func cloneTerminationSignal(in *TerminationSignal) *TerminationSignal {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
