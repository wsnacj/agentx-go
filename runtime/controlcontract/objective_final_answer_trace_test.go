package controlcontract

import (
	"context"
	"errors"
	"testing"
)

func TestBuildObjectiveFinalAnswerTraceBackedSatisfiedWithSynthesizer(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:source_summary",
		Kind:     "summary",
		Strength: EvidenceStrong,
		Source:   "capability:public_source",
	}}
	synthesizer := ObjectiveFinalAnswerSynthesizerFunc(func(_ context.Context, request ObjectiveFinalAnswerSynthesisRequest) (ObjectiveFinalAnswerSynthesisResponse, error) {
		if request.Trace.Status != VerificationSatisfied || len(request.Trace.EvidenceRefs) != 1 {
			t.Fatalf("synthesizer did not receive trace-backed evidence: %#v", request.Trace)
		}
		if request.Trace.EvidenceRefs[0].Ref != "evidence:source_summary" {
			t.Fatalf("unexpected evidence ref: %#v", request.Trace.EvidenceRefs)
		}
		return ObjectiveFinalAnswerSynthesisResponse{
			ResponseRef: "llm:final_answer:satisfied",
			Draft: ObjectiveFinalAnswerDraft{
				DraftRef:     "draft:final:satisfied",
				Status:       VerificationSatisfied,
				AnswerText:   "已完成目标，结论来自已验证的摘要证据。",
				Conclusion:   "目标已完成。",
				Steps:        []string{"获取并总结公开来源"},
				Evidence:     []string{"摘要证据满足目标要求"},
				EvidenceRefs: evidence,
			},
			Boundaries: []Boundary{"llm_phrase_only"},
		}, nil
	})

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		TraceInput: ObjectiveFinalAnswerTraceInput{
			Spec: ObjectiveSpec{
				SpecRef:     "spec:source_summary",
				ObjectiveID: "objective:source_summary",
				GoalSummary: "Summarize a public source",
			},
			Verification: ObjectiveVerificationGateResult{
				Status:       VerificationSatisfied,
				Satisfied:    true,
				Frame:        ObjectiveFrame{ID: "objective:source_summary"},
				EvidenceRefs: evidence,
			},
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:       "step:source_summary",
				NodeRef:       "node:source_summary",
				CapabilityRef: "capability:public_source",
				StrategyRef:   "strategy:source_summary",
				Status:        VerificationSatisfied,
				Action:        "获取并总结公开来源",
				EvidenceRefs:  evidence,
			}},
		},
		EnableSynthesizer: true,
		Synthesizer:       synthesizer,
	})

	if !got.ReadyForUser || got.Status != VerificationSatisfied || got.SynthesisRef != "llm:final_answer:satisfied" {
		t.Fatalf("unexpected final answer = %#v", got)
	}
	if got.FailureClass != FailureNone {
		t.Fatalf("expected no failure, got %s", got.FailureClass)
	}
	if !objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "trace_backed_synthesis_only") ||
		!objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "answer_text_must_be_supported_by_trace") ||
		!objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "llm_phrase_only") {
		t.Fatalf("missing final answer boundaries: %#v", got.Boundaries)
	}
}

func TestBuildObjectiveFinalAnswerPartialRequiresLimitations(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:partial_inventory",
		Kind:     "inventory",
		Strength: EvidenceAdequate,
	}}

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: ObjectiveFinalAnswerTrace{
			TraceRef:     "trace:partial",
			ObjectiveID:  "objective:partial",
			Status:       VerificationPartial,
			EvidenceRefs: evidence,
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:      "step:partial",
				Status:       VerificationPartial,
				Action:       "执行部分只读查询",
				EvidenceRefs: evidence,
			}},
			NextHostAction: "provide_partial_next_step",
		},
		Draft: ObjectiveFinalAnswerDraft{
			DraftRef:     "draft:partial",
			Status:       VerificationPartial,
			AnswerText:   "已得到部分结果。",
			Conclusion:   "部分完成。",
			Steps:        []string{"执行部分只读查询"},
			EvidenceRefs: evidence,
			NextStep:     "补充缺失证据",
		},
	})

	if got.ReadyForUser || got.Status != VerificationReviewRequired {
		t.Fatalf("expected review-required partial answer, got %#v", got)
	}
	if !objectiveFinalAnswerTestMissingContains(got.MissingInputs, "host:objective_final_answer_limitations") ||
		!objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "partial_answer_limitations_missing") {
		t.Fatalf("missing partial completeness block: %#v", got)
	}
}

