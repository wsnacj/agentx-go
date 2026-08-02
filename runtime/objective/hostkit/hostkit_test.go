package hostkit

import (
	"context"
	"errors"
	"testing"

	objective "github.com/wsnacj/agentx-go/runtime/objective"
)

func TestNewRequiresHandler(t *testing.T) {
	_, err := New(Config{})
	if err == nil || err.Error() != "agentx objective host kit: handler is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRuntimeRunExecutesExactlyOnceAndVerifies(t *testing.T) {
	input := readyIngressInput()
	called := 0
	runtime, err := New(Config{Handlers: map[objective.DisplaySafeRef]Handler{
		"adapter:test_objective": func(ctx context.Context, request objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult {
			called++
			if ctx == nil || request.AdapterRef != "adapter:test_objective" {
				t.Fatalf("unexpected handler request: %#v", request)
			}
			return satisfiedResult(request)
		},
	}})
	if err != nil {
		t.Fatal(err)
	}

	result := runtime.Run(context.Background(), RunRequest{
		Ingress:               input,
		DispatchEnabled:       true,
		DispatchHostConfirmed: true,
	})
	if called != 1 || !result.ReadyForDispatch || !result.DispatchAttempted || !result.Completed || result.Status != "satisfied" {
		t.Fatalf("unexpected run result called=%d result=%#v", called, result)
	}
	if result.RunnerDispatched || result.ToolExecuted || result.WorkflowDispatched || result.SchedulerApplied || result.InstallerExecuted {
		t.Fatalf("host kit must not claim core side effects: %#v", result)
	}
}

func TestRuntimeRunRequiresExplicitConfirmation(t *testing.T) {
	called := 0
	runtime, err := New(Config{Handler: func(context.Context, objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult {
		called++
		return objective.RuntimeAdapterResult{}
	}})
	if err != nil {
		t.Fatal(err)
	}

	result := runtime.Run(context.Background(), RunRequest{
		Ingress:         readyIngressInput(),
		DispatchEnabled: true,
	})
	if called != 0 || result.DispatchAttempted || result.Completed || result.NextHostAction != "request_runtime_adapter_dispatch_confirmation" {
		t.Fatalf("unexpected confirmation gate result called=%d result=%#v", called, result)
	}
}

func TestRuntimeRunPassesCancellationToHost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var observed error
	runtime, err := New(Config{Handler: func(ctx context.Context, request objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult {
		observed = ctx.Err()
		return objective.BuildRuntimeAdapterResult(objective.RuntimeAdapterResultInput{
			Request:           request,
			AdapterRef:        request.AdapterRef,
			StrategyRef:       request.StrategyRef,
			HostAdapterRunRef: "adapter_run:cancelled",
			Status:            objective.VerificationBlocked,
			FailureReason:     "cancelled",
		})
	}})
	if err != nil {
		t.Fatal(err)
	}

	result := runtime.Run(ctx, RunRequest{
		Ingress:               readyIngressInput(),
		DispatchEnabled:       true,
		DispatchHostConfirmed: true,
	})
	if !errors.Is(observed, context.Canceled) || !result.DispatchAttempted || result.Completed {
		t.Fatalf("cancellation not preserved observed=%v result=%#v", observed, result)
	}
}

func TestDispatchMissingHandlerFailsClosed(t *testing.T) {
	ingress := objective.BuildManagedIngress(readyIngressInput())
	result := Dispatch(DispatchInput{
		Enabled:       true,
		HostConfirmed: true,
		Request:       ingress.RuntimeAdapterRequest,
	})
	if result.HandlerReady || result.RuntimeAdapterExecuted || result.NextHostAction != "provide_runtime_adapter_dispatch_handler" || !contains(result.MissingInputs, "host:runtime_adapter_dispatch_handler") {
		t.Fatalf("unexpected missing-handler result: %#v", result)
	}
}

func readyIngressInput() objective.ManagedObjectiveIngressInput {
	evidence := objective.EvidenceRef{
		Ref:      "evidence:test_objective",
		Kind:     "objective_answer",
		Strength: objective.EvidenceAdequate,
		Source:   "adapter_run:test_objective",
	}
	strategy := objective.StrategyCandidate{
		ID:               "strategy:test_objective",
		Kind:             "host_declared_strategy",
		ControlMode:      objective.ControlModeObjective,
		MinIntensity:     objective.IntensityL3ManagedObjective,
		MaxIntensity:     objective.IntensityL3ManagedObjective,
		CapabilityRefs:   []objective.DisplaySafeRef{"capability:test_objective"},
		ExpectedEvidence: []objective.EvidenceRef{evidence},
		Risk:             "read_only",
		SideEffectClass:  "read_only",
		RequiresApproval: true,
		Owner:            "host",
	}
	descriptor := objective.ProductionAdapterDescriptor{
		AdapterRef:             "adapter:test_objective",
		Owner:                  "host",
		OwnerRef:               "host:test_objective",
		Version:                "v1",
		Kind:                   objective.ProductionAdapterSourceReadback,
		SupportedCandidateRefs: []objective.DisplaySafeRef{objective.DisplaySafeRef(strategy.ID)},
		RequiresCapabilityRefs: strategy.CapabilityRefs,
		InputContractRef:       "contract:test_objective_input",
		OutputContractRef:      "contract:test_objective_output",
		ReadbackContractRef:    "contract:test_objective_readback",
		RequiredPolicyRefs:     []objective.DisplaySafeRef{"policy:test_objective"},
		RequiredApprovalRefs:   []objective.DisplaySafeRef{"approval:test_objective"},
		RequiredBudgetRef:      "budget:test_objective",
		IdempotencyContractRef: "contract:test_objective_idempotency",
		SideEffectClass:        "read_only",
		DisplaySafeInputRefs:   []objective.DisplaySafeRef{"input:test_objective"},
		DisplaySafeOutputRefs:  []objective.DisplaySafeRef{"output:test_objective"},
	}
	return objective.ManagedObjectiveIngressInput{
		Activation:       objective.ActivationManaged,
		ObjectiveID:      "objective:test_objective",
		UserGoalDigest:   "sha256:test_objective",
		SourceRef:        "host:test_objective",
		SuccessCriteria:  []string{"produce verified objective evidence"},
		RequiredEvidence: []objective.EvidenceRef{evidence},
		LedgerRef:        "ledger:test_objective",
		Policy: objective.ExecutionIntensityPolicy{
			PolicyRef:            "policy:test_objective",
			ExecutionContractRef: "contract:test_objective",
			Activation:           objective.ActivationManaged,
			DefaultIntensity:     objective.IntensityL3ManagedObjective,
			MaxDefaultIntensity:  objective.IntensityL3ManagedObjective,
			MaxAllowedIntensity:  objective.IntensityL3ManagedObjective,
			AllowedControlModesByIntensity: map[objective.ExecutionIntensity][]objective.ControlMode{
				objective.IntensityL3ManagedObjective: {objective.ControlModeObjective},
			},
			DeniedSideEffectsByIntensity: map[objective.ExecutionIntensity][]string{
				objective.IntensityL3ManagedObjective: {"external_write", "schedule_apply", "install"},
			},
			PolicyRefs: []objective.DisplaySafeRef{
				"contract:intensity_gate",
				"contract:budget",
				"contract:approval_policy",
				"contract:strategy_scope",
				"contract:redaction_policy",
				"policy:test_objective",
			},
		},
		Budget:              objective.ObjectiveBudgetSnapshot{BudgetRef: "budget:test_objective", Limit: 3, Remaining: 3},
		UserConfirmed:       true,
		HostApproved:        true,
		ApprovalRefs:        []objective.DisplaySafeRef{"approval:test_objective"},
		AllowedStrategyRefs: []objective.DisplaySafeRef{objective.DisplaySafeRef(strategy.ID)},
		StrategyCatalog: objective.StrategyCatalogSnapshot{
			CatalogRef: "catalog:test_objective",
			Entries: []objective.StrategyCatalogEntry{{
				SourceKind: objective.StrategyCatalogSourceHostAdapter,
				SourceRef:  "host:test_objective",
				Candidate:  strategy,
				Status:     objective.VerificationSatisfied,
			}},
			PolicyRefs: []objective.DisplaySafeRef{"policy:test_objective"},
		}.Normalize(),
		AvailableCapabilityRefs: strategy.CapabilityRefs,
		AdapterRegistry: objective.BuildHostAdapterRegistry(objective.HostAdapterRegistryInput{
			RegistryRef: "registry:test_objective",
			Descriptors: []objective.ProductionAdapterDescriptor{descriptor},
			PolicyRefs:  []objective.DisplaySafeRef{"policy:test_objective", "contract:test_objective"},
		}),
		RequestedAdapterRef:      descriptor.AdapterRef,
		IdempotencyRef:           "idempotency:test_objective",
		RuntimeInputRefs:         descriptor.DisplaySafeInputRefs,
		ExpectedObservationKinds: []string{"objective_answer"},
		PolicyRefs: []objective.DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
			"policy:test_objective",
		},
	}
}

func satisfiedResult(request objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult {
	evidence := objective.EvidenceRef{
		Ref:      "evidence:test_objective",
		Kind:     "objective_answer",
		Strength: objective.EvidenceAdequate,
		Source:   "adapter_run:test_objective",
	}
	return objective.BuildRuntimeAdapterResult(objective.RuntimeAdapterResultInput{
		Request:           request,
		AdapterRef:        request.AdapterRef,
		StrategyRef:       request.StrategyRef,
		HostAdapterRunRef: "adapter_run:test_objective",
		Status:            objective.VerificationSatisfied,
		Observations: []objective.Observation{{
			Kind:            "objective_answer",
			Source:          "adapter_run:test_objective",
			Subject:         "objective:test_objective",
			Strength:        objective.EvidenceAdequate,
			EvidenceRefs:    []objective.EvidenceRef{evidence},
			DisplaySafeRefs: []objective.DisplaySafeRef{evidence.Ref},
		}},
		EvidenceRefs: []objective.EvidenceRef{evidence},
		OutputRefs:   []objective.DisplaySafeRef{"output:test_objective"},
	})
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
