package controlcontract

import (
	"reflect"
	"testing"
)

func TestMergeBoundaryMissingAndEvidenceRefs(t *testing.T) {
	boundaries := MergeBoundaries(
		[]Boundary{" approval required ", "approval_required", "https://example.invalid/raw"},
		[]Boundary{"same_strategy_only"},
	)
	if len(boundaries) != 2 || boundaries[0] != "approval_required" || boundaries[1] != "same_strategy_only" {
		t.Fatalf("boundaries = %#v", boundaries)
	}

	missing := AppendMissingInputs([]MissingInput{"host:approval", "host:approval"}, "contract:budget.max_retries", "/tmp/raw")
	if len(missing) != 2 || missing[0] != "host:approval" || missing[1] != "contract:budget.max_retries" {
		t.Fatalf("missing inputs = %#v", missing)
	}

	refs := MergeEvidenceRefs(
		[]EvidenceRef{{Ref: "evidence:one", Kind: " metric ", Strength: "strong", Source: "host:collector"}},
		[]EvidenceRef{{Ref: "evidence:one", Kind: "metric", Strength: "weak", Source: "host:collector"}},
		[]EvidenceRef{{Ref: "https://example.invalid/raw", Kind: "raw"}},
	)
	if len(refs) != 1 ||
		refs[0].Ref != "evidence:one" ||
		refs[0].Kind != "metric" ||
		refs[0].Strength != EvidenceStrong {
		t.Fatalf("evidence refs = %#v", refs)
	}
}

func TestEvaluateHostApprovalGate(t *testing.T) {
	notRequired := EvaluateHostApprovalGate(ApprovalGateInput{Kind: "retry_same_strategy"})
	if notRequired.Status != HostActionReady ||
		notRequired.RequiresApproval ||
		notRequired.FailureClass != FailureNone ||
		notRequired.NextHostAction != "host_may_continue" {
		t.Fatalf("not required = %#v", notRequired)
	}

	required := EvaluateHostApprovalGate(ApprovalGateInput{
		Kind:             "retry_same_strategy",
		RequiresApproval: true,
	})
	if required.Status != HostActionRequiresApproval ||
		required.FailureClass != FailureApprovalRequired ||
		len(required.MissingInputs) != 1 ||
		required.MissingInputs[0] != "host:approval" ||
		required.NextHostAction != "request_host_approval" {
		t.Fatalf("required = %#v", required)
	}

	approved := EvaluateHostApprovalGate(ApprovalGateInput{
		Kind:             "retry_same_strategy",
		RequiresApproval: true,
		Approved:         true,
		ApprovalRefs:     []DisplaySafeRef{"approval:retry_1"},
	})
	if approved.Status != HostActionReady ||
		approved.FailureClass != FailureNone ||
		len(approved.ApprovalRefs) != 1 ||
		approved.ApprovalRefs[0] != "approval:retry_1" ||
		len(approved.Boundaries) != 1 ||
		approved.Boundaries[0] != "host_approval_granted" {
		t.Fatalf("approved = %#v", approved)
	}

	approvedWithoutRef := EvaluateHostApprovalGate(ApprovalGateInput{
		Kind:             "retry_same_strategy",
		RequiresApproval: true,
		Approved:         true,
		ApprovalRefs:     []DisplaySafeRef{"https://example.invalid/raw"},
	})
	if approvedWithoutRef.Status != HostActionReviewRequired ||
		approvedWithoutRef.FailureClass != FailureEvidenceMissing ||
		len(approvedWithoutRef.MissingInputs) != 1 ||
		approvedWithoutRef.MissingInputs[0] != "host:approval_ref" {
		t.Fatalf("approved without ref = %#v", approvedWithoutRef)
	}
}