func TestBuildObjectiveFinalAnswerBlockedCapabilityGapCanBeDisplayed(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:capability_gap",
		Kind:     "capability_gap",
		Strength: EvidenceAdequate,
	}}

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: ObjectiveFinalAnswerTrace{
			TraceRef:       "trace:capability_gap",
			ObjectiveID:    "objective:capability_gap",
			Status:         VerificationBlocked,
			EvidenceRefs:   evidence,
			FailureClass:   FailureCapabilityMissing,
			MissingInputs:  []MissingInput{"host:capability_descriptor"},
			Limitations:    []string{"当前 catalog 中没有可用能力完成目标。"},
			NextHostAction: "enter_capability_resolution",
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:       "step:capability_gap",
				CapabilityRef: "capability:missing",
				Status:        VerificationBlocked,
				Action:        "确认能力缺口",
				EvidenceRefs:  evidence,
				FailureClass:  FailureCapabilityMissing,
			}},
		},
		Draft: ObjectiveFinalAnswerDraft{
			DraftRef:     "draft:capability_gap",
			Status:       VerificationBlocked,
			AnswerText:   "当前缺少可执行能力，无法继续完成目标。",
			Conclusion:   "目标被能力缺口阻塞。",
			Steps:        []string{"确认能力缺口"},
			EvidenceRefs: evidence,
			Limitations:  []string{"当前 catalog 中没有可用能力完成目标。"},
			NextStep:     "enter_capability_resolution",
		},
	})

	if !got.ReadyForUser || got.Status != VerificationBlocked || got.FailureClass != FailureCapabilityMissing {
		t.Fatalf("expected displayable blocked capability gap, got %#v", got)
	}
	if got.NextHostAction != "enter_capability_resolution" {
		t.Fatalf("expected capability resolution action, got %s", got.NextHostAction)
	}
}

func TestBuildObjectiveFinalAnswerBlockedExternalLoginExpiredCanBeDisplayed(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:login_invalid",
		Kind:     "auth_state",
		Strength: EvidenceAdequate,
	}}

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: ObjectiveFinalAnswerTrace{
			TraceRef:       "trace:external_login_expired",
			ObjectiveID:    "objective:external_login_expired",
			Status:         VerificationBlocked,
			EvidenceRefs:   evidence,
			FailureClass:   FailureAuthorizationMissing,
			MissingInputs:  []MissingInput{"host:external_login_state"},
			Limitations:    []string{"外部登录态已失效，需要宿主重新授权。"},
			NextHostAction: "open_external_login",
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:       "step:login_check",
				CapabilityRef: "capability:external_login_check",
				Status:        VerificationBlocked,
				Action:        "检查外部登录态",
				EvidenceRefs:  evidence,
				FailureClass:  FailureAuthorizationMissing,
			}},
		},
		Draft: ObjectiveFinalAnswerDraft{
			DraftRef:     "draft:external_login_expired",
			Status:       VerificationBlocked,
			AnswerText:   "当前登录态已失效，无法继续访问受保护资源。",
			Conclusion:   "需要重新授权后再继续。",
			Steps:        []string{"检查外部登录态"},
			EvidenceRefs: evidence,
			Limitations:  []string{"外部登录态已失效，需要宿主重新授权。"},
			NextStep:     "open_external_login",
		},
	})

	if !got.ReadyForUser || got.Status != VerificationBlocked || got.FailureClass != FailureAuthorizationMissing {
		t.Fatalf("expected displayable external-login blocked answer, got %#v", got)
	}
	if got.NextHostAction != "open_external_login" {
		t.Fatalf("expected open_external_login action, got %s", got.NextHostAction)
	}
}

