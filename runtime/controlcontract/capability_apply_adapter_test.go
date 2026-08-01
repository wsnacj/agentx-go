package controlcontract

import (
	"testing"
)

func TestHostOwnedCapabilityApplyRequestReady(t *testing.T) {
	request := BuildHostOwnedCapabilityApplyRequest(capabilityApplyReadyRequestInput())
	if request.Status != HostActionReady ||
		!request.ReadyForHostCapabilityApply ||
		!request.HostCapabilityApplyAuthorized ||
		!request.HostMayApplyCapabilityMutation ||
		request.Action != CapabilityApplyInstall ||
		request.FailureClass != FailureNone ||
		request.NextHostAction != "host_may_apply_capability_mutation" {
		t.Fatalf("unexpected capability apply request: %#v", request)
	}
	if request.CoreInvocationExecuted ||
		request.InstallerExecutedByCore ||
		request.InstallExecutedByCore ||
		request.EnableExecutedByCore ||
		request.PackageManagerExecutedByCore ||
		request.SkillWriteByCore ||
		request.RuntimeReloadByCore {
		t.Fatalf("core must not execute capability apply side effects: %#v", request)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "host owned capability apply request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "host_owned_capability_apply_request", "capability_install_proposal_not_apply", "no_package_manager_execution_by_core", "no_skill_write_by_core", "no_runtime_reload_by_core")
}

func TestHostOwnedCapabilityApplyResultAndReadbackRecorded(t *testing.T) {
	request := capabilityApplyReadyRequest()
	result := capabilityApplyReadyResult(request)
	if result.Status != HostActionRecorded ||
		!result.HostCapabilityApplyReported ||
		!result.HostCapabilityApplySucceeded ||
		result.HostCapabilityApplyFailed ||
		!result.HostCapabilityApplyRecorded ||
		!result.ReadyForCapabilityReadback ||
		!result.HostCapabilityInstalled ||
		result.HostCapabilityEnabled ||
		result.HostSkillApplied ||
		result.CapabilityApplyResultRef != request.ExpectedCapabilityResultRef ||
		result.AppliedCapabilityRef != request.ExpectedCapabilityRef ||
		result.AppliedCapabilityStateRef != request.ExpectedCapabilityStateRef ||
		result.NextHostAction != "bind_capability_apply_readback" {
		t.Fatalf("unexpected capability apply result: %#v", result)
	}
	readback := BuildHostOwnedCapabilityApplyReadback(HostOwnedCapabilityApplyReadbackInput{
		Result:                     result,
		CapabilityReadbackRef:      result.ExpectedReadbackRef,
		ObservedCapabilityRef:      result.AppliedCapabilityRef,
		ObservedCapabilityStateRef: result.AppliedCapabilityStateRef,
		ObservedRollbackPathRef:    result.RollbackPathRef,
		ReadbackEvidenceRefs:       []DisplaySafeRef{"evidence:capability_apply_readback"},
	})
	if readback.Status != HostActionRecorded ||
		!readback.CapabilityReadbackBound ||
		!readback.RollbackPathVerified ||
		!readback.ReadyForRuntimeLoopContinuation ||
		readback.CapabilityReadbackRef != result.ExpectedReadbackRef ||
		readback.NextHostAction != "continue_objective_runtime_loop" {
		t.Fatalf("unexpected capability apply readback: %#v", readback)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "host owned capability apply readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "host_owned_capability_apply_readback", "capability_apply_readback_bound", "rollback_path_verified", "no_install_apply_by_core")
}

func TestHostOwnedCapabilityApplyRequiresInstallerGateAndL4(t *testing.T) {
	wrongGateInput := capabilityApplyReadyRequestInput()
	wrongGateInput.IndependentGate = schedulerApplyReadyIndependentGate()
	wrongGate := BuildHostOwnedCapabilityApplyRequest(wrongGateInput)
	if wrongGate.ReadyForHostCapabilityApply ||
		wrongGate.HostMayApplyCapabilityMutation ||
		wrongGate.FailureClass != FailurePolicyBlocked ||
		!productionAdapterStringContains(wrongGate.BlockedReasons, "capability_independent_gate_not_ready") {
		t.Fatalf("expected installer gate block, got %#v", wrongGate)
	}

	l3Input := capabilityApplyReadyRequestInput()
	l3Input.FinalGate.ApprovedIntensity = IntensityL3ManagedObjective
	l3 := BuildHostOwnedCapabilityApplyRequest(l3Input)
	if l3.ReadyForHostCapabilityApply ||
		l3.HostMayApplyCapabilityMutation ||
		l3.FailureClass != FailurePolicyBlocked ||
		!productionAdapterStringContains(l3.BlockedReasons, "capability_apply_requires_l4") ||
		!productionAdapterMissingContains(l3.MissingInputs, "contract:l4_durable_long_run") {
		t.Fatalf("expected L4 gate block, got %#v", l3)
	}
}

func TestHostOwnedCapabilityApplyBlocksMissingRequiredApplyInputs(t *testing.T) {
	cases := []struct {
		name        string
		mutate      func(*HostOwnedCapabilityApplyRequestInput)
		wantMissing MissingInput
		wantReason  string
	}{
		{
			name: "host approval",
			mutate: func(input *HostOwnedCapabilityApplyRequestInput) {
				input.HostCapabilityConfirmationRef = ""
			},
			wantMissing: "host:capability_confirmation_ref",
			wantReason:  "host_capability_confirmation_ref_missing",
		},
		{
			name: "dry run proof",
			mutate: func(input *HostOwnedCapabilityApplyRequestInput) {
				input.CapabilityDryRunProofRef = ""
			},
			wantMissing: "host:capability_dry_run_proof_ref",
			wantReason:  "capability_dry_run_proof_ref_missing",
		},
		{
			name: "idempotency",
			mutate: func(input *HostOwnedCapabilityApplyRequestInput) {
				input.IdempotencyRef = ""
				input.IndependentGate.IdempotencyRef = ""
			},
			wantMissing: "host:capability_idempotency_ref",
			wantReason:  "capability_independent_gate_not_ready",
		},
		{
			name: "rollback path",
			mutate: func(input *HostOwnedCapabilityApplyRequestInput) {
				input.RollbackPathRef = ""
			},
			wantMissing: "host:capability_rollback_path_ref",
			wantReason:  "capability_rollback_path_ref_missing",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			input := capabilityApplyReadyRequestInput()
			tc.mutate(&input)
			request := BuildHostOwnedCapabilityApplyRequest(input)
			if request.ReadyForHostCapabilityApply ||
				request.HostMayApplyCapabilityMutation ||
				!productionAdapterMissingContains(request.MissingInputs, tc.wantMissing) ||
				!productionAdapterStringContains(request.BlockedReasons, tc.wantReason) {
				t.Fatalf("expected capability apply block for missing %s, got %#v", tc.name, request)
			}
			if request.CoreInvocationExecuted ||
				request.InstallerExecutedByCore ||
				request.PackageManagerExecutedByCore ||
				request.SkillWriteByCore ||
				request.RuntimeReloadByCore {
				t.Fatalf("missing input block must not execute core side effects: %#v", request)
			}
		})
	}
}

func TestHostOwnedCapabilityApplyRequestRejectsUnsafeRefWithoutLeak(t *testing.T) {
	input := capabilityApplyReadyRequestInput()
	input.CapabilityCandidateRef = "postgresql://secret@example.invalid/db"
	input.CapabilityDryRunProofRef = "/Users/mason/raw/install-output"
	request := BuildHostOwnedCapabilityApplyRequest(input)
	if request.ReadyForHostCapabilityApply ||
		request.FailureClass != FailureEvidenceWeak ||
		!productionAdapterStringContains(request.BlockedReasons, "unsafe_input_ref") ||
		!productionAdapterMissingContains(request.MissingInputs, "host:display_safe_refs") {
		t.Fatalf("expected unsafe ref block, got %#v", request)
	}
	assertNoRawPayload(t, "capability apply unsafe request", request, "postgresql://", "example.invalid", "/Users/mason/raw")
}

func TestHostOwnedCapabilityApplyReadbackBlocksMismatch(t *testing.T) {
	result := capabilityApplyReadyResult(capabilityApplyReadyRequest())
	readback := BuildHostOwnedCapabilityApplyReadback(HostOwnedCapabilityApplyReadbackInput{
		Result:                     result,
		CapabilityReadbackRef:      "readback:wrong",
		ObservedCapabilityRef:      "capability:wrong",
		ObservedCapabilityStateRef: result.AppliedCapabilityStateRef,
		ObservedRollbackPathRef:    result.RollbackPathRef,
	})
	if readback.ReadyForRuntimeLoopContinuation ||
		readback.FailureClass != FailureVerificationFailed ||
		!productionAdapterStringContains(readback.BlockedReasons, "capability_readback_ref_mismatch") ||
		!productionAdapterStringContains(readback.BlockedReasons, "observed_capability_ref_mismatch") {
		t.Fatalf("expected readback mismatch block, got %#v", readback)
	}
}

func TestHostOwnedCapabilityApplyLifecycleActionsRecorded(t *testing.T) {
	cases := []struct {
		action       CapabilityApplyAction
		targetNeeded bool
		assertFlag   func(HostOwnedCapabilityApplyResult) bool
		rejectFlags  func(HostOwnedCapabilityApplyResult) bool
	}{
		{
			action: CapabilityApplyEnable, targetNeeded: true,
			assertFlag: func(result HostOwnedCapabilityApplyResult) bool { return result.HostCapabilityEnabled },
			rejectFlags: func(result HostOwnedCapabilityApplyResult) bool {
				return result.HostCapabilityInstalled || result.HostSkillApplied || result.HostRuntimeReloaded || result.HostCapabilityRolledBack
			},
		},
		{
			action:     CapabilityApplySkillApply,
			assertFlag: func(result HostOwnedCapabilityApplyResult) bool { return result.HostSkillApplied },
			rejectFlags: func(result HostOwnedCapabilityApplyResult) bool {
				return result.HostCapabilityInstalled || result.HostCapabilityEnabled || result.HostRuntimeReloaded || result.HostCapabilityRolledBack
			},
		},
		{
			action:     CapabilityApplyAuthorize,
			assertFlag: func(result HostOwnedCapabilityApplyResult) bool { return result.HostCapabilityAuthorized },
			rejectFlags: func(result HostOwnedCapabilityApplyResult) bool {
				return result.HostCapabilityInstalled || result.HostCapabilityEnabled || result.HostSkillApplied || result.HostRuntimeReloaded
			},
		},
		{
			action: CapabilityApplyReload, targetNeeded: true,
			assertFlag: func(result HostOwnedCapabilityApplyResult) bool { return result.HostRuntimeReloaded },
			rejectFlags: func(result HostOwnedCapabilityApplyResult) bool {
				return result.HostCapabilityInstalled || result.HostCapabilityEnabled || result.HostSkillApplied || result.HostCapabilityRolledBack
			},
		},
		{
			action: CapabilityApplyRollback, targetNeeded: true,
			assertFlag: func(result HostOwnedCapabilityApplyResult) bool { return result.HostCapabilityRolledBack },
			rejectFlags: func(result HostOwnedCapabilityApplyResult) bool {
				return result.HostCapabilityInstalled || result.HostCapabilityEnabled || result.HostSkillApplied || result.HostRuntimeReloaded
			},
		},
	}
	for _, tc := range cases {
		t.Run(string(tc.action), func(t *testing.T) {
			input := capabilityApplyReadyRequestInput()
			input.Action = tc.action
			if tc.targetNeeded {
				input.TargetCapabilityRef = input.ExpectedCapabilityRef
			}
			request := BuildHostOwnedCapabilityApplyRequest(input)
			if request.Status != HostActionReady || !request.ReadyForHostCapabilityApply {
				t.Fatalf("unexpected capability apply request for %s: %#v", tc.action, request)
			}
			result := capabilityApplyReadyResult(request)
			if result.Status != HostActionRecorded || !tc.assertFlag(result) || tc.rejectFlags(result) {
				t.Fatalf("unexpected capability apply action flag for %s: %#v", tc.action, result)
			}
		})
	}
}

func capabilityApplyReadyRequestInput() HostOwnedCapabilityApplyRequestInput {
	return HostOwnedCapabilityApplyRequestInput{
		Descriptor:                    capabilityApplyReadyDescriptor(),
		IndependentGate:               capabilityApplyReadyIndependentGate(),
		FinalGate:                     capabilityApplyReadyFinalGate(),
		Action:                        CapabilityApplyInstall,
		CapabilityApplyRequestRef:     "capability_request:missing_tool_install",
		CapabilityProposalRef:         "proposal:capability_install_missing_tool",
		CapabilityCandidateRef:        "candidate:host_approved_tool",
		CapabilityGuardRef:            "guard:capability_install_missing_tool",
		CapabilityDryRunProofRef:      "dry_run:capability_install_missing_tool",
		StrategyRef:                   "strategy:capability_resolution",
		ObjectiveRunRef:               "objective_run:capability_resolution",
		ExpectedCapabilityRef:         "capability:host_approved_tool",
		ExpectedCapabilityStateRef:    "capability_state:host_approved_tool_active",
		HostCapabilityConfirmationRef: "approval:capability_apply",
		IdempotencyRef:                "idempotency:capability_apply_missing_tool",
		ExpectedCapabilityResultRef:   "capability_result:missing_tool_install",
		ExpectedReadbackRef:           "readback:capability_apply_missing_tool",
		RollbackPathRef:               "rollback:capability_apply_missing_tool",
		ApprovalRefs:                  []DisplaySafeRef{"approval:capability_apply"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:capability_install_guard",
			Kind:     "capability_guard",
			Strength: EvidenceAdequate,
		}},
	}
}