func TestBuildManagedObjectiveProjectionRequiresActivationAndInputs(t *testing.T) {
	inactive := BuildManagedObjectiveProjection(ManagedObjectiveProjectionInput{
		Activation: ActivationObserveOnly,
		Frame: ObjectiveFrame{
			ID:             "objective:one",
			UserGoalDigest: "sha256:goal",
		},
	})
	if inactive.Status != HostActionNotReady ||
		inactive.FailureClass != FailurePolicyBlocked ||
		inactive.NextHostAction != "enable_managed_objective" ||
		inactive.Ready ||
		inactive.RunnerEffect != "none" ||
		inactive.PromptEffect != "none" {
		t.Fatalf("inactive projection = %#v", inactive)
	}
	if !managedObjectiveMissingInputContains(inactive.MissingInputs, "control_plane:managed_activation") ||
		!managedObjectiveBoundaryContains(inactive.Boundaries, "managed_objective_activation_required") {
		t.Fatalf("inactive missing/boundaries = %#v / %#v", inactive.MissingInputs, inactive.Boundaries)
	}

	missing := BuildManagedObjectiveProjection(ManagedObjectiveProjectionInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:             "objective:one",
			UserGoalDigest: "sha256:goal",
		},
	})
	if missing.Status != HostActionNotReady ||
		missing.FailureClass != FailureConfigMissing ||
		missing.NextHostAction != "provide_managed_objective_contract" ||
		missing.Ready {
		t.Fatalf("missing contract projection = %#v", missing)
	}
	for _, want := range []MissingInput{
		"host:success_criteria",
		"host:managed_objective_ledger_ref",
		"host:strategy_scope",
		"contract:intensity_gate",
		"contract:budget",
		"contract:approval_policy",
		"contract:strategy_scope",
		"contract:redaction_policy",
	} {
		if !managedObjectiveMissingInputContains(missing.MissingInputs, want) {
			t.Fatalf("expected missing input %q, got %#v", want, missing.MissingInputs)
		}
	}
	if missing.Frame.ControlMode != ControlModeObjective ||
		missing.Frame.Intensity != IntensityL3ManagedObjective ||
		!managedObjectiveBoundaryContains(missing.Boundaries, "no_strategy_dispatch") {
		t.Fatalf("unexpected normalized managed frame/boundaries = %#v / %#v", missing.Frame, missing.Boundaries)
	}
}

func TestBuildManagedObjectiveProjectionReadyOnlyAfterApproval(t *testing.T) {
	input := ManagedObjectiveProjectionInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:approved",
			UserGoalDigest:  "sha256:goal",
			SuccessCriteria: []string{"produce evidence"},
		},
		LedgerRef: "ledger:approved",
		PolicyRefs: []DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
		},
		AllowedStrategyRefs: []DisplaySafeRef{"strategy:tool_loop"},
	}
	needsApproval := BuildManagedObjectiveProjection(input)
	if needsApproval.Status != HostActionRequiresApproval ||
		needsApproval.FailureClass != FailureApprovalRequired ||
		needsApproval.NextHostAction != "request_host_approval" ||
		needsApproval.Ready {
		t.Fatalf("needs approval projection = %#v", needsApproval)
	}
	if !managedObjectiveMissingInputContains(needsApproval.MissingInputs, "host:managed_objective_approval") ||
		!managedObjectiveBoundaryContains(needsApproval.Boundaries, "managed_objective_requires_host_approval") {
		t.Fatalf("approval missing/boundaries = %#v / %#v", needsApproval.MissingInputs, needsApproval.Boundaries)
	}

	input.Approved = true
	approvedWithoutRef := BuildManagedObjectiveProjection(input)
	if approvedWithoutRef.Status != HostActionReviewRequired ||
		approvedWithoutRef.FailureClass != FailureEvidenceMissing ||
		approvedWithoutRef.NextHostAction != "provide_host_approval_ref" ||
		approvedWithoutRef.Ready {
		t.Fatalf("approved without ref projection = %#v", approvedWithoutRef)
	}
	if !managedObjectiveMissingInputContains(approvedWithoutRef.MissingInputs, "host:approval_ref") {
		t.Fatalf("expected approval ref missing, got %#v", approvedWithoutRef.MissingInputs)
	}

	input.ApprovalRefs = []DisplaySafeRef{"approval:approved"}
	ready := BuildManagedObjectiveProjection(input)
	if ready.Status != HostActionReady ||
		ready.FailureClass != FailureNone ||
		!ready.Ready ||
		ready.Ledger.LedgerRef != "ledger:approved" ||
		len(ready.AllowedStrategyRefs) != 1 ||
		ready.AllowedStrategyRefs[0] != "strategy:tool_loop" ||
		ready.NextHostAction != "host_may_plan_managed_objective" {
		t.Fatalf("ready projection = %#v", ready)
	}
	if !managedObjectiveBoundaryContains(ready.Boundaries, "managed_objective_contract_ready") ||
		!managedObjectiveBoundaryContains(ready.Ledger.Boundaries, "host_owned_durable_write") {
		t.Fatalf("ready boundaries = %#v / %#v", ready.Boundaries, ready.Ledger.Boundaries)
	}
}

