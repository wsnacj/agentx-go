package controlcontract

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHostOwnedSchedulerApplyRequestReady(t *testing.T) {
	request := schedulerApplyReadyRequest()
	if request.Status != HostActionReady ||
		!request.ReadyForHostSchedulerApply ||
		!request.HostSchedulerApplyAuthorized ||
		!request.HostMayApplySchedulerMutation ||
		request.Action != SchedulerApplyCreate ||
		request.ScheduleProposalRef == "" ||
		request.ScheduleDryRunProofRef == "" ||
		request.HostSchedulerConfirmationRef == "" ||
		request.IdempotencyRef == "" ||
		request.CancelPathRef == "" ||
		request.DeletePathRef == "" ||
		request.DisablePathRef == "" ||
		request.CoreInvocationExecuted ||
		request.SchedulerMutationByCore ||
		request.CoreScheduleCreated ||
		request.AutomationCreatedByCore ||
		request.NextHostAction != "host_may_apply_scheduler_mutation" {
		t.Fatalf("unexpected scheduler apply request: %#v", request)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "host owned scheduler apply request",
		RunnerEffect: request.RunnerEffect,
		PromptEffect: request.PromptEffect,
		Boundaries:   request.Boundaries,
		Payload:      request,
	}, "host_owned_scheduler_apply_request", "explicit_host_approval_required", "l4_scheduler_apply_required", "schedule_dry_run_proof_required", "core_scheduler_apply_not_executed")
}

func TestHostOwnedSchedulerApplyResultAndReadbackRecorded(t *testing.T) {
	request := schedulerApplyReadyRequest()
	result := schedulerApplyReadyResult(request)
	if result.Status != HostActionRecorded ||
		!result.HostSchedulerApplyReported ||
		!result.HostSchedulerApplySucceeded ||
		result.HostSchedulerApplyFailed ||
		!result.HostSchedulerApplyRecorded ||
		!result.ReadyForSchedulerReadback ||
		!result.HostScheduleCreated ||
		result.SchedulerApplyResultRef != request.ExpectedSchedulerResultRef ||
		result.AppliedScheduleRef != request.ExpectedScheduleRef ||
		result.AppliedLifecycleStateRef != request.ExpectedLifecycleStateRef ||
		result.CoreInvocationExecuted ||
		result.SchedulerMutationByCore ||
		result.CoreScheduleCreated ||
		result.AutomationCreatedByCore ||
		result.NextHostAction != "bind_scheduler_apply_readback" {
		t.Fatalf("unexpected scheduler apply result: %#v", result)
	}

	readback := BuildHostOwnedSchedulerApplyReadback(HostOwnedSchedulerApplyReadbackInput{
		SchedulerReadbackRef:      result.ExpectedReadbackRef,
		Result:                    result,
		ObservedScheduleRef:       result.AppliedScheduleRef,
		ObservedLifecycleStateRef: result.AppliedLifecycleStateRef,
		ObservedCancelPathRef:     result.CancelPathRef,
		ObservedDeletePathRef:     result.DeletePathRef,
		ObservedDisablePathRef:    result.DisablePathRef,
		ReadbackEvidenceRefs:      []DisplaySafeRef{"evidence:scheduler_readback"},
	})
	if readback.Status != HostActionRecorded ||
		!readback.SchedulerReadbackBound ||
		!readback.LifecyclePathVerified ||
		!readback.ReadyForRuntimeLoopContinuation ||
		readback.SchedulerApplyResultRef != result.SchedulerApplyResultRef ||
		readback.ObservedScheduleRef != result.AppliedScheduleRef ||
		readback.ObservedLifecycleStateRef != result.AppliedLifecycleStateRef ||
		readback.CoreInvocationExecuted ||
		readback.SchedulerMutationByCore ||
		readback.CoreScheduleCreated ||
		readback.AutomationCreatedByCore ||
		readback.NextHostAction != "continue_objective_runtime_loop" {
		t.Fatalf("unexpected scheduler apply readback: %#v", readback)
	}
	assertHostOwnedProjectionOnly(t, testProjection[Boundary]{
		Name:         "host owned scheduler apply readback",
		RunnerEffect: readback.RunnerEffect,
		PromptEffect: readback.PromptEffect,
		Boundaries:   readback.Boundaries,
		Payload:      readback,
	}, "host_owned_scheduler_apply_readback", "scheduler_apply_readback_bound", "cancel_delete_disable_path_verified", "ready_for_runtime_loop_continuation")
}