func TestBuildObjectiveFinalAnswerShowsRecoveryPatchClosure(t *testing.T) {
	recovery := BuildObjectiveRecoveryContract(ObjectiveRecoveryContractInput{
		ContractRef:            "contract:final_trace_recovery",
		SourceRef:              "source:first_attempt",
		Producer:               "adapter:first_attempt",
		ObjectiveID:            "objective:final_trace_recovery",
		CurrentStrategyRef:     "strategy:first_attempt",
		RecoveryRecommended:    true,
		FinalAnswerRecommended: false,
		FailureClass:           FailureEvidenceMissing,
		Targets: []ObjectiveRecoveryTarget{{
			TargetRef:    "target:missing_dimension",
			MissingInput: "evidence:recovery_dimension",
			MissingEvidence: EvidenceRef{
				Ref:      "evidence:recovery_dimension",
				Kind:     "recovery_dimension",
				Strength: EvidenceAdequate,
				Source:   "adapter:first_attempt",
			},
			SuggestedToolRefs: []DisplaySafeRef{"capability:recovery_lookup"},
		}},
		SuggestedToolRefs: []DisplaySafeRef{"capability:recovery_lookup"},
	})
	patch := BuildObjectiveReplanGraphPatch(ObjectiveReplanGraphPatchInput{
		PatchRef:       "patch:final_trace_recovery",
		SourceGraphRef: "graph:first_attempt",
		SourceNodeRef:  "node:first_attempt",
		Proposal:       recovery.ReplanProposal,
	})
	if !patch.ReadyForHostReview || len(patch.PatchNodes) != 1 {
		t.Fatalf("expected host-reviewable patch: %#v", patch)
	}
	boundNode := objectiveFinalAnswerTestBoundPatchNode(patch.PatchNodes[0])
	graph := ObjectiveGraph{
		GraphRef:         "graph:final_trace_recovery",
		SpecRef:          "spec:final_trace_recovery",
		ObjectiveID:      "objective:final_trace_recovery",
		CatalogRef:       "catalog:final_trace_recovery",
		Nodes:            []ObjectiveNode{boundNode},
		RequiredEvidence: boundNode.RequiredEvidence,
	}.Normalize()

	trace := BuildObjectiveFinalAnswerTrace(ObjectiveFinalAnswerTraceInput{
		TraceRef:         "trace:final_trace_recovery",
		Spec:             ObjectiveSpec{SpecRef: "spec:final_trace_recovery", ObjectiveID: "objective:final_trace_recovery", GoalSummary: "Recover missing evidence"},
		Graph:            graph,
		Replan:           recovery.ReplanProposal,
		ReplanGraphPatch: patch,
		Status:           VerificationSatisfied,
		NextHostAction:   "synthesize_trace_backed_final_answer",
	})

	if trace.Status != VerificationSatisfied ||
		len(trace.Steps) < 4 ||
		!objectiveFinalAnswerTestStepActionContains(trace.Steps, "add_evidence_node") ||
		!objectiveFinalAnswerTestStepActionContains(trace.Steps, "objective_replan_graph_patch_review") ||
		!objectiveFinalAnswerTestStepActionContains(trace.Steps, "objective_replan_graph_patch_node") ||
		!objectiveFinalAnswerTestStepActionContains(trace.Steps, "host_adapter") ||
		!objectiveFinalAnswerTestEvidenceContains(trace.EvidenceRefs, "evidence:recovery_dimension") ||
		!objectiveFinalAnswerTestBoundaryContains(trace.Boundaries, "objective_replan_graph_patch_proposal_only") ||
		!objectiveFinalAnswerTestBoundaryContains(trace.Boundaries, "host_reviewed_replan_graph_patch") {
		t.Fatalf("trace should expose recovery patch closure: %#v", trace)
	}

	answer := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: trace,
	})
	if !answer.ReadyForUser ||
		answer.Status != VerificationSatisfied ||
		!objectiveFinalAnswerTestDraftStepContains(answer.Draft.Steps, "add_evidence_node") ||
		!objectiveFinalAnswerTestDraftStepContains(answer.Draft.Steps, "objective_replan_graph_patch_review") ||
		!objectiveFinalAnswerTestDraftStepContains(answer.Draft.Steps, "host_adapter") {
		t.Fatalf("final answer should expose recovery path steps: %#v", answer)
	}
}

