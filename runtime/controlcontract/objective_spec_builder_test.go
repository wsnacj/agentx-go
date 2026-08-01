package controlcontract

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
)

func TestBuildObjectiveSpecFromJSONStrictDecodeReady(t *testing.T) {
	raw := mustObjectiveSpecJSON(t, objectiveSpecBuilderTestSpec("capability:public_source_fetch"))

	got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{
		RawJSON:                   raw,
		SpecRef:                   "spec:fallback",
		RawGoalRef:                "goal:public_source",
		UserGoalDigest:            "sha256:abc",
		ProjectionRef:             "projection:public_source",
		SourceRef:                 "host:objective_spec_builder",
		AllowedCapabilityRefs:     []DisplaySafeRef{"capability:public_source_fetch"},
		AllowedSideEffectPolicies: []ObjectiveSpecSideEffectPolicy{ObjectiveSpecSideEffectReadOnly},
	})
	if got.Status != VerificationSatisfied ||
		!got.Decoded ||
		!got.ReadyForObjectiveFrame ||
		got.FailureClass != FailureNone ||
		got.NextHostAction != "run_objective_graph_planner" {
		t.Fatalf("unexpected decode report = %#v", got)
	}
	if got.Spec.RawGoalRef != "goal:public_source" || got.Spec.UserGoalDigest != "sha256:abc" {
		t.Fatalf("fallback refs not applied = %#v", got.Spec)
	}
	if !objectiveSpecTestBoundaryContains(got.Boundaries, "strict_json_decoder") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_json_validated") {
		t.Fatalf("boundaries = %#v", got.Boundaries)
	}
}

func TestBuildObjectiveSpecFromJSONRejectsUnknownField(t *testing.T) {
	raw := []byte(`{"goal_summary":"summarize public site","unknown_field":true}`)

	got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{RawJSON: raw})
	if got.Status != VerificationBlocked ||
		got.Decoded ||
		got.ReadyForObjectiveFrame ||
		got.FailureClass != FailureInvalidInput ||
		got.NextHostAction != "provide_objective_spec_json" ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_json_decode_failed") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "deterministic_blocked_fallback") {
		t.Fatalf("unexpected unknown field report = %#v", got)
	}
}

func TestBuildObjectiveSpecFromJSONRejectsTrailingTokens(t *testing.T) {
	raw := append(mustObjectiveSpecJSON(t, objectiveSpecBuilderTestSpec("capability:public_source_fetch")), []byte(` {}`)...)

	got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{RawJSON: raw})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureInvalidInput ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_json_trailing_tokens") {
		t.Fatalf("unexpected trailing tokens report = %#v", got)
	}
}

func TestBuildObjectiveSpecFromJSONRejectsUnknownCapability(t *testing.T) {
	raw := mustObjectiveSpecJSON(t, objectiveSpecBuilderTestSpec("capability:not_in_catalog"))

	got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{
		RawJSON:               raw,
		AllowedCapabilityRefs: []DisplaySafeRef{"capability:public_source_fetch"},
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureCapabilityMissing ||
		got.NextHostAction != "provide_strategy_scope" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:allowed_capability_ref") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_candidate_capability_not_allowed") {
		t.Fatalf("unexpected capability report = %#v", got)
	}
}

func TestBuildObjectiveSpecFromJSONRejectsUnauthorizedSideEffect(t *testing.T) {
	spec := objectiveSpecBuilderTestSpec("capability:ticket_purchase")
	spec.SideEffectPolicy = ObjectiveSpecSideEffectAllowed
	raw := mustObjectiveSpecJSON(t, spec)

	got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{
		RawJSON:                   raw,
		AllowedSideEffectPolicies: []ObjectiveSpecSideEffectPolicy{ObjectiveSpecSideEffectReadOnly},
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailurePolicyBlocked ||
		got.NextHostAction != "request_host_approval" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_side_effect_policy") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_side_effect_policy_not_allowed") {
		t.Fatalf("unexpected side effect report = %#v", got)
	}
}