func TestBuildManagedObjectiveReplannerProjectionCandidateAndBlocked(t *testing.T) {
	managed := BuildManagedObjectiveProjection(ManagedObjectiveProjectionInput{
		Activation: ActivationManaged,
		Frame: ObjectiveFrame{
			ID:              "objective:replan",
			UserGoalDigest:  "sha256:goal",
			SuccessCriteria: []string{"produce evidence"},
		},
		LedgerRef: "ledger:replan",
		PolicyRefs: []DisplaySafeRef{
			"contract:intensity_gate",
			"contract:budget",
			"contract:approval_policy",
			"contract:strategy_scope",
			"contract:redaction_policy",
		},
		AllowedStrategyRefs: []DisplaySafeRef{"strategy:tool_loop", "strategy:workflow"},
		Approved:            true,
		ApprovalRefs:        []DisplaySafeRef{"approval:replan"},
	})
	if !managed.Ready {
		t.Fatalf("managed objective should be ready for replanner test: %#v", managed)
	}
	candidate := BuildManagedObjectiveReplannerProjection(ManagedObjectiveReplannerInput{
		Activation:       ActivationManaged,
		ManagedObjective: managed,
		Verification: VerificationResult{
			Status:        VerificationPartial,
			FailureClass:  FailureEvidenceMissing,
			FailureReason: "missing host evidence",
			EvidenceRefs:  []EvidenceRef{{Ref: "evidence:partial", Kind: "verification", Strength: EvidenceWeak}},
		},
	})
	if candidate.Status != "candidate" ||
		candidate.FailureClass != FailureEvidenceMissing ||
		candidate.NextHostAction != "request_host_replanner_decision" ||
		candidate.RunnerEffect != "none" ||
		candidate.PromptEffect != "none" ||
		len(candidate.Candidates) != 2 {
		t.Fatalf("candidate replanner projection = %#v", candidate)
	}
	if candidate.Proposal.Kind != "managed_objective_replan" ||
		candidate.Proposal.Status != HostActionRequiresApproval ||
		!candidate.Proposal.RequiresApproval ||
		len(candidate.Proposal.ActionRefs) != 2 {
		t.Fatalf("candidate proposal = %#v", candidate.Proposal)
	}
	for _, want := range []Boundary{"proposal_only", "no_runner_dispatch", "host_must_apply_strategy"} {
		if !managedObjectiveBoundaryContains(candidate.Boundaries, want) {
			t.Fatalf("expected boundary %q, got %#v", want, candidate.Boundaries)
		}
	}
	if !managedObjectiveDisplaySafeRefContains(candidate.DecisionBasis, "managed_replanner_candidate") ||
		!managedObjectiveDisplaySafeRefContains(candidate.DecisionBasis, "verification_status:partial") {
		t.Fatalf("decision basis = %#v", candidate.DecisionBasis)
	}

	blocked := BuildManagedObjectiveReplannerProjection(ManagedObjectiveReplannerInput{
		Activation:       ActivationManaged,
		ManagedObjective: managed,
		Verification: VerificationResult{
			Status:        VerificationBlocked,
			FailureClass:  FailureBudgetExhausted,
			FailureReason: "budget exhausted",
		},
	})
	if blocked.Status != "blocked" ||
		blocked.FailureClass != FailureBudgetExhausted ||
		blocked.NextHostAction != "request_host_policy_or_budget_review" ||
		len(blocked.Candidates) != 0 ||
		blocked.Proposal.Kind != "" {
		t.Fatalf("blocked replanner projection = %#v", blocked)
	}
}