func TestHostOwnedSchedulerApplyRequiresL4GateAndLifecyclePaths(t *testing.T) {
	l3Input := schedulerApplyReadyRequestInput()
	l3Input.FinalGate.ApprovedIntensity = IntensityL3ManagedObjective
	l3 := BuildHostOwnedSchedulerApplyRequest(l3Input)
	if l3.Status != HostActionBlocked ||
		l3.ReadyForHostSchedulerApply ||
		l3.HostMayApplySchedulerMutation ||
		l3.FailureClass != FailurePolicyBlocked ||
		!controlTokenListContains(l3.BlockedReasons, "scheduler_apply_requires_l4") {
		t.Fatalf("expected L4 gate block, got %#v", l3)
	}

	missingPathInput := schedulerApplyReadyRequestInput()
	missingPathInput.CancelPathRef = ""
	missingPath := BuildHostOwnedSchedulerApplyRequest(missingPathInput)
	if missingPath.Status != HostActionBlocked ||
		missingPath.ReadyForHostSchedulerApply ||
		missingPath.HostMayApplySchedulerMutation ||
		!missingInputContains(missingPath.MissingInputs, "host:scheduler_cancel_path_ref") ||
		!controlTokenListContains(missingPath.BlockedReasons, "scheduler_cancel_path_ref_missing") {
		t.Fatalf("expected cancel path block, got %#v", missingPath)
	}
}

func TestHostOwnedSchedulerApplyRequestRejectsUnsafeRefWithoutLeak(t *testing.T) {
	input := schedulerApplyReadyRequestInput()
	rawRef := "https://example.invalid/raw/schedule-proposal"
	input.ScheduleProposalRef = DisplaySafeRef(rawRef)
	request := BuildHostOwnedSchedulerApplyRequest(input)
	if request.Status != HostActionBlocked ||
		request.ReadyForHostSchedulerApply ||
		request.HostMayApplySchedulerMutation ||
		request.FailureClass != FailureEvidenceWeak ||
		request.RawOutputLoaded ||
		request.ScheduleProposalRef != "" ||
		!missingInputContains(request.MissingInputs, "host:display_safe_refs") ||
		!controlTokenListContains(request.BlockedReasons, "unsafe_input_ref") {
		t.Fatalf("expected unsafe-ref block without raw output retention, got %#v", request)
	}
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	if strings.Contains(string(payload), rawRef) {
		t.Fatalf("request leaked unsafe ref %q in %s", rawRef, payload)
	}
}

func TestHostOwnedSchedulerApplyReadbackBlocksMismatch(t *testing.T) {
	result := schedulerApplyReadyResult(schedulerApplyReadyRequest())
	readback := BuildHostOwnedSchedulerApplyReadback(HostOwnedSchedulerApplyReadbackInput{
		SchedulerReadbackRef:      result.ExpectedReadbackRef,
		Result:                    result,
		ObservedScheduleRef:       result.AppliedScheduleRef,
		ObservedLifecycleStateRef: "schedule_state:wrong",
		ObservedCancelPathRef:     result.CancelPathRef,
		ObservedDeletePathRef:     result.DeletePathRef,
		ObservedDisablePathRef:    result.DisablePathRef,
	})
	if readback.Status != HostActionBlocked ||
		readback.SchedulerReadbackBound ||
		readback.LifecyclePathVerified ||
		readback.ReadyForRuntimeLoopContinuation ||
		readback.FailureClass != FailureVerificationFailed ||
		!controlTokenListContains(readback.BlockedReasons, "scheduler_observed_lifecycle_state_ref_mismatch") {
		t.Fatalf("expected lifecycle mismatch block, got %#v", readback)
	}
}