func TestBuildObjectiveSpecFromJSONRepresentativeObjectiveContracts(t *testing.T) {
	cases := []struct {
		name       string
		summary    string
		capability DisplaySafeRef
		evidence   []string
	}{
		{
			name:       "url_summary",
			summary:    "summarize public webpage in requested language",
			capability: "capability:public_source_fetch",
			evidence:   []string{"source", "summary"},
		},
		{
			name:       "transport_query",
			summary:    "query transport schedule and availability",
			capability: "capability:public_transport_lookup",
			evidence:   []string{"normalized_query", "transport_results"},
		},
		{
			name:       "container_database",
			summary:    "inspect container runtime and database schema",
			capability: "capability:container_database_inspection",
			evidence:   []string{"container_inventory", "database_schema"},
		},
		{
			name:       "wechat_article",
			summary:    "search public account and download article markdown",
			capability: "capability:wechat_article_sync",
			evidence:   []string{"login_status", "account_match", "article_markdown"},
		},
		{
			name:       "life_question",
			summary:    "answer practical life planning question with available evidence",
			capability: "capability:general_public_research",
			evidence:   []string{"source", "recommendation"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spec := objectiveSpecBuilderScenarioSpec(tc.name, tc.summary, tc.capability, ObjectiveSpecSideEffectReadOnly, tc.evidence...)
			got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{
				RawJSON:                   mustObjectiveSpecJSON(t, spec),
				RawGoalRef:                DisplaySafeRef("goal:" + tc.name),
				UserGoalDigest:            "sha256:" + tc.name,
				ProjectionRef:             DisplaySafeRef("projection:" + tc.name),
				SourceRef:                 "host:objective_spec_builder",
				AllowedCapabilityRefs:     []DisplaySafeRef{tc.capability},
				AllowedSideEffectPolicies: []ObjectiveSpecSideEffectPolicy{ObjectiveSpecSideEffectReadOnly},
			})
			if got.Status != VerificationSatisfied ||
				!got.Decoded ||
				!got.ReadyForObjectiveFrame ||
				got.FailureClass != FailureNone ||
				got.NextHostAction != "run_objective_graph_planner" {
				t.Fatalf("unexpected representative report = %#v", got)
			}
			if len(got.Projection.Frame.RequiredEvidence) < len(tc.evidence) {
				t.Fatalf("required evidence not projected: spec=%#v frame=%#v", got.Spec.RequiredEvidence, got.Projection.Frame.RequiredEvidence)
			}
			if len(got.Projection.Frame.CandidateCapabilities) != 1 || got.Projection.Frame.CandidateCapabilities[0] != tc.capability {
				t.Fatalf("candidate capabilities = %#v", got.Projection.Frame.CandidateCapabilities)
			}
		})
	}
}

func TestBuildObjectiveSpecFromJSONRepresentativeSideEffectfulGoalBlocks(t *testing.T) {
	spec := objectiveSpecBuilderScenarioSpec(
		"booking_order",
		"book or place order after checking availability",
		"capability:purchase_action",
		ObjectiveSpecSideEffectAllowed,
		"availability",
		"purchase_confirmation",
	)

	got := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{
		RawJSON:                   mustObjectiveSpecJSON(t, spec),
		RawGoalRef:                "goal:booking_order",
		ProjectionRef:             "projection:booking_order",
		SourceRef:                 "host:objective_spec_builder",
		AllowedCapabilityRefs:     []DisplaySafeRef{"capability:purchase_action"},
		AllowedSideEffectPolicies: []ObjectiveSpecSideEffectPolicy{ObjectiveSpecSideEffectReadOnly},
	})
	if got.Status != VerificationBlocked ||
		got.ReadyForObjectiveFrame ||
		got.FailureClass != FailurePolicyBlocked ||
		got.NextHostAction != "request_host_approval" ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_side_effect_policy_not_allowed") {
		t.Fatalf("unexpected side-effectful representative report = %#v", got)
	}
}