func TestBuildManagedObjectiveReplannerProjectionRequiresReadyManagedObjective(t *testing.T) {
	notReady := BuildManagedObjectiveReplannerProjection(ManagedObjectiveReplannerInput{
		Activation: ActivationManaged,
		ManagedObjective: ManagedObjectiveProjection{
			Activation:    ActivationManaged,
			Status:        HostActionNotReady,
			FailureClass:  FailureConfigMissing,
			MissingInputs: []MissingInput{"host:success_criteria"},
		},
		Verification: VerificationResult{Status: VerificationPartial},
	})
	if notReady.Status != "not_candidate" ||
		notReady.FailureClass != FailureConfigMissing ||
		notReady.NextHostAction != "provide_managed_objective_contract" ||
		!managedObjectiveMissingInputContains(notReady.MissingInputs, "host:success_criteria") {
		t.Fatalf("not-ready replanner projection = %#v", notReady)
	}

	satisfied := BuildManagedObjectiveReplannerProjection(ManagedObjectiveReplannerInput{
		Activation: ActivationManaged,
		ManagedObjective: BuildManagedObjectiveProjection(ManagedObjectiveProjectionInput{
			Activation: ActivationManaged,
			Frame: ObjectiveFrame{
				ID:              "objective:no_action",
				UserGoalDigest:  "sha256:goal",
				SuccessCriteria: []string{"done"},
			},
			LedgerRef: "ledger:no_action",
			PolicyRefs: []DisplaySafeRef{
				"contract:intensity_gate",
				"contract:budget",
				"contract:approval_policy",
				"contract:strategy_scope",
				"contract:redaction_policy",
			},
			AllowedStrategyRefs: []DisplaySafeRef{"strategy:tool_loop"},
			Approved:            true,
			ApprovalRefs:        []DisplaySafeRef{"approval:no_action"},
		}),
		Verification: VerificationResult{Status: VerificationSatisfied},
	})
	if satisfied.Status != "no_action" ||
		satisfied.NextHostAction != "none" ||
		len(satisfied.Candidates) != 0 {
		t.Fatalf("satisfied replanner projection = %#v", satisfied)
	}
}

func TestBuildReplannerSourceProjectionOperationsCandidate(t *testing.T) {
	projection := BuildReplannerSourceProjection(ReplannerSourceInput{
		SourceKind:           ReplannerSourceOperations,
		SourceRef:            "operations:run_1",
		Producer:             "scene:agentx_operations",
		Status:               VerificationPartial,
		FailureClass:         FailureEvidenceMissing,
		FailureReason:        "metric evidence missing",
		CandidateStrategyRef: "strategy:operations_metric_refresh",
		CapabilityRefs:       []DisplaySafeRef{"capability:operations_local_metrics"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:operations_partial",
			Kind:     "metric",
			Strength: EvidenceWeak,
			Source:   "operations:run_1",
		}},
		SourceRefs:    []DisplaySafeRef{"operations_task_graph"},
		ProposalRefs:  []DisplaySafeRef{"proposal:operations_refresh"},
		MissingInputs: []MissingInput{"host:metric_observation"},
		Boundaries:    []Boundary{"scene_owned_projection"},
	})
	if projection.SourceKind != ReplannerSourceOperations ||
		projection.ControlMode != ControlModeOperations ||
		projection.Status != VerificationPartial ||
		projection.FailureClass != FailureEvidenceMissing ||
		projection.RunnerEffect != "none" ||
		projection.PromptEffect != "none" {
		t.Fatalf("operations source projection = %#v", projection)
	}
	if projection.Observation.Kind != "replanner_source" ||
		projection.Observation.Source != "operations:run_1" ||
		projection.Observation.Subject != "strategy:operations_metric_refresh" ||
		projection.Observation.Strength != EvidenceWeak {
		t.Fatalf("operations observation = %#v", projection.Observation)
	}
	if projection.Candidate.ID != "strategy:operations_metric_refresh" ||
		projection.Candidate.ControlMode != ControlModeOperations ||
		projection.Candidate.MinIntensity != IntensityL3ManagedObjective ||
		projection.Candidate.MaxIntensity != IntensityL3ManagedObjective ||
		!projection.Candidate.RequiresApproval ||
		projection.Candidate.Owner != "host" {
		t.Fatalf("operations candidate = %#v", projection.Candidate)
	}
	if projection.Proposal.Kind != "operations_replanner_source_proposal" ||
		projection.Proposal.Status != HostActionRequiresApproval ||
		!projection.Proposal.RequiresApproval ||
		!managedObjectiveDisplaySafeRefContains(projection.Proposal.ActionRefs, "strategy:operations_metric_refresh") ||
		!managedObjectiveDisplaySafeRefContains(projection.Proposal.ActionRefs, "proposal:operations_refresh") {
		t.Fatalf("operations proposal = %#v", projection.Proposal)
	}
	for _, want := range []Boundary{"proposal_only", "no_runner_dispatch", "no_apply_or_dispatch_implementation", "operations_source_projection"} {
		if !managedObjectiveBoundaryContains(projection.Boundaries, want) {
			t.Fatalf("expected boundary %q, got %#v", want, projection.Boundaries)
		}
	}
	if projection.RunnerEffect != "none" || projection.PromptEffect != "none" {
		t.Fatalf("operations replanner source must be projection-only: %#v", projection)
	}
}

