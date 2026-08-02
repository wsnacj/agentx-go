package main

import (
	"context"
	"fmt"

	objective "github.com/wsnacj/agentx-go/runtime/objective"
	objectivehostkit "github.com/wsnacj/agentx-go/runtime/objective/hostkit"
)

func run(ctx context.Context) (string, error) {
	runtime, err := objectivehostkit.New(objectivehostkit.Config{
		Handlers: map[objective.DisplaySafeRef]objectivehostkit.Handler{
			"adapter:conformance": execute,
		},
	})
	if err != nil {
		return "", err
	}
	result := runtime.Run(ctx, objectivehostkit.RunRequest{
		Ingress:               ingress(),
		DispatchEnabled:       true,
		DispatchHostConfirmed: true,
	})
	if !result.Completed {
		return "", fmt.Errorf("objective blocked: status=%s failure=%s next=%s", result.Status, result.FailureClass, result.NextHostAction)
	}
	return fmt.Sprintf("agentx-objective-hostkit-ok:%s:%s:%t", result.Status, result.Dispatch.AdapterRef, result.Dispatch.HostExecutionReported), nil
}

func execute(ctx context.Context, request objective.RuntimeAdapterRequest) objective.RuntimeAdapterResult {
	if err := ctx.Err(); err != nil {
		return objective.BuildRuntimeAdapterResult(objective.RuntimeAdapterResultInput{
			Request:           request,
			AdapterRef:        request.AdapterRef,
			StrategyRef:       request.StrategyRef,
			HostAdapterRunRef: "adapter_run:conformance_cancelled",
			Status:            objective.VerificationBlocked,
			FailureReason:     "context cancelled",
		})
	}
	evidence := objective.EvidenceRef{Ref: "evidence:conformance", Kind: "objective_answer", Strength: objective.EvidenceAdequate, Source: "adapter_run:conformance"}
	return objective.BuildRuntimeAdapterResult(objective.RuntimeAdapterResultInput{
		Request:           request,
		AdapterRef:        request.AdapterRef,
		StrategyRef:       request.StrategyRef,
		HostAdapterRunRef: "adapter_run:conformance",
		Status:            objective.VerificationSatisfied,
		Observations: []objective.Observation{{
			Kind:            "objective_answer",
			Source:          "adapter_run:conformance",
			Subject:         "objective:conformance",
			Strength:        objective.EvidenceAdequate,
			EvidenceRefs:    []objective.EvidenceRef{evidence},
			DisplaySafeRefs: []objective.DisplaySafeRef{evidence.Ref},
		}},
		EvidenceRefs: []objective.EvidenceRef{evidence},
		OutputRefs:   []objective.DisplaySafeRef{"output:conformance"},
	})
}

func ingress() objective.ManagedObjectiveIngressInput {
	strategyRef := objective.DisplaySafeRef("strategy:conformance")
	evidence := objective.EvidenceRef{Ref: "evidence:conformance", Kind: "objective_answer", Strength: objective.EvidenceAdequate, Source: "adapter_run:conformance"}
	strategy := objective.StrategyCandidate{
		ID:               string(strategyRef),
		Kind:             "host_declared_strategy",
		ControlMode:      objective.ControlModeObjective,
		MinIntensity:     objective.IntensityL3ManagedObjective,
		MaxIntensity:     objective.IntensityL3ManagedObjective,
		ExpectedEvidence: []objective.EvidenceRef{evidence},
		RequiresApproval: true,
		Owner:            "host",
	}
	descriptor := objective.ProductionAdapterDescriptor{
		AdapterRef:             "adapter:conformance",
		Owner:                  "host",
		OwnerRef:               "host:conformance",
		Version:                "v1",
		Kind:                   objective.ProductionAdapterSourceReadback,
		SupportedCandidateRefs: []objective.DisplaySafeRef{strategyRef},
		InputContractRef:       "contract:conformance_input",
		OutputContractRef:      "contract:conformance_output",
		ReadbackContractRef:    "contract:conformance_readback",
		RequiredPolicyRefs:     []objective.DisplaySafeRef{"policy:conformance"},
		RequiredApprovalRefs:   []objective.DisplaySafeRef{"approval:conformance"},
		RequiredBudgetRef:      "budget:conformance",
		IdempotencyContractRef: "contract:conformance_idempotency",
		SideEffectClass:        "read_only",
		DisplaySafeInputRefs:   []objective.DisplaySafeRef{"input:conformance"},
		DisplaySafeOutputRefs:  []objective.DisplaySafeRef{"output:conformance"},
	}
	contractRefs := []objective.DisplaySafeRef{
		"contract:intensity_gate",
		"contract:budget",
		"contract:approval_policy",
		"contract:strategy_scope",
		"contract:redaction_policy",
		"policy:conformance",
	}
	return objective.ManagedObjectiveIngressInput{
		Activation:       objective.ActivationManaged,
		ObjectiveID:      "objective:conformance",
		UserGoalDigest:   "sha256:conformance",
		SourceRef:        "host:conformance",
		SuccessCriteria:  []string{"produce verified conformance evidence"},
		RequiredEvidence: []objective.EvidenceRef{evidence},
		LedgerRef:        "ledger:conformance",
		Policy: objective.ExecutionIntensityPolicy{
			PolicyRef:            "policy:conformance",
			ExecutionContractRef: "contract:conformance",
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
			PolicyRefs: contractRefs,
		},
		Budget:              objective.ObjectiveBudgetSnapshot{BudgetRef: "budget:conformance", Limit: 1, Remaining: 1},
		UserConfirmed:       true,
		HostApproved:        true,
		ApprovalRefs:        []objective.DisplaySafeRef{"approval:conformance"},
		AllowedStrategyRefs: []objective.DisplaySafeRef{strategyRef},
		StrategyCatalog: objective.StrategyCatalogSnapshot{
			CatalogRef: "catalog:conformance",
			Entries: []objective.StrategyCatalogEntry{{
				SourceKind: objective.StrategyCatalogSourceHostAdapter,
				SourceRef:  "host:conformance",
				Candidate:  strategy,
				Status:     objective.VerificationSatisfied,
			}},
			PolicyRefs: []objective.DisplaySafeRef{"policy:conformance"},
		}.Normalize(),
		AdapterRegistry: objective.BuildHostAdapterRegistry(objective.HostAdapterRegistryInput{
			RegistryRef: "registry:conformance",
			Descriptors: []objective.ProductionAdapterDescriptor{descriptor},
			PolicyRefs:  []objective.DisplaySafeRef{"policy:conformance", "contract:conformance"},
		}),
		RequestedAdapterRef:      descriptor.AdapterRef,
		IdempotencyRef:           "idempotency:conformance",
		RuntimeInputRefs:         descriptor.DisplaySafeInputRefs,
		ExpectedObservationKinds: []string{"objective_answer"},
		PolicyRefs:               contractRefs,
	}
}

func main() {
	output, err := run(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(output)
}