func TestBuildObjectiveSpecWithBuilderDisabledDoesNotCallBuilder(t *testing.T) {
	calls := 0
	builder := ObjectiveSpecBuilderFunc(func(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error) {
		calls++
		return ObjectiveSpecBuilderResponse{}, nil
	})

	got := BuildObjectiveSpecWithBuilder(context.Background(), ObjectiveSpecBuildInput{
		Enabled: false,
		Builder: builder,
		Request: ObjectiveSpecBuilderRequest{RequestRef: "request:disabled"},
	})
	if calls != 0 {
		t.Fatalf("builder called while disabled")
	}
	if got.Status != VerificationBlocked ||
		got.BuilderCalled ||
		got.FailureClass != FailureInsufficientInformation ||
		got.NextHostAction != "enable_objective_closure" {
		t.Fatalf("unexpected disabled report = %#v", got)
	}
}

func TestBuildObjectiveSpecWithBuilderInvalidJSONFallback(t *testing.T) {
	got := BuildObjectiveSpecWithBuilder(context.Background(), ObjectiveSpecBuildInput{
		Enabled: true,
		Builder: ObjectiveSpecBuilderFunc(func(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error) {
			return ObjectiveSpecBuilderResponse{
				ResponseRef: "response:invalid_json",
				SpecJSON:    []byte(`{"goal_summary":"safe summary","extra":true}`),
			}, nil
		}),
		Request: ObjectiveSpecBuilderRequest{
			RequestRef:                "request:invalid_json",
			RawGoalRef:                "goal:invalid_json",
			AllowedCapabilityRefs:     []DisplaySafeRef{"capability:public_source_fetch"},
			AllowedSideEffectPolicies: []ObjectiveSpecSideEffectPolicy{ObjectiveSpecSideEffectReadOnly},
		},
	})
	if got.Status != VerificationBlocked ||
		!got.BuilderCalled ||
		!got.DecodeAttempted ||
		got.Built ||
		got.FailureClass != FailureInvalidInput ||
		got.NextHostAction != "provide_objective_spec_json" ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "deterministic_blocked_fallback") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "no_prompt_heuristic_fallback") {
		t.Fatalf("unexpected invalid json builder report = %#v", got)
	}
}

func TestBuildObjectiveSpecWithBuilderDirectSpecReady(t *testing.T) {
	got := BuildObjectiveSpecWithBuilder(context.Background(), ObjectiveSpecBuildInput{
		Enabled: true,
		Builder: ObjectiveSpecBuilderFunc(func(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error) {
			return ObjectiveSpecBuilderResponse{
				ResponseRef: "response:direct_spec",
				Spec:        objectiveSpecBuilderTestSpec("capability:public_source_fetch"),
			}, nil
		}),
		Request: ObjectiveSpecBuilderRequest{
			RequestRef:                "request:direct_spec",
			RawGoalRef:                "goal:direct_spec",
			UserGoalDigest:            "sha256:def",
			AllowedCapabilityRefs:     []DisplaySafeRef{"capability:public_source_fetch"},
			AllowedSideEffectPolicies: []ObjectiveSpecSideEffectPolicy{ObjectiveSpecSideEffectReadOnly},
		},
		ProjectionRef: "projection:direct_spec",
		SourceRef:     "host:objective_spec_builder",
	})
	if got.Status != VerificationSatisfied ||
		!got.Built ||
		!got.ReadyForObjectiveFrame ||
		got.DecodeAttempted ||
		got.FailureClass != FailureNone ||
		got.Spec.RawGoalRef != "goal:direct_spec" ||
		got.Spec.UserGoalDigest != "sha256:def" {
		t.Fatalf("unexpected direct spec report = %#v", got)
	}
}

