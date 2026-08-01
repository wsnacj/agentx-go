package controlcontract_test

import (
	"encoding/json"
	"reflect"
	"testing"

	controlcontract "github.com/wsnacj/agentx-go/runtime/controlcontract"
)

func TestExternalConsumerBuildsPortableControlProjection(t *testing.T) {
	projection := controlcontract.BuildManagedObjectiveProjection(controlcontract.ManagedObjectiveProjectionInput{
		Activation: controlcontract.ActivationManaged,
		Frame: controlcontract.ObjectiveFrame{
			ID:              "objective:example",
			UserGoalDigest:  "sha256:goal",
			SuccessCriteria: []string{"result is verified"},
		},
		LedgerRef:           "ledger:example",
		Approved:            true,
		ApprovalRefs:        []controlcontract.DisplaySafeRef{"approval:example"},
		PolicyRefs:          []controlcontract.DisplaySafeRef{"contract:intensity_gate", "contract:budget", "contract:approval_policy", "contract:strategy_scope", "contract:redaction_policy"},
		AllowedStrategyRefs: []controlcontract.DisplaySafeRef{"strategy:example"},
	})
	if !projection.Ready || projection.Status != controlcontract.HostActionReady {
		t.Fatalf("projection = %#v", projection)
	}
	if projection.NextHostAction != "host_may_plan_managed_objective" || projection.RunnerEffect != "none" || projection.PromptEffect != "none" {
		t.Fatalf("unexpected portable effects: %#v", projection)
	}

	raw, err := json.Marshal(projection)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var roundTrip controlcontract.ManagedObjectiveProjection
	if err := json.Unmarshal(raw, &roundTrip); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	roundTrip = roundTrip.Normalize()
	if !reflect.DeepEqual(projection, roundTrip) {
		t.Fatalf("round trip mismatch:\nwant %#v\n got %#v", projection, roundTrip)
	}
}

func TestExternalConsumerUsesDeterministicGates(t *testing.T) {
	budget := controlcontract.EvaluateRetryBudgetGate(controlcontract.BudgetGateInput{
		Limit:     3,
		Used:      1,
		Increment: 1,
		Scope:     "objective:example",
	})
	if !budget.Allowed || budget.RetryBudgetRemaining != 1 || budget.Status != controlcontract.VerificationSatisfied {
		t.Fatalf("budget = %#v", budget)
	}

	transition := controlcontract.CheckLifecycleTransition(controlcontract.LifecycleStageReady, controlcontract.LifecycleStageApplied)
	if !transition.Allowed || transition.Status != controlcontract.HostActionReady {
		t.Fatalf("transition = %#v", transition)
	}

	unsafe := controlcontract.VerifyDisplaySafeOnly(false, []string{"https://example.invalid/raw"})
	if unsafe.Satisfied || unsafe.Status != controlcontract.VerificationBlocked || unsafe.FailureClass != controlcontract.FailureEvidenceWeak {
		t.Fatalf("unsafe projection = %#v", unsafe)
	}
}