func TestBuildObjectiveFinalAnswerTraceSkipsAbsentRecoveryPatch(t *testing.T) {
	trace := BuildObjectiveFinalAnswerTrace(ObjectiveFinalAnswerTraceInput{
		TraceRef: "trace:no_recovery_patch",
		Graph: ObjectiveGraph{
			GraphRef: "graph:no_recovery_patch",
			Nodes: []ObjectiveNode{{
				NodeRef:       "node:answer",
				Kind:          "answer",
				State:         ObjectiveNodeStateSatisfied,
				CapabilityRef: "capability:answer",
				StrategyRef:   "strategy:answer",
				RequiredEvidence: []EvidenceRef{{
					Ref:      "evidence:answer",
					Kind:     "answer",
					Strength: EvidenceStrong,
					Source:   "tool:answer",
				}},
			}},
		},
		Status: VerificationSatisfied,
	})
	if objectiveFinalAnswerTestStepActionContains(trace.Steps, "objective_replan_graph_patch_review") ||
		objectiveFinalAnswerTestStepActionContains(trace.Steps, "objective_replan_graph_patch_node") {
		t.Fatalf("empty patch should not produce recovery patch trace steps: %#v", trace.Steps)
	}
}

func TestBuildObjectiveFinalAnswerRejectsUnsafeDraftOutput(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:summary",
		Kind:     "summary",
		Strength: EvidenceStrong,
	}}

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: ObjectiveFinalAnswerTrace{
			TraceRef:     "trace:unsafe_draft",
			ObjectiveID:  "objective:unsafe_draft",
			Status:       VerificationSatisfied,
			EvidenceRefs: evidence,
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:      "step:summary",
				Status:       VerificationSatisfied,
				Action:       "总结公开来源",
				EvidenceRefs: evidence,
			}},
		},
		Draft: ObjectiveFinalAnswerDraft{
			DraftRef:     "draft:unsafe",
			Status:       VerificationSatisfied,
			AnswerText:   "raw output is available at https://example.com/internal",
			Conclusion:   "目标已完成。",
			Steps:        []string{"总结公开来源"},
			EvidenceRefs: evidence,
		},
	})

	if got.ReadyForUser || got.Status != VerificationReviewRequired {
		t.Fatalf("expected unsafe draft review, got %#v", got)
	}
	if !objectiveFinalAnswerTestMissingContains(got.MissingInputs, "host:display_safe_final_answer_trace") ||
		!objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("missing display-safe block: %#v", got)
	}
}

func TestBuildObjectiveFinalAnswerSynthesizerFailsClosed(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:summary",
		Kind:     "summary",
		Strength: EvidenceStrong,
	}}
	synthesizer := ObjectiveFinalAnswerSynthesizerFunc(func(context.Context, ObjectiveFinalAnswerSynthesisRequest) (ObjectiveFinalAnswerSynthesisResponse, error) {
		return ObjectiveFinalAnswerSynthesisResponse{}, errors.New("synthesizer unavailable")
	})

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: ObjectiveFinalAnswerTrace{
			TraceRef:     "trace:synthesizer_error",
			ObjectiveID:  "objective:synthesizer_error",
			Status:       VerificationSatisfied,
			EvidenceRefs: evidence,
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:      "step:summary",
				Status:       VerificationSatisfied,
				Action:       "总结公开来源",
				EvidenceRefs: evidence,
			}},
		},
		EnableSynthesizer: true,
		Synthesizer:       synthesizer,
	})

	if got.ReadyForUser || got.Status != VerificationReviewRequired || got.FailureClass != FailureInternalError {
		t.Fatalf("expected synthesizer fail-closed, got %#v", got)
	}
	if !objectiveFinalAnswerTestMissingContains(got.MissingInputs, "host:objective_final_answer_synthesizer") ||
		!objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "objective_final_answer_synthesizer_failed") {
		t.Fatalf("missing synthesizer failure block: %#v", got)
	}
}