func TestHostOwnedSchedulerApplyLifecycleActionsRecorded(t *testing.T) {
	cases := []struct {
		name       string
		action     SchedulerApplyAction
		stateRef   DisplaySafeRef
		assertFlag func(HostOwnedSchedulerApplyResult) bool
	}{
		{
			name:     "delete",
			action:   SchedulerApplyDelete,
			stateRef: "schedule_state:agentx_release_notes_weekday_9_deleted",
			assertFlag: func(result HostOwnedSchedulerApplyResult) bool {
				return result.HostScheduleDeleted &&
					!result.HostScheduleCreated &&
					!result.HostScheduleUpdated &&
					!result.HostScheduleDisabled &&
					!result.HostScheduleCanceled
			},
		},
		{
			name:     "disable",
			action:   SchedulerApplyDisable,
			stateRef: "schedule_state:agentx_release_notes_weekday_9_disabled",
			assertFlag: func(result HostOwnedSchedulerApplyResult) bool {
				return result.HostScheduleDisabled &&
					!result.HostScheduleCreated &&
					!result.HostScheduleUpdated &&
					!result.HostScheduleDeleted &&
					!result.HostScheduleCanceled
			},
		},
		{
			name:     "cancel",
			action:   SchedulerApplyCancel,
			stateRef: "schedule_state:agentx_release_notes_weekday_9_canceled",
			assertFlag: func(result HostOwnedSchedulerApplyResult) bool {
				return result.HostScheduleCanceled &&
					!result.HostScheduleCreated &&
					!result.HostScheduleUpdated &&
					!result.HostScheduleDeleted &&
					!result.HostScheduleDisabled
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			input := schedulerApplyReadyRequestInput()
			input.Action = tt.action
			input.TargetScheduleRef = input.ExpectedScheduleRef
			input.ExpectedLifecycleStateRef = tt.stateRef
			request := BuildHostOwnedSchedulerApplyRequest(input)
			if request.Status != HostActionReady ||
				!request.ReadyForHostSchedulerApply ||
				request.Action != tt.action ||
				request.TargetScheduleRef != request.ExpectedScheduleRef ||
				request.CoreInvocationExecuted ||
				request.SchedulerMutationByCore ||
				request.CoreScheduleCreated ||
				request.AutomationCreatedByCore {
				t.Fatalf("unexpected lifecycle scheduler apply request: %#v", request)
			}
			result := schedulerApplyReadyResult(request)
			if result.Status != HostActionRecorded ||
				!result.ReadyForSchedulerReadback ||
				!tt.assertFlag(result) ||
				result.CoreInvocationExecuted ||
				result.SchedulerMutationByCore ||
				result.CoreScheduleCreated ||
				result.AutomationCreatedByCore {
				t.Fatalf("unexpected lifecycle scheduler apply result: %#v", result)
			}
			readback := BuildHostOwnedSchedulerApplyReadback(HostOwnedSchedulerApplyReadbackInput{
				SchedulerReadbackRef:      result.ExpectedReadbackRef,
				Result:                    result,
				ObservedScheduleRef:       result.AppliedScheduleRef,
				ObservedLifecycleStateRef: result.AppliedLifecycleStateRef,
				ObservedCancelPathRef:     result.CancelPathRef,
				ObservedDeletePathRef:     result.DeletePathRef,
				ObservedDisablePathRef:    result.DisablePathRef,
				ReadbackEvidenceRefs:      []DisplaySafeRef{"evidence:scheduler_lifecycle_readback"},
			})
			if readback.Status != HostActionRecorded ||
				!readback.SchedulerReadbackBound ||
				!readback.LifecyclePathVerified ||
				!readback.ReadyForRuntimeLoopContinuation ||
				readback.CoreInvocationExecuted ||
				readback.SchedulerMutationByCore ||
				readback.CoreScheduleCreated ||
				readback.AutomationCreatedByCore {
				t.Fatalf("unexpected lifecycle scheduler apply readback: %#v", readback)
			}
		})
	}
}

func schedulerApplyReadyRequestInput() HostOwnedSchedulerApplyRequestInput {
	return HostOwnedSchedulerApplyRequestInput{
		Descriptor:                   schedulerApplyReadyDescriptor(),
		IndependentGate:              schedulerApplyReadyIndependentGate(),
		FinalGate:                    schedulerApplyReadyFinalGate(),
		Action:                       SchedulerApplyCreate,
		SchedulerApplyRequestRef:     "scheduler_request:agentx_release_notes_weekday_9",
		ScheduleProposalRef:          "schedule_proposal:agentx_release_notes_weekday_9",
		ScheduleDryRunProofRef:       "dry_run:scheduler_apply_agentx_release_notes_weekday_9",
		StrategyRef:                  "strategy:operations_schedule_review",
		ObjectiveRunRef:              "objective_run:operations_schedule_review",
		ExpectedScheduleRef:          "schedule:agentx_release_notes_weekday_9",
		HostSchedulerConfirmationRef: "approval:scheduler_apply",
		IdempotencyRef:               "idempotency:scheduler_apply_agentx_release_notes_weekday_9",
		ExpectedSchedulerResultRef:   "scheduler_result:agentx_release_notes_weekday_9",
		ExpectedLifecycleStateRef:    "schedule_state:agentx_release_notes_weekday_9_active",
		ExpectedReadbackRef:          "readback:scheduler_apply_agentx_release_notes_weekday_9",
		CancelPathRef:                "cancel:scheduler_apply_agentx_release_notes_weekday_9",
		DeletePathRef:                "delete:scheduler_apply_agentx_release_notes_weekday_9",
		DisablePathRef:               "disable:scheduler_apply_agentx_release_notes_weekday_9",
		ApprovalRefs:                 []DisplaySafeRef{"approval:scheduler_apply"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:schedule_dry_run",
			Kind:     "schedule_dry_run",
			Strength: EvidenceAdequate,
		}},
	}
}