func capabilityApplyReadyRequest() HostOwnedCapabilityApplyRequest {
	return BuildHostOwnedCapabilityApplyRequest(capabilityApplyReadyRequestInput())
}

func capabilityApplyReadyResult(request HostOwnedCapabilityApplyRequest) HostOwnedCapabilityApplyResult {
	return BuildHostOwnedCapabilityApplyResult(HostOwnedCapabilityApplyResultInput{
		Request:                      request,
		CapabilityApplyResultRef:     request.ExpectedCapabilityResultRef,
		HostCapabilityRunRef:         "run:capability_apply_missing_tool",
		HostCapabilityApplyReported:  true,
		HostCapabilityApplySucceeded: true,
		AppliedCapabilityRef:         request.ExpectedCapabilityRef,
		AppliedCapabilityStateRef:    request.ExpectedCapabilityStateRef,
		CapabilityEvidenceRefs:       []DisplaySafeRef{"evidence:capability_apply_result"},
	})
}

func capabilityApplyReadyDescriptor() HostOwnedCapabilityApplyDescriptor {
	return BuildHostOwnedCapabilityApplyDescriptor(HostOwnedCapabilityApplyDescriptor{
		Available:               true,
		CapabilityDescriptorRef: "capability_descriptor:host_installer",
		CapabilityAdapterRef:    "capability_adapter:host_installer",
		OwnerRef:                "owner:host_reference",
		SupportedActions: []CapabilityApplyAction{
			CapabilityApplyInstall,
			CapabilityApplyEnable,
			CapabilityApplySkillApply,
			CapabilityApplyAuthorize,
			CapabilityApplyReload,
			CapabilityApplyRollback,
		},
		ProposalContractRef:    "contract:capability_install_proposal",
		CandidateContractRef:   "contract:capability_install_candidate",
		GuardContractRef:       "contract:capability_install_guard",
		DryRunContractRef:      "contract:capability_apply_dry_run",
		IdempotencyContractRef: "contract:capability_apply_idempotency",
		ReadbackContractRef:    "contract:capability_apply_readback",
		RollbackContractRef:    "contract:capability_apply_rollback",
		ApprovalPolicyRef:      "policy:capability_apply_host_approval",
		RedactionPolicyRef:     "policy:display_safe_refs",
		TimeoutPolicyRef:       "policy:capability_apply_timeout",
		PolicyRefs:             []DisplaySafeRef{"policy:capability_apply_host_approval"},
		RequiredApprovalRefs:   []DisplaySafeRef{"approval:capability_apply"},
	})
}

