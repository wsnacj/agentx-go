package hostkit

import (
	"context"
	"fmt"

	objective "github.com/wsnacj/agentx-go/runtime/objective"
)

// Config contains only host-owned adapter capabilities. It does not install a
// provider, backend, credential, approval, policy, or product default.
type Config struct {
	Handler  Handler
	Handlers map[objective.DisplaySafeRef]Handler
}

// RunRequest contains display-safe Objective inputs and explicit dispatch
// authorization for one Host Kit invocation.
type RunRequest struct {
	Ingress               objective.ManagedObjectiveIngressInput
	DispatchEnabled       bool
	DispatchHostConfirmed bool
	Boundaries            []objective.Boundary
}

// RunResult preserves both the side-effect-free ingress projection and the
// optional host-owned dispatch readback.
type RunResult struct {
	Ingress            objective.ManagedObjectiveIngressResult `json:"ingress,omitempty"`
	Dispatch           DispatchResult                          `json:"dispatch,omitempty"`
	ReadyForDispatch   bool                                    `json:"ready_for_dispatch"`
	DispatchAttempted  bool                                    `json:"dispatch_attempted"`
	Completed          bool                                    `json:"completed"`
	Status             string                                  `json:"status,omitempty"`
	FailureClass       string                                  `json:"failure_class,omitempty"`
	NextHostAction     string                                  `json:"next_host_action,omitempty"`
	MissingInputs      []string                                `json:"missing_inputs,omitempty"`
	Boundaries         []string                                `json:"boundaries,omitempty"`
	RawOutputLoaded    bool                                    `json:"raw_output_loaded"`
	RunnerDispatched   bool                                    `json:"runner_dispatched"`
	ToolExecuted       bool                                    `json:"tool_executed"`
	WorkflowDispatched bool                                    `json:"workflow_dispatched"`
	SchedulerApplied   bool                                    `json:"scheduler_applied"`
	InstallerExecuted  bool                                    `json:"installer_executed"`
}

// Runtime is an immutable Objective composition. Handler concurrency safety
// remains the responsibility of the Host.
type Runtime struct {
	handler  Handler
	handlers map[objective.DisplaySafeRef]Handler
}

// New validates the host-owned adapter capabilities and creates an immutable
// Objective Runtime.
func New(config Config) (*Runtime, error) {
	handlers := make(map[objective.DisplaySafeRef]Handler, len(config.Handlers))
	for ref, handler := range config.Handlers {
		if handler == nil {
			continue
		}
		normalized, ok := objective.NormalizeDisplaySafeRef(string(ref))
		if !ok || normalized == "" {
			return nil, fmt.Errorf("agentx objective host kit: handler ref must be display-safe")
		}
		handlers[normalized] = handler
	}
	if config.Handler == nil && len(handlers) == 0 {
		return nil, fmt.Errorf("agentx objective host kit: handler is required")
	}
	return &Runtime{handler: config.Handler, handlers: handlers}, nil
}

// Run projects one managed Objective and, only after explicit enable and host
// confirmation, invokes exactly one selected handler before verification.
func (runtime *Runtime) Run(ctx context.Context, request RunRequest) RunResult {
	if runtime == nil {
		return RunResult{
			Status:         "blocked",
			FailureClass:   string(objective.FailureConfigMissing),
			NextHostAction: "construct_objective_host_kit",
			MissingInputs:  []string{"host:objective_host_kit"},
			Boundaries:     []string{"objective_host_kit_missing"},
		}
	}
	ingress := objective.BuildManagedIngress(request.Ingress).Normalize()
	result := RunResult{
		Ingress:          ingress,
		ReadyForDispatch: ingress.ReadyForRuntimeAdapter,
		Status:           string(ingress.Status),
		FailureClass:     string(ingress.FailureClass),
		NextHostAction:   string(ingress.NextHostAction),
		MissingInputs:    missingInputsToStrings(ingress.MissingInputs),
		Boundaries:       boundariesToStrings(ingress.Boundaries),
		RawOutputLoaded:  ingress.RawOutputLoaded,
	}
	if !ingress.ReadyForRuntimeAdapter {
		return result
	}
	dispatch := Dispatch(DispatchInput{
		Enabled:                  request.DispatchEnabled,
		HostConfirmed:            request.DispatchHostConfirmed,
		Request:                  ingress.RuntimeAdapterRequest,
		Handler:                  runtime.handler,
		Handlers:                 runtime.handlers,
		ExpectedObservationKinds: request.Ingress.ExpectedObservationKinds,
		Boundaries: objective.AppendBoundaries(request.Ingress.Boundaries,
			request.Boundaries...,
		),
		Context: ctx,
	})
	result.Dispatch = dispatch
	result.DispatchAttempted = dispatch.HandlerReady
	result.Completed = dispatch.Satisfied
	result.Status = dispatch.Status
	result.FailureClass = dispatch.FailureClass
	result.NextHostAction = dispatch.NextHostAction
	result.MissingInputs = appendUniqueStrings(result.MissingInputs, dispatch.MissingInputs...)
	result.Boundaries = appendUniqueStrings(result.Boundaries, dispatch.Boundaries...)
	result.RawOutputLoaded = result.RawOutputLoaded || dispatch.RawOutputLoaded
	result.RunnerDispatched = dispatch.RunnerDispatched
	result.ToolExecuted = dispatch.ToolExecuted
	result.WorkflowDispatched = dispatch.WorkflowDispatched
	result.SchedulerApplied = dispatch.SchedulerApplied
	result.InstallerExecuted = dispatch.InstallerExecuted
	return result
}
