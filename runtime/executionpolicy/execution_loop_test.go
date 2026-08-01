package executionpolicy

import "testing"

func TestBuildExecutionLoopReportAllowsHostExecutionForAllowPacket(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-execution-loop-allow",
		Source:     "authorization",
		Subject:    "tool_call:exec",
		RuntimeDecisions: []RuntimeDecision{
			{
				Checked:            true,
				Tool:               "exec",
				Action:             "exec",
				Allowed:            true,
				Reason:             "command_allowed",
				DecisionSubject:    "command",
				TargetKind:         "command",
				PolicySource:       RuntimePolicySourceRuntimeGuard,
				ControlSource:      RuntimeControlSourceExecutionContract,
				EnforcementSurface: RuntimeEnforcementSurfaceRuntimePreflight,
			},
		},
	})

	report := BuildExecutionLoopReport(ExecutionLoopReportInput{
		Enabled:        true,
		DecisionPacket: packet,
	})

	if report.Status != ExecutionLoopStatusReadyForHostExecution ||
		!report.Available ||
		!report.ExecutionAllowed ||
		!report.ReadyForHostExecution ||
		report.ReadyForRetry ||
		report.ExecutionBlocked ||
		report.DecisionPacketFinalAction != DecisionActionAllow ||
		report.ExecutionLoopRef == "" ||
		report.DecisionPacketRef == "" ||
		report.Contract.Owner != "execution" ||
		!report.Contract.ObservationOnly ||
		report.Contract.PolicySource ||
		report.Contract.AuthorizationSource ||
		report.Contract.RuntimeInvocationSource {
		t.Fatalf("unexpected allow execution loop report: %#v", report)
	}
	for _, boundary := range []string{
		"execution_loop_report",
		"decision_packet_execution_loop_evidence",
		"execution_loop_does_not_invoke_tools",
		"execution_loop_does_not_call_host_adapter",
		"execution_loop_ready_for_host_execution",
		"no_core_execution",
	} {
		if !executionLoopTestHasString(report.Boundaries, boundary) {
			t.Fatalf("missing boundary %q in %#v", boundary, report.Boundaries)
		}
	}
}

func TestBuildExecutionLoopReportBlocksDeniedReadUntilReviewAndRetryRefs(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-execution-loop-denied-read",
		Source:     "authorization",
		Subject:    "tool_call:read",
		Steps: []DecisionStep{
			{
				Kind:            DecisionStepKindRuntime,
				Action:          DecisionActionForbidden,
				Tool:            "read",
				RuntimeAction:   "read",
				Reason:          "read_policy_denied",
				DecisionSubject: "file_read",
				TargetKind:      "path",
				PolicySource:    RuntimePolicySourceToolCallPolicy,
				ControlSource:   RuntimeControlSourceExecutionContract,
			},
		},
	})

	blocked := BuildExecutionLoopReport(ExecutionLoopReportInput{
		Enabled:        true,
		DecisionPacket: packet,
	})
	if blocked.Status != ExecutionLoopStatusBlocked ||
		!blocked.ExecutionBlocked ||
		!blocked.RequiresDeniedReadReview ||
		blocked.ReadyForDeniedReadReview ||
		blocked.ReadyForRetry ||
		blocked.ReadyForHostExecution ||
		!executionLoopTestHasString(blocked.MissingInputs, "host:execution_loop_denied_read_review_ref") ||
		!executionLoopTestHasString(blocked.MissingInputs, "host:execution_loop_retry_request_ref") {
		t.Fatalf("denied read should require review and retry refs, got %#v", blocked)
	}
	if len(blocked.Actions) != 1 ||
		blocked.Actions[0].Kind != ExecutionLoopActionDeniedReadReview ||
		blocked.Actions[0].Ready {
		t.Fatalf("unexpected denied read action: %#v", blocked.Actions)
	}

	ready := BuildExecutionLoopReport(ExecutionLoopReportInput{
		Enabled:             true,
		DecisionPacket:      packet,
		DeniedReadReviewRef: "review:denied_read",
		RetryRequestRef:     "retry:denied_read",
	})
	if ready.Status != ExecutionLoopStatusReadyForRetry ||
		!ready.Available ||
		!ready.ReadyForRetry ||
		ready.ReadyForHostExecution ||
		!ready.ReadyForDeniedReadReview ||
		!ready.RetryRequired ||
		ready.NextHostAction != "retry_with_new_execution_decision_packet" {
		t.Fatalf("denied read review should prepare retry only, got %#v", ready)
	}
	if len(ready.Actions) != 1 ||
		ready.Actions[0].Kind != ExecutionLoopActionDeniedReadReview ||
		!ready.Actions[0].Ready ||
		ready.Actions[0].ActionRef != "review:denied_read" {
		t.Fatalf("unexpected ready denied read action: %#v", ready.Actions)
	}
}