func capabilityApplyReadyIndependentGate() ProductionAdapterIndependentEffectGate {
	return BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
		Kind:                  ProductionAdapterEffectGateInstallerApply,
		GateRef:               "gate:installer_apply",
		AdapterRef:            "capability_adapter:host_installer",
		ContractRef:           "contract:capability_apply",
		PolicyRef:             "policy:capability_apply_host_approval",
		ApprovalRef:           "approval:capability_apply",
		BudgetRef:             "budget:capability_apply",
		IdempotencyRef:        "idempotency:capability_apply_missing_tool",
		ReadbackRef:           "readback:capability_apply_missing_tool",
		EvalRef:               "eval:capability_apply",
		FailureReviewRef:      "review:capability_apply_failure",
		CompensationReviewRef: "review:capability_apply_compensation",
	})
}

func capabilityApplyReadyFinalGate() IntensityGateResult {
	return IntensityGateResult{
		ContractVersion:          ContractVersion,
		Projected:                true,
		Stage:                    IntensityGateFinal,
		Activation:               ActivationManaged,
		Status:                   VerificationSatisfied,
		Allowed:                  true,
		ApprovedControlMode:      ControlModeCapabilityResolution,
		ApprovedIntensity:        IntensityL4DurableLongRun,
		StrategyRef:              "strategy:capability_resolution",
		RequiresUserConfirmation: true,
		RequiresHostApproval:     true,
		UserConfirmed:            true,
		HostApproved:             true,
		ApprovalRefs:             []DisplaySafeRef{"approval:capability_apply"},
		PolicyRefs:               []DisplaySafeRef{"policy:capability_apply_host_approval"},
		FailureClass:             FailureNone,
		Boundaries:               []Boundary{"final_gate_satisfied", "capability_l4_approved"},
		NextHostAction:           "host_may_plan_strategy",
		RunnerEffect:             "none",
		PromptEffect:             "none",
	}
}