func schedulerApplyReadyRequest() HostOwnedSchedulerApplyRequest {
	return BuildHostOwnedSchedulerApplyRequest(schedulerApplyReadyRequestInput())
}

func schedulerApplyReadyResult(request HostOwnedSchedulerApplyRequest) HostOwnedSchedulerApplyResult {
	return BuildHostOwnedSchedulerApplyResult(HostOwnedSchedulerApplyResultInput{
		Request:                     request,
		SchedulerApplyResultRef:     request.ExpectedSchedulerResultRef,
		HostSchedulerRunRef:         "run:scheduler_apply_agentx_release_notes_weekday_9",
		HostSchedulerApplyReported:  true,
		HostSchedulerApplySucceeded: true,
		AppliedScheduleRef:          request.ExpectedScheduleRef,
		AppliedLifecycleStateRef:    request.ExpectedLifecycleStateRef,
		SchedulerEvidenceRefs:       []DisplaySafeRef{"evidence:scheduler_apply_result"},
	})
}

func schedulerApplyReadyDescriptor() HostOwnedSchedulerApplyDescriptor {
	return BuildHostOwnedSchedulerApplyDescriptor(HostOwnedSchedulerApplyDescriptor{
		Available:              true,
		SchedulerDescriptorRef: "scheduler_descriptor:operations_host",
		SchedulerAdapterRef:    "scheduler_adapter:operations_host",
		OwnerRef:               "owner:host_reference",
		SupportedActions: []SchedulerApplyAction{
			SchedulerApplyCreate,
			SchedulerApplyUpdate,
			SchedulerApplyDelete,
			SchedulerApplyDisable,
			SchedulerApplyCancel,
		},
		ScheduleContractRef:     "contract:scheduler_schedule",
		IdempotencyContractRef:  "contract:scheduler_idempotency",
		ReadbackContractRef:     "contract:scheduler_readback",
		CancellationContractRef: "contract:scheduler_cancellation",
		DeleteContractRef:       "contract:scheduler_delete",
		DisableContractRef:      "contract:scheduler_disable",
		ApprovalPolicyRef:       "policy:scheduler_host_approval",
		RedactionPolicyRef:      "policy:display_safe_refs",
		TimeoutPolicyRef:        "policy:scheduler_apply_timeout",
		PolicyRefs:              []DisplaySafeRef{"policy:scheduler_host_approval"},
		RequiredApprovalRefs:    []DisplaySafeRef{"approval:scheduler_apply"},
	})
}

func schedulerApplyReadyIndependentGate() ProductionAdapterIndependentEffectGate {
	return BuildProductionAdapterIndependentEffectGate(ProductionAdapterIndependentEffectGateSpec{
		Kind:                  ProductionAdapterEffectGateSchedulerApply,
		GateRef:               "gate:scheduler_apply",
		AdapterRef:            "scheduler_adapter:operations_host",
		ContractRef:           "contract:scheduler_apply",
		PolicyRef:             "policy:scheduler_host_approval",
		ApprovalRef:           "approval:scheduler_apply",
		BudgetRef:             "budget:scheduler_apply",
		IdempotencyRef:        "idempotency:scheduler_apply_agentx_release_notes_weekday_9",
		ReadbackRef:           "readback:scheduler_apply_agentx_release_notes_weekday_9",
		EvalRef:               "eval:scheduler_apply",
		FailureReviewRef:      "review:scheduler_apply_failure",
		CompensationReviewRef: "review:scheduler_apply_compensation",
	})
}

func schedulerApplyReadyFinalGate() IntensityGateResult {
	return IntensityGateResult{
		ContractVersion:          ContractVersion,
		Projected:                true,
		Stage:                    IntensityGateFinal,
		Activation:               ActivationManaged,
		Status:                   VerificationSatisfied,
		Allowed:                  true,
		ApprovedControlMode:      ControlModeOperations,
		ApprovedIntensity:        IntensityL4DurableLongRun,
		StrategyRef:              "strategy:operations_schedule_review",
		RequiresUserConfirmation: true,
		RequiresHostApproval:     true,
		UserConfirmed:            true,
		HostApproved:             true,
		ApprovalRefs:             []DisplaySafeRef{"approval:scheduler_apply"},
		PolicyRefs:               []DisplaySafeRef{"policy:scheduler_host_approval"},
		FailureClass:             FailureNone,
		Boundaries:               []Boundary{"final_gate_satisfied", "scheduler_l4_approved"},
		NextHostAction:           "host_may_plan_strategy",
		RunnerEffect:             "none",
		PromptEffect:             "none",
	}
}