func TestBuildReplannerSourceProjectionCapabilityDefaultsFromFailure(t *testing.T) {
	projection := BuildReplannerSourceProjection(ReplannerSourceInput{
		SourceKind:           "capability_install",
		SourceRef:            "capability:gap_1",
		Producer:             "scene:agentx_capability_install",
		FailureClass:         FailureCapabilityMissing,
		CandidateStrategyRef: "strategy:capability_resolution",
		ProposalRefs:         []DisplaySafeRef{"proposal:capability_install"},
		MissingInputs:        []MissingInput{"host:approval"},
	})
	if projection.SourceKind != ReplannerSourceCapability ||
		projection.ControlMode != ControlModeCapabilityResolution ||
		projection.Status != VerificationBlocked ||
		projection.FailureClass != FailureCapabilityMissing ||
		projection.NextHostAction != "request_host_replanner_decision" {
		t.Fatalf("capability source projection = %#v", projection)
	}
	if projection.Candidate.ControlMode != ControlModeCapabilityResolution ||
		projection.Proposal.Kind != "capability_resolution_proposal" ||
		projection.Proposal.Status != HostActionRequiresApproval ||
		!managedObjectiveDisplaySafeRefContains(projection.Proposal.ActionRefs, "proposal:capability_install") {
		t.Fatalf("capability candidate/proposal = %#v / %#v", projection.Candidate, projection.Proposal)
	}
	if !managedObjectiveBoundaryContains(projection.Boundaries, "capability_source_projection") {
		t.Fatalf("capability boundaries = %#v", projection.Boundaries)
	}
}

func TestBuildReplannerSourceProjectionWorkflowStrategyRef(t *testing.T) {
	projection := BuildReplannerSourceProjection(ReplannerSourceInput{
		SourceKind:           "workflow_node",
		SourceRef:            "workflow:node_7",
		Status:               VerificationFailed,
		FailureClass:         FailureVerificationFailed,
		CandidateStrategyRef: "strategy:workflow_fallback",
		MinIntensity:         IntensityL2BoundedToolLoop,
		MaxIntensity:         IntensityL3ManagedObjective,
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:workflow_failure",
			Kind:     "workflow",
			Strength: EvidenceAdequate,
			Source:   "workflow:node_7",
		}},
	})
	if projection.SourceKind != ReplannerSourceWorkflow ||
		projection.ControlMode != ControlModeWorkflow ||
		projection.Status != VerificationFailed ||
		projection.Candidate.ID != "strategy:workflow_fallback" ||
		projection.Candidate.MinIntensity != IntensityL2BoundedToolLoop ||
		projection.Candidate.MaxIntensity != IntensityL3ManagedObjective ||
		projection.Proposal.Kind != "workflow_strategy_proposal" {
		t.Fatalf("workflow source projection = %#v", projection)
	}
	if !managedObjectiveBoundaryContains(projection.Boundaries, "workflow_source_projection") ||
		!managedObjectiveBoundaryContains(projection.Proposal.Boundaries, "runner_does_not_execute_replan") {
		t.Fatalf("workflow boundaries = %#v / %#v", projection.Boundaries, projection.Proposal.Boundaries)
	}
}