func TestBuildObjectiveSpecWithBuilderEmptyResponseBlocked(t *testing.T) {
	got := BuildObjectiveSpecWithBuilder(context.Background(), ObjectiveSpecBuildInput{
		Enabled: true,
		Builder: ObjectiveSpecBuilderFunc(func(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error) {
			return ObjectiveSpecBuilderResponse{ResponseRef: "response:empty"}, nil
		}),
		Request: ObjectiveSpecBuilderRequest{RequestRef: "request:empty"},
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureInvalidInput ||
		got.NextHostAction != "provide_objective_spec" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_spec") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_builder_empty_response") {
		t.Fatalf("unexpected empty response report = %#v", got)
	}
}

func TestBuildObjectiveSpecWithBuilderErrorIsDisplaySafeBlocked(t *testing.T) {
	got := BuildObjectiveSpecWithBuilder(context.Background(), ObjectiveSpecBuildInput{
		Enabled: true,
		Builder: ObjectiveSpecBuilderFunc(func(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error) {
			return ObjectiveSpecBuilderResponse{}, errors.New("secret: should not be surfaced")
		}),
		Request: ObjectiveSpecBuilderRequest{RequestRef: "request:error"},
	})
	if got.Status != VerificationBlocked ||
		got.FailureClass != FailureExternalDependencyUnavailable ||
		got.NextHostAction != "provide_objective_spec" ||
		!objectiveSpecTestMissingContains(got.MissingInputs, "host:objective_spec_builder_result") ||
		!objectiveSpecTestBoundaryContains(got.Boundaries, "objective_spec_builder_failed") {
		t.Fatalf("unexpected error report = %#v", got)
	}
}

func objectiveSpecBuilderTestSpec(capability DisplaySafeRef) ObjectiveSpec {
	return ObjectiveSpec{
		SpecRef:          "spec:builder_test",
		ObjectiveID:      "objective:builder_test",
		GoalSummary:      "answer public information request",
		SideEffectPolicy: ObjectiveSpecSideEffectReadOnly,
		SuccessCriteria: []ObjectiveSuccessCriterion{{
			CriteriaRef: "criteria:answer",
			Text:        "produce grounded answer",
			RequiredEvidence: []EvidenceRef{{
				Ref:      "evidence:answer",
				Kind:     "answer",
				Strength: EvidenceAdequate,
				Source:   "scene:public_source",
			}},
		}},
		RequiredEvidence: []EvidenceRef{{
			Ref:      "evidence:source",
			Kind:     "source",
			Strength: EvidenceAdequate,
			Source:   "scene:public_source",
		}},
		CandidateCapabilities: []DisplaySafeRef{capability},
		SourceContext:         []DisplaySafeRef{"catalog:objective_capabilities"},
	}
}

func objectiveSpecBuilderScenarioSpec(name, summary string, capability DisplaySafeRef, policy ObjectiveSpecSideEffectPolicy, evidenceKinds ...string) ObjectiveSpec {
	requiredEvidence := make([]EvidenceRef, 0, len(evidenceKinds))
	criteriaEvidence := make([]EvidenceRef, 0, len(evidenceKinds))
	for _, kind := range evidenceKinds {
		ref := DisplaySafeRef("evidence:" + name + "_" + kind)
		evidence := EvidenceRef{
			Ref:      ref,
			Kind:     kind,
			Strength: EvidenceAdequate,
			Source:   capability,
		}
		requiredEvidence = append(requiredEvidence, evidence)
		criteriaEvidence = append(criteriaEvidence, evidence)
	}
	return ObjectiveSpec{
		SpecRef:          DisplaySafeRef("spec:" + name),
		ObjectiveID:      "objective:" + name,
		GoalSummary:      summary,
		SideEffectPolicy: policy,
		SuccessCriteria: []ObjectiveSuccessCriterion{{
			CriteriaRef:      DisplaySafeRef("criteria:" + name),
			Text:             "satisfy " + name + " objective with required evidence",
			RequiredEvidence: criteriaEvidence,
		}},
		RequiredEvidence:      requiredEvidence,
		CandidateCapabilities: []DisplaySafeRef{capability},
		SourceContext:         []DisplaySafeRef{"catalog:objective_capabilities"},
	}
}

func mustObjectiveSpecJSON(t *testing.T, spec ObjectiveSpec) []byte {
	t.Helper()
	raw, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}
	return raw
}