func TestExternalConsumerBuildsObjectiveGraphKernel(t *testing.T) {
	const (
		capability = controlcontract.DisplaySafeRef("capability:inventory")
		strategy   = controlcontract.DisplaySafeRef("strategy:inventory")
	)
	spec := controlcontract.ObjectiveSpec{
		SpecRef:               "spec:inventory",
		ObjectiveID:           "objective:inventory",
		UserGoalDigest:        "sha256:inventory",
		RawGoalRef:            "goal:inventory",
		GoalSummary:           "collect inventory evidence",
		ControlMode:           controlcontract.ControlModeObjective,
		Intensity:             controlcontract.IntensityL3ManagedObjective,
		CandidateCapabilities: []controlcontract.DisplaySafeRef{capability},
		SuccessCriteria: []controlcontract.ObjectiveSuccessCriterion{{
			CriteriaRef: "criteria:inventory",
			Text:        "inventory evidence exists",
			RequiredEvidence: []controlcontract.EvidenceRef{{
				Ref:      "evidence:inventory",
				Kind:     "inventory",
				Strength: controlcontract.EvidenceAdequate,
				Source:   "source:inventory",
			}},
		}},
		RequiredEvidence: []controlcontract.EvidenceRef{{
			Ref:      "evidence:inventory",
			Kind:     "inventory",
			Strength: controlcontract.EvidenceAdequate,
			Source:   "source:inventory",
		}},
		SideEffectPolicy:  controlcontract.ObjectiveSpecSideEffectReadOnly,
		MissingInfoPolicy: controlcontract.ObjectiveSpecMissingInfoAskUser,
		Budget: controlcontract.ObjectiveSpecBudget{
			BudgetRef:   "budget:inventory",
			MaxNodes:    1,
			MaxAttempts: 1,
		},
		PolicyRefs: []controlcontract.DisplaySafeRef{"policy:inventory"},
	}
	catalog := controlcontract.StrategyCatalogSnapshot{
		CatalogRef: "catalog:inventory",
		Entries: []controlcontract.StrategyCatalogEntry{{
			SourceKind: controlcontract.StrategyCatalogSourceHostAdapter,
			SourceRef:  "source:inventory",
			Status:     controlcontract.VerificationSatisfied,
			Candidate: controlcontract.StrategyCandidate{
				ID:              string(strategy),
				Kind:            "host_adapter",
				ControlMode:     controlcontract.ControlModeObjective,
				MinIntensity:    controlcontract.IntensityL3ManagedObjective,
				CapabilityRefs:  []controlcontract.DisplaySafeRef{capability},
				SideEffectClass: string(controlcontract.ObjectiveCapabilitySideEffectReadOnly),
				ExpectedEvidence: []controlcontract.EvidenceRef{{
					Ref:      "evidence:inventory",
					Kind:     "inventory",
					Strength: controlcontract.EvidenceAdequate,
					Source:   "source:inventory",
				}},
				Owner: "host",
			},
		}},
	}
	graph := controlcontract.ObjectiveGraph{
		GraphRef:   "graph:inventory",
		CatalogRef: catalog.CatalogRef,
		Nodes: []controlcontract.ObjectiveNode{{
			NodeRef:             "node:inventory",
			Kind:                "host_adapter",
			CapabilityRef:       capability,
			StrategyRef:         strategy,
			DescriptorRef:       "descriptor:inventory",
			SourceRef:           "source:inventory",
			InputSchemaRef:      "schema:inventory.input.v1",
			OutputSchemaRef:     "schema:inventory.output.v1",
			EvidenceContractRef: "evidence:inventory.contract.v1",
			RequiredEvidence:    spec.RequiredEvidence,
			AttemptPolicy: controlcontract.ObjectiveNodeAttemptPolicy{
				MaxAttempts:    1,
				TimeoutSeconds: 30,
				NoProgressGate: true,
			},
			SideEffectClass: controlcontract.ObjectiveCapabilitySideEffectReadOnly,
			PolicyRefs:      []controlcontract.DisplaySafeRef{"policy:inventory"},
		}},
	}
	result := controlcontract.BuildObjectiveGraphValidation(controlcontract.ObjectiveGraphValidationInput{
		Graph:   graph,
		Spec:    spec,
		Catalog: catalog,
		Policy: controlcontract.ExecutionIntensityPolicy{
			PolicyRef:            "policy:inventory",
			ExecutionContractRef: "contract:inventory",
			Activation:           controlcontract.ActivationManaged,
			DefaultIntensity:     controlcontract.IntensityL3ManagedObjective,
			MaxDefaultIntensity:  controlcontract.IntensityL3ManagedObjective,
			MaxAllowedIntensity:  controlcontract.IntensityL3ManagedObjective,
			AllowedControlModesByIntensity: map[controlcontract.ExecutionIntensity][]controlcontract.ControlMode{
				controlcontract.IntensityL3ManagedObjective: {controlcontract.ControlModeObjective},
			},
			PolicyRefs: []controlcontract.DisplaySafeRef{"policy:inventory"},
		},
	})
	if !result.Validated || !result.ReadyForRuntimeLoop || result.Status != controlcontract.VerificationSatisfied || result.ReadyNodeCount != 1 {
		t.Fatalf("graph validation = %#v", result)
	}
}