func TestBuildReplannerSourceProjectionRejectsUnsafeRefs(t *testing.T) {
	projection := BuildReplannerSourceProjection(ReplannerSourceInput{
		SourceKind:           ReplannerSourceOperations,
		SourceRef:            "/tmp/raw-output.json",
		Producer:             "scene:agentx_operations",
		Status:               VerificationPartial,
		FailureClass:         FailureEvidenceWeak,
		CandidateStrategyRef: "https://example.invalid/strategy",
		SourceRefs:           []DisplaySafeRef{"operations_task_graph", "/Users/mason/raw.txt"},
		ProposalRefs:         []DisplaySafeRef{"proposal:safe", "service://user:pass@example.invalid/db"},
		EvidenceRefs: []EvidenceRef{{
			Ref:      "evidence:safe",
			Kind:     "metric",
			Strength: EvidenceWeak,
			Source:   "https://example.invalid/raw",
		}},
	})
	if projection.Status != VerificationReviewRequired ||
		projection.FailureClass != FailureEvidenceWeak ||
		projection.NextHostAction != "provide_display_safe_refs" ||
		projection.SourceRef != "" ||
		projection.Candidate.ID != "" {
		t.Fatalf("unsafe projection = %#v", projection)
	}
	if !managedObjectiveMissingInputContains(projection.MissingInputs, "host:display_safe_refs") ||
		!managedObjectiveBoundaryContains(projection.Boundaries, "raw_output_not_allowed") ||
		!managedObjectiveBoundaryContains(projection.Verification.Boundaries, "raw_output_not_allowed") ||
		!managedObjectiveBoundaryContains(projection.Proposal.Boundaries, "raw_output_not_allowed") {
		t.Fatalf("unsafe review metadata = %#v / %#v / %#v", projection.Boundaries, projection.Verification, projection.Proposal)
	}
	if managedObjectiveDisplaySafeRefContains(projection.SourceRefs, "/Users/mason/raw.txt") ||
		managedObjectiveDisplaySafeRefContains(projection.ProposalRefs, "service://user:pass@example.invalid/db") ||
		managedObjectiveDisplaySafeRefContains(projection.Proposal.ActionRefs, "https://example.invalid/strategy") {
		t.Fatalf("unsafe refs leaked = %#v / %#v / %#v", projection.SourceRefs, projection.ProposalRefs, projection.Proposal.ActionRefs)
	}
}

func TestBuildReplannerSourceProjectionRequiresSourceKind(t *testing.T) {
	projection := BuildReplannerSourceProjection(ReplannerSourceInput{
		SourceKind: "domain_specific_goal",
		SourceRef:  "host:source",
	})
	if projection.SourceKind != "" ||
		projection.Status != VerificationReviewRequired ||
		projection.FailureClass != FailureInvalidInput ||
		projection.Verification.Status != VerificationReviewRequired ||
		projection.Verification.FailureClass != FailureInvalidInput ||
		projection.NextHostAction != "provide_replanner_source_kind" {
		t.Fatalf("invalid source kind projection = %#v", projection)
	}
	if !managedObjectiveMissingInputContains(projection.MissingInputs, "controlplane:replanner_source_kind") ||
		!managedObjectiveBoundaryContains(projection.Boundaries, "replanner_source_kind_missing") {
		t.Fatalf("invalid source kind metadata = %#v / %#v", projection.MissingInputs, projection.Boundaries)
	}
}

