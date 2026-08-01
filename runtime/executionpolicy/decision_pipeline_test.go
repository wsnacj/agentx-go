package executionpolicy

import "testing"

func TestDecisionActionFromRuntimeDecisionPrecedence(t *testing.T) {
	cases := []struct {
		name     string
		decision RuntimeDecision
		want     DecisionAction
	}{
		{
			name:     "unchecked",
			decision: RuntimeDecision{},
			want:     DecisionActionNone,
		},
		{
			name: "allowed",
			decision: RuntimeDecision{
				Checked: true,
				Allowed: true,
			},
			want: DecisionActionAllow,
		},
		{
			name: "prompt",
			decision: RuntimeDecision{
				Checked:         true,
				RequiresConfirm: true,
			},
			want: DecisionActionPrompt,
		},
		{
			name: "forbidden wins",
			decision: RuntimeDecision{
				Checked:         true,
				Allowed:         true,
				Denied:          true,
				RequiresConfirm: true,
			},
			want: DecisionActionForbidden,
		},
		{
			name: "degraded allow",
			decision: RuntimeDecision{
				Checked:  true,
				Degraded: true,
			},
			want: DecisionActionAllow,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DecisionActionFromRuntimeDecision(tc.decision); got != tc.want {
				t.Fatalf("unexpected decision action: got %q want %q", got, tc.want)
			}
		})
	}
}

func TestDecisionPacketClassifiesRuntimeDecisionsAndSummarizes(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		ContractID: " contract-1 ",
		Source:     " runtime_preflight ",
		Subject:    " tool_call ",
		RuntimeDecisions: []RuntimeDecision{
			{
				Checked:            true,
				Tool:               "exec",
				Action:             "exec",
				Allowed:            true,
				Reason:             "exec_allowed",
				DecisionSubject:    "command",
				TargetKind:         "command",
				PolicySource:       RuntimePolicySourceRuntimeGuard,
				ControlSource:      RuntimeControlSourceExecutionContract,
				EnforcementSurface: RuntimeEnforcementSurfaceRuntimePreflight,
			},
			{
				Checked:            true,
				Tool:               "http_request",
				Action:             "request",
				Denied:             true,
				Reason:             "url_runtime_guard_denied",
				DecisionSubject:    "http_request_url",
				TargetKind:         "url",
				PolicySource:       RuntimePolicySourceRuntimeGuard,
				ControlSource:      RuntimeControlSourceExecutionContract,
				EnforcementSurface: RuntimeEnforcementSurfaceRuntimePreflight,
			},
			{
				Checked:            true,
				Tool:               "exec",
				Action:             "exec",
				Allowed:            true,
				Reason:             "runtime_approval_allowed",
				PolicySource:       RuntimePolicySourceApprovalHook,
				ControlSource:      RuntimeControlSourceExecutionContract,
				EnforcementSurface: RuntimeEnforcementSurfaceRuntimePreflight,
			},
		},
		Steps: []DecisionStep{
			{
				Kind:          DecisionStepKindRetry,
				Action:        DecisionActionPrompt,
				Reason:        "sandbox_escalation_requires_approval",
				PolicySource:  RuntimePolicySourceToolCallPolicy,
				ControlSource: RuntimeControlSourceExecutionContract,
			},
		},
	})

	if packet.SchemaVersion != DecisionPacketSchemaV1 ||
		packet.ContractID != "contract-1" ||
		packet.Source != "runtime_preflight" ||
		packet.Subject != "tool_call" {
		t.Fatalf("unexpected packet envelope: %#v", packet)
	}
	if packet.FinalAction != DecisionActionForbidden {
		t.Fatalf("expected forbidden final action, got %#v", packet.FinalAction)
	}
	if len(packet.Steps) != 4 {
		t.Fatalf("expected four decision steps, got %#v", packet.Steps)
	}
	if packet.Steps[0].Kind != DecisionStepKindSandbox ||
		packet.Steps[1].Kind != DecisionStepKindNetwork ||
		packet.Steps[2].Kind != DecisionStepKindApproval ||
		packet.Steps[3].Kind != DecisionStepKindRetry {
		t.Fatalf("unexpected step classification: %#v", packet.Steps)
	}
	if packet.Summary.Steps != 4 ||
		packet.Summary.Allowed != 2 ||
		packet.Summary.Prompted != 1 ||
		packet.Summary.Forbidden != 1 ||
		packet.Summary.ByKind[string(DecisionStepKindSandbox)] != 1 ||
		packet.Summary.ByKind[string(DecisionStepKindNetwork)] != 1 ||
		packet.Summary.ByKind[string(DecisionStepKindApproval)] != 1 ||
		packet.Summary.ByKind[string(DecisionStepKindRetry)] != 1 ||
		packet.Summary.ByPolicySource[RuntimePolicySourceRuntimeGuard] != 2 {
		t.Fatalf("unexpected decision summary: %#v", packet.Summary)
	}
	if packet.Audit == nil ||
		packet.Audit.ContractID != "contract-1" ||
		packet.Audit.FinalAction != DecisionActionForbidden ||
		len(packet.Audit.Reasons) != 4 ||
		len(packet.Audit.PolicySources) != 3 ||
		len(packet.Audit.ControlSources) != 1 ||
		len(packet.Audit.Tools) != 2 {
		t.Fatalf("unexpected audit record: %#v", packet.Audit)
	}
}

func TestDecisionPacketOmitsUncheckedRuntimeDecision(t *testing.T) {
	packet := NewDecisionPacket(DecisionPacketInput{
		RuntimeDecisions: []RuntimeDecision{
			{Tool: "exec", Action: "exec", Allowed: true},
		},
	})
	if len(packet.Steps) != 0 ||
		packet.FinalAction != DecisionActionNone ||
		packet.Audit != nil ||
		packet.Summary.Steps != 0 {
		t.Fatalf("unchecked runtime decision must not produce packet facts, got %#v", packet)
	}
}