func TestBuildExecutionLoopReportPreparesSandboxNetworkApprovalRetry(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-execution-loop-retry",
		Source:     "authorization",
		Subject:    "tool_call:exec",
		Steps: []DecisionStep{
			{
				Kind:            DecisionStepKindSandbox,
				Action:          DecisionActionForbidden,
				Tool:            "exec",
				RuntimeAction:   "exec",
				Reason:          "command_runtime_guard_denied",
				DecisionSubject: "command",
				TargetKind:      "command",
				PolicySource:    RuntimePolicySourceRuntimeGuard,
				ControlSource:   RuntimeControlSourceExecutionContract,
			},
			{
				Kind:            DecisionStepKindNetwork,
				Action:          DecisionActionForbidden,
				Tool:            "http_request",
				RuntimeAction:   "request",
				Reason:          "url_runtime_guard_denied",
				DecisionSubject: "http_request_url",
				TargetKind:      "url",
				PolicySource:    RuntimePolicySourceRuntimeGuard,
				ControlSource:   RuntimeControlSourceExecutionContract,
			},
			{
				Kind:            DecisionStepKindApproval,
				Action:          DecisionActionPrompt,
				Tool:            "exec",
				RuntimeAction:   "exec",
				Reason:          "runtime_approval_required",
				DecisionSubject: "tool_call:exec",
				TargetKind:      "tool_call",
				PolicySource:    RuntimePolicySourceApprovalHook,
				ControlSource:   RuntimeControlSourceExecutionContract,
			},
		},
	})

	report := BuildExecutionLoopReport(ExecutionLoopReportInput{
		Enabled:              true,
		DecisionPacket:       packet,
		SandboxEscalationRef: "sandbox_escalation:exec",
		NetworkApprovalRef:   "network_approval:http_request",
		HostApprovalRef:      "approval:exec",
		RetryRequestRef:      "retry:exec",
	})

	if report.Status != ExecutionLoopStatusReadyForRetry ||
		!report.ReadyForRetry ||
		report.ReadyForHostExecution ||
		!report.RequiresSandboxEscalation ||
		!report.ReadyForSandboxEscalation ||
		!report.RequiresNetworkApproval ||
		!report.ReadyForNetworkApproval ||
		!report.RequiresHostApproval ||
		!report.ReadyForHostApproval ||
		len(report.Actions) != 3 ||
		len(report.MissingInputs) != 0 ||
		len(report.BlockedReasons) != 0 {
		t.Fatalf("unexpected retry-ready report: %#v", report)
	}
	for _, kind := range []string{
		ExecutionLoopActionSandboxEscalation,
		ExecutionLoopActionNetworkApproval,
		ExecutionLoopActionHostApproval,
	} {
		if !executionLoopTestHasActionKind(report.Actions, kind) {
			t.Fatalf("missing action %q in %#v", kind, report.Actions)
		}
		if !executionLoopTestHasString(report.Boundaries, "execution_loop_action_"+kind) {
			t.Fatalf("missing action boundary for %q in %#v", kind, report.Boundaries)
		}
	}
}

func TestBuildExecutionLoopReportRejectsUnsafeRefs(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-execution-loop-unsafe",
		Source:     "authorization",
		Subject:    "tool_call:exec",
		RuntimeDecisions: []RuntimeDecision{
			{
				Checked:       true,
				Tool:          "exec",
				Action:        "exec",
				Allowed:       true,
				Reason:        "command_allowed",
				PolicySource:  RuntimePolicySourceToolCallPolicy,
				ControlSource: RuntimeControlSourceExecutionContract,
			},
		},
	})

	report := BuildExecutionLoopReport(ExecutionLoopReportInput{
		Enabled:          true,
		ExecutionLoopRef: "/tmp/raw-loop",
		DecisionPacket:   packet,
	})

	if !report.RawOutputLoaded ||
		report.Available ||
		report.ReadyForHostExecution ||
		report.Status != ExecutionLoopStatusBlocked ||
		!executionLoopTestHasString(report.MissingInputs, "host:display_safe_refs") ||
		!executionLoopTestHasString(report.BlockedReasons, "unsafe_execution_loop_ref") ||
		!executionLoopTestHasString(report.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe refs should fail closed, got %#v", report)
	}
}

func TestBuildExecutionLoopReportDefaultOffDoesNotPrepareLoop(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: "contract-execution-loop-default-off",
		Source:     "authorization",
		Subject:    "tool_call:exec",
		RuntimeDecisions: []RuntimeDecision{
			{
				Checked:       true,
				Tool:          "exec",
				Action:        "exec",
				Allowed:       true,
				Reason:        "command_allowed",
				PolicySource:  RuntimePolicySourceToolCallPolicy,
				ControlSource: RuntimeControlSourceExecutionContract,
			},
		},
	})

	report := BuildExecutionLoopReport(ExecutionLoopReportInput{
		DecisionPacket: packet,
	})

	if report.Available ||
		report.ReadyForHostExecution ||
		report.Status != ExecutionLoopStatusBlocked ||
		report.NextHostAction != "enable_execution_loop_report" ||
		!executionLoopTestHasString(report.MissingInputs, "host:execution_loop_enabled") ||
		!executionLoopTestHasString(report.Boundaries, "execution_loop_default_off") {
		t.Fatalf("default-off loop should stay blocked, got %#v", report)
	}
}

func executionLoopTestHasString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func executionLoopTestHasActionKind(actions []ExecutionLoopAction, want string) bool {
	for _, action := range actions {
		if action.Kind == want && action.Ready {
			return true
		}
	}
	return false
}