func TestEvaluateRetryBudgetGate(t *testing.T) {
	missing := EvaluateRetryBudgetGate(BudgetGateInput{})
	if missing.Allowed ||
		missing.Status != VerificationBlocked ||
		missing.FailureClass != FailureConfigMissing ||
		missing.RetryBudgetRemaining != 0 ||
		missing.NextHostAction != "provide_retry_budget_policy" {
		t.Fatalf("missing budget = %#v", missing)
	}

	allowed := EvaluateRetryBudgetGate(BudgetGateInput{Limit: 3, Used: 1, Increment: 1, Scope: "retry:l2"})
	if !allowed.Allowed ||
		allowed.Status != VerificationSatisfied ||
		allowed.RetryBudgetRemaining != 1 ||
		len(allowed.Boundaries) != 1 ||
		allowed.Boundaries[0] != "retry_budget_available" {
		t.Fatalf("allowed budget = %#v", allowed)
	}

	exhausted := EvaluateRetryBudgetGate(BudgetGateInput{Limit: 2, Used: 2, Increment: 1})
	if exhausted.Allowed ||
		exhausted.Status != VerificationBlocked ||
		exhausted.FailureClass != FailureBudgetExhausted ||
		exhausted.NextHostAction != "return_partial_or_request_upgrade" {
		t.Fatalf("exhausted budget = %#v", exhausted)
	}
}

func managedObjectiveMissingInputContains(values []MissingInput, want MissingInput) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func managedObjectiveBoundaryContains(values []Boundary, want Boundary) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func managedObjectiveDisplaySafeRefContains(values []DisplaySafeRef, want DisplaySafeRef) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestCheckIdempotencyCreateReplayAndConflict(t *testing.T) {
	create := CheckIdempotency(IdempotencyCheckInput{
		RequestedKey: "idempotency:abc",
		ActionKind:   "ledger_apply",
	})
	if create.Operation != IdempotencyCreate ||
		!create.ReadyForCreate ||
		create.Status != HostActionReady ||
		create.NextHostAction != "host_may_apply_action" {
		t.Fatalf("create = %#v", create)
	}

	replay := CheckIdempotency(IdempotencyCheckInput{
		RequestedKey:         "idempotency:abc",
		ExistingKey:          "idempotency:abc",
		RequestedFingerprint: "sha256:one",
		ExistingFingerprint:  "sha256:one",
		ExistingRef:          "ledger:existing",
		ActionKind:           "ledger_apply",
	})
	if replay.Operation != IdempotencyReplay ||
		!replay.Replayed ||
		replay.Status != HostActionRecorded ||
		replay.ExistingRef != "ledger:existing" {
		t.Fatalf("replay = %#v", replay)
	}

	conflict := CheckIdempotency(IdempotencyCheckInput{
		RequestedKey:         "idempotency:abc",
		ExistingKey:          "idempotency:abc",
		RequestedFingerprint: "sha256:one",
		ExistingFingerprint:  "sha256:two",
	})
	if conflict.Operation != IdempotencyConflict ||
		!conflict.Conflict ||
		conflict.Status != HostActionReviewRequired ||
		conflict.FailureClass != FailureVerificationFailed ||
		conflict.NextHostAction != "review_idempotency_conflict" {
		t.Fatalf("conflict = %#v", conflict)
	}

	missing := CheckIdempotency(IdempotencyCheckInput{})
	if missing.Operation != IdempotencyKeyMissing ||
		missing.Status != HostActionBlocked ||
		missing.FailureClass != FailureInvalidInput ||
		len(missing.MissingInputs) != 1 ||
		missing.MissingInputs[0] != "host:idempotency_key" {
		t.Fatalf("missing key = %#v", missing)
	}
}

func TestSelectLatestEventAndStaleSourceCheck(t *testing.T) {
	events := []EventRef{
		{Ref: "event:old", Kind: "applied", CreatedAt: "2026-05-31T10:00:00Z", Sequence: 3},
		{Ref: "event:new_a", Kind: "applied", CreatedAt: "2026-05-31T11:00:00Z", Sequence: 1},
		{Ref: "event:new_b", Kind: "applied", CreatedAt: "2026-05-31T11:00:00Z", Sequence: 2},
		{Ref: "https://example.invalid/raw", Kind: "ignored", CreatedAt: "2026-05-31T12:00:00Z"},
	}
	latest, ok := SelectLatestEvent(events)
	if !ok || latest.Ref != "event:new_b" {
		t.Fatalf("latest = %#v ok=%v", latest, ok)
	}

	current := CheckLatestSourceEvent(latest, latest)
	if current.Status != VerificationSatisfied ||
		current.Stale ||
		current.NextHostAction != "host_may_continue" {
		t.Fatalf("current source = %#v", current)
	}

	stale := CheckLatestSourceEvent(events[0], latest)
	if stale.Status != VerificationReviewRequired ||
		!stale.Stale ||
		stale.FailureClass != FailureVerificationFailed ||
		stale.NextHostAction != "review_stale_source_event" {
		t.Fatalf("stale source = %#v", stale)
	}

	missingLatest := CheckLatestSourceEvent(events[0], EventRef{})
	if missingLatest.Status != VerificationBlocked ||
		missingLatest.FailureClass != FailureInsufficientInformation ||
		len(missingLatest.MissingInputs) != 1 ||
		missingLatest.MissingInputs[0] != "host:latest_event_ref" {
		t.Fatalf("missing latest = %#v", missingLatest)
	}
}