func TestBuildObjectiveFinalAnswerSynthesizerRawResponseRequiresDisplaySafeReview(t *testing.T) {
	evidence := []EvidenceRef{{
		Ref:      "evidence:summary",
		Kind:     "summary",
		Strength: EvidenceStrong,
	}}
	synthesizer := ObjectiveFinalAnswerSynthesizerFunc(func(context.Context, ObjectiveFinalAnswerSynthesisRequest) (ObjectiveFinalAnswerSynthesisResponse, error) {
		return ObjectiveFinalAnswerSynthesisResponse{
			ResponseRef: "llm:final_answer:raw",
			Draft: ObjectiveFinalAnswerDraft{
				DraftRef:     "draft:raw",
				Status:       VerificationSatisfied,
				AnswerText:   "read /Users/mason/private/raw-output.txt",
				Conclusion:   "目标已完成。",
				Steps:        []string{"总结公开来源"},
				EvidenceRefs: evidence,
			},
		}, nil
	})

	got := BuildObjectiveFinalAnswer(context.Background(), ObjectiveFinalAnswerInput{
		Trace: ObjectiveFinalAnswerTrace{
			TraceRef:     "trace:raw_synthesizer",
			ObjectiveID:  "objective:raw_synthesizer",
			Status:       VerificationSatisfied,
			EvidenceRefs: evidence,
			Steps: []ObjectiveFinalAnswerTraceStep{{
				StepRef:      "step:summary",
				Status:       VerificationSatisfied,
				Action:       "总结公开来源",
				EvidenceRefs: evidence,
			}},
		},
		EnableSynthesizer: true,
		Synthesizer:       synthesizer,
	})

	if got.ReadyForUser || got.Status != VerificationReviewRequired {
		t.Fatalf("expected unsafe synthesizer review, got %#v", got)
	}
	if !objectiveFinalAnswerTestBoundaryContains(got.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("missing raw-output boundary: %#v", got.Boundaries)
	}
}

func objectiveFinalAnswerTestBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveFinalAnswerTestMissingContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func objectiveFinalAnswerTestBoundPatchNode(node ObjectiveNode) ObjectiveNode {
	out := node.Normalize()
	out.Kind = "host_adapter"
	out.State = ObjectiveNodeStateSatisfied
	out.DescriptorRef = "descriptor:recovery_lookup"
	out.SourceRef = "source:recovery_lookup"
	out.InputSchemaRef = "schema:recovery_lookup.input.v1"
	out.OutputSchemaRef = "schema:recovery_lookup.output.v1"
	out.EvidenceContractRef = "evidence:recovery_lookup.contract.v1"
	out.SideEffectClass = ObjectiveCapabilitySideEffectReadOnly
	out.MissingInputs = nil
	out.Boundaries = AppendBoundaries(out.Boundaries, "host_reviewed_replan_graph_patch")
	return out.Normalize()
}

func objectiveFinalAnswerTestStepActionContains(steps []ObjectiveFinalAnswerTraceStep, want string) bool {
	for _, step := range steps {
		if step.Action == want {
			return true
		}
	}
	return false
}

func objectiveFinalAnswerTestEvidenceContains(values []EvidenceRef, want DisplaySafeRef) bool {
	for _, value := range values {
		if value.Normalize().Ref == want {
			return true
		}
	}
	return false
}

func objectiveFinalAnswerTestDraftStepContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