func TestRequireVerificationAndHostActionReview(t *testing.T) {
	result := VerificationResult{
		Status:       VerificationSatisfied,
		Satisfied:    true,
		FailureClass: FailureNone,
		EvidenceRefs: []EvidenceRef{{Ref: "evidence:ok", Kind: "summary", Strength: "adequate"}},
	}
	review := RequireVerificationReview(result, ReviewRequiredInput{
		FailureClass:   FailureEvidenceWeak,
		Boundary:       "manual_review_required",
		MissingInput:   "host:manual_review",
		Finding:        "needs host review",
		NextHostAction: "request_host_review",
	})
	if review.Status != VerificationReviewRequired ||
		review.Satisfied ||
		review.FailureClass != FailureEvidenceWeak ||
		len(review.Boundaries) != 1 ||
		review.Boundaries[0] != "manual_review_required" ||
		len(review.Findings) != 1 ||
		review.Findings[0] != "needs host review" {
		t.Fatalf("verification review = %#v", review)
	}

	proposal := HostActionProposal{Kind: "apply", Status: HostActionReady, FailureClass: FailureNone}
	hostReview := RequireHostActionReview(proposal, ReviewRequiredInput{Boundary: "stale_review_required"})
	if hostReview.Status != HostActionReviewRequired ||
		hostReview.FailureClass != FailureVerificationFailed ||
		len(hostReview.MissingInputs) != 1 ||
		hostReview.MissingInputs[0] != "host:review" {
		t.Fatalf("host action review = %#v", hostReview)
	}
}

func TestCheckLifecycleTransition(t *testing.T) {
	ready := CheckLifecycleTransition(LifecycleStageReady, LifecycleStageApplied)
	if !ready.Allowed ||
		ready.Status != HostActionReady ||
		ready.NextHostAction != "host_may_apply_action" {
		t.Fatalf("ready transition = %#v", ready)
	}

	replay := CheckLifecycleTransition(LifecycleStageApplied, LifecycleStageApplied)
	if !replay.Allowed ||
		!replay.IdempotentReplay ||
		replay.Status != HostActionRecorded ||
		replay.NextHostAction != "host_may_report_existing_state" {
		t.Fatalf("replay transition = %#v", replay)
	}

	gap := CheckLifecycleTransition(LifecycleStageReady, LifecycleStageReadback)
	if gap.Allowed ||
		gap.Status != HostActionReviewRequired ||
		gap.FailureClass != FailureInsufficientInformation ||
		len(gap.MissingInputs) != 1 ||
		gap.MissingInputs[0] != "host:lifecycle_intermediate_state" {
		t.Fatalf("gap transition = %#v", gap)
	}

	regression := CheckLifecycleTransition(LifecycleStageAudit, LifecycleStageApplied)
	if regression.Allowed ||
		regression.Status != HostActionReviewRequired ||
		regression.FailureClass != FailureVerificationFailed ||
		regression.NextHostAction != "review_lifecycle_regression" {
		t.Fatalf("regression transition = %#v", regression)
	}

	owner := checkLifecycleTransition(LifecycleStageReady, LifecycleStageApplied)
	if !reflect.DeepEqual(ready, owner) {
		t.Fatalf("deprecated compatibility wrapper drifted from owner reducer: wrapper=%#v owner=%#v", ready, owner)
	}
}
