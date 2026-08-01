package controlcontract

import "testing"

func TestNormalizeEnums(t *testing.T) {
	if got := NormalizeActivation("observe-only"); got != ActivationObserveOnly {
		t.Fatalf("NormalizeActivation observe-only = %q", got)
	}
	if got := NormalizeActivation("unexpected"); got != ActivationOff {
		t.Fatalf("NormalizeActivation unexpected = %q", got)
	}
	if got := NormalizeControlMode("capability install"); got != ControlModeCapabilityResolution {
		t.Fatalf("NormalizeControlMode capability install = %q", got)
	}
	if got := NormalizeExecutionIntensity("L2"); got != IntensityL2BoundedToolLoop {
		t.Fatalf("NormalizeExecutionIntensity L2 = %q", got)
	}
	if got := NormalizeEvidenceStrength("bad-value"); got != EvidenceWeak {
		t.Fatalf("NormalizeEvidenceStrength bad-value = %q", got)
	}
	if got := NormalizeVerificationStatus("done-ish"); got != VerificationReviewRequired {
		t.Fatalf("NormalizeVerificationStatus done-ish = %q", got)
	}
	if got := NormalizeVerificationStatus("not applicable"); got != VerificationNotApplicable {
		t.Fatalf("NormalizeVerificationStatus not applicable = %q", got)
	}
	if got := NormalizeFailureClass("credential-missing"); got != FailureCredentialMissing {
		t.Fatalf("NormalizeFailureClass credential-missing = %q", got)
	}
	if got := NormalizeFailureClass("unknown_failure"); got != FailureInternalError {
		t.Fatalf("NormalizeFailureClass unknown_failure = %q", got)
	}
}

func TestObjectiveFrameNormalizeAndClone(t *testing.T) {
	frame := ObjectiveFrame{
		ID:              " objective:one ",
		ControlMode:     "workflow",
		Intensity:       "l3",
		SuccessCriteria: []string{" complete ", "complete", ""},
		Constraints:     []string{"no raw output", "/tmp/agentx.env"},
		RequiredEvidence: []EvidenceRef{
			{Ref: "evidence:metric_1", Kind: " metric ", Strength: "strong", Source: "host:collector"},
			{Ref: "https://example.com/raw", Kind: "source", Strength: "adequate"},
		},
		CandidateCapabilities: []DisplaySafeRef{"capability:sample_runtime", "capability:sample_runtime", "service://user:pass@example.invalid/db"},
		SourceContext:         []DisplaySafeRef{"host:context", "/tmp/agentx.env", "host:context"},
		Boundaries:            []Boundary{" approval required ", "approval_required"},
		MissingInputs:         []MissingInput{" host:approval ", "host:approval"},
	}

	normalized := frame.Normalize()
	if normalized.ContractVersion != ContractVersion {
		t.Fatalf("contract version = %q", normalized.ContractVersion)
	}
	if normalized.ID != "objective:one" {
		t.Fatalf("id = %q", normalized.ID)
	}
	if normalized.ControlMode != ControlModeWorkflow || normalized.Intensity != IntensityL3ManagedObjective {
		t.Fatalf("mode/intensity = %q/%q", normalized.ControlMode, normalized.Intensity)
	}
	if len(normalized.SuccessCriteria) != 1 || normalized.SuccessCriteria[0] != "complete" {
		t.Fatalf("success criteria = %#v", normalized.SuccessCriteria)
	}
	if len(normalized.Constraints) != 1 || normalized.Constraints[0] != "no raw output" {
		t.Fatalf("constraints = %#v", normalized.Constraints)
	}
	if len(normalized.RequiredEvidence) != 1 || normalized.RequiredEvidence[0].Ref != "evidence:metric_1" {
		t.Fatalf("required evidence = %#v", normalized.RequiredEvidence)
	}
	if len(normalized.CandidateCapabilities) != 1 || normalized.CandidateCapabilities[0] != "capability:sample_runtime" {
		t.Fatalf("candidate capabilities = %#v", normalized.CandidateCapabilities)
	}
	if len(normalized.SourceContext) != 1 || normalized.SourceContext[0] != "host:context" {
		t.Fatalf("source context = %#v", normalized.SourceContext)
	}
	if len(normalized.Boundaries) != 1 || normalized.Boundaries[0] != "approval_required" {
		t.Fatalf("boundaries = %#v", normalized.Boundaries)
	}
	if len(normalized.MissingInputs) != 1 || normalized.MissingInputs[0] != "host:approval" {
		t.Fatalf("missing inputs = %#v", normalized.MissingInputs)
	}

	clone := CloneObjectiveFrame(normalized)
	clone.SuccessCriteria[0] = "changed"
	clone.RequiredEvidence[0].Ref = "evidence:changed"
	clone.SourceContext[0] = "host:changed"
	if normalized.SuccessCriteria[0] != "complete" ||
		normalized.RequiredEvidence[0].Ref != "evidence:metric_1" ||
		normalized.SourceContext[0] != "host:context" {
		t.Fatalf("clone mutated original: %#v", normalized)
	}
}

func TestAttemptLedgerAndHostActionCloneSlices(t *testing.T) {
	patch := AttemptLedgerPatch{
		ObjectiveID: "objective-1",
		LedgerRef:   "ledger:one",
		Attempts: []AttemptSummary{
			{
				Ref:          "attempt:one",
				Status:       "partial",
				EvidenceRefs: []EvidenceRef{{Ref: "evidence:one", Kind: "tool", Strength: "adequate"}},
				Boundaries:   []Boundary{"same_strategy_only"},
			},
		},
		EvidenceRefs:   []EvidenceRef{{Ref: "evidence:ledger", Kind: "summary", Strength: "adequate"}},
		Boundaries:     []Boundary{"no_strategy_switch"},
		MissingInputs:  []MissingInput{"host:approval"},
		NextHostAction: "host_may_dispatch",
	}
	normalizedPatch := patch.Normalize()
	clonePatch := CloneAttemptLedgerPatch(normalizedPatch)
	clonePatch.Attempts[0].EvidenceRefs[0].Ref = "evidence:changed"
	clonePatch.EvidenceRefs[0].Ref = "evidence:changed"
	clonePatch.Boundaries[0] = "changed"
	if normalizedPatch.Attempts[0].EvidenceRefs[0].Ref != "evidence:one" ||
		normalizedPatch.EvidenceRefs[0].Ref != "evidence:ledger" ||
		normalizedPatch.Boundaries[0] != "no_strategy_switch" {
		t.Fatalf("ledger clone mutated original: %#v", normalizedPatch)
	}

	proposal := HostActionProposal{
		Kind:             " retry same strategy ",
		Status:           "ready",
		RequiresApproval: true,
		ApprovalRefs:     []DisplaySafeRef{"host:approval"},
		ActionRefs:       []DisplaySafeRef{"action:retry"},
		EvidenceRefs:     []EvidenceRef{{Ref: "evidence:retry", Kind: "retry", Strength: "adequate"}},
	}
	normalizedProposal := proposal.Normalize()
	cloneProposal := CloneHostActionProposal(normalizedProposal)
	cloneProposal.ApprovalRefs[0] = "host:changed"
	cloneProposal.ActionRefs[0] = "action:changed"
	cloneProposal.EvidenceRefs[0].Ref = "evidence:changed"
	if normalizedProposal.ApprovalRefs[0] != "host:approval" ||
		normalizedProposal.ActionRefs[0] != "action:retry" ||
		normalizedProposal.EvidenceRefs[0].Ref != "evidence:retry" {
		t.Fatalf("proposal clone mutated original: %#v", normalizedProposal)
	}
}

func TestManagedObjectiveProjectionNormalizeAndClone(t *testing.T) {
	projection := ManagedObjectiveProjection{
		Activation: ActivationManaged,
		Status:     "ready",
		Frame: ObjectiveFrame{
			ID:              "objective:managed_1",
			UserGoalDigest:  "sha256:abc123",
			ControlMode:     "objective",
			Intensity:       "l3",
			SuccessCriteria: []string{"collect evidence", "collect evidence"},
			SourceContext:   []DisplaySafeRef{"host:context", "https://example.invalid/raw"},
		},
		Ledger: AttemptLedgerPatch{
			ObjectiveID: "objective:managed_1",
			LedgerRef:   "ledger:managed_1",
			Attempts: []AttemptSummary{{
				Ref:    "attempt:one",
				Status: VerificationPartial,
			}},
		},
		ApprovalRefs:        []DisplaySafeRef{"approval:managed_1", "approval:managed_1", "/tmp/raw"},
		PolicyRefs:          []DisplaySafeRef{"contract:intensity_gate", "service://user:pass@example.invalid/db"},
		AllowedStrategyRefs: []DisplaySafeRef{"strategy:tool_loop", "strategy:tool_loop", "/tmp/raw"},
		EvidenceRefs:        []EvidenceRef{{Ref: "evidence:managed_1", Kind: "metric", Strength: "adequate"}},
		Boundaries:          []Boundary{"diagnostics_only", "diagnostics_only"},
	}

	normalized := projection.Normalize()
	if normalized.ContractVersion != ContractVersion || !normalized.Projected {
		t.Fatalf("unexpected projection header: %#v", normalized)
	}
	if !normalized.Ready ||
		normalized.Status != HostActionReady ||
		normalized.RunnerEffect != "none" ||
		normalized.PromptEffect != "none" ||
		!normalized.RequiresApproval {
		t.Fatalf("unexpected readiness/effects: %#v", normalized)
	}
	if len(normalized.ApprovalRefs) != 1 || normalized.ApprovalRefs[0] != "approval:managed_1" {
		t.Fatalf("approval refs = %#v", normalized.ApprovalRefs)
	}
	if len(normalized.PolicyRefs) != 1 || normalized.PolicyRefs[0] != "contract:intensity_gate" {
		t.Fatalf("policy refs = %#v", normalized.PolicyRefs)
	}
	if len(normalized.AllowedStrategyRefs) != 1 || normalized.AllowedStrategyRefs[0] != "strategy:tool_loop" {
		t.Fatalf("allowed strategy refs = %#v", normalized.AllowedStrategyRefs)
	}
	if len(normalized.Frame.SourceContext) != 1 || normalized.Frame.SourceContext[0] != "host:context" {
		t.Fatalf("source context = %#v", normalized.Frame.SourceContext)
	}

	clone := CloneManagedObjectiveProjection(normalized)
	clone.Frame.SuccessCriteria[0] = "changed"
	clone.Ledger.Attempts[0].Ref = "attempt:changed"
	clone.ApprovalRefs[0] = "approval:changed"
	clone.AllowedStrategyRefs[0] = "strategy:changed"
	if normalized.Frame.SuccessCriteria[0] != "collect evidence" ||
		normalized.Ledger.Attempts[0].Ref != "attempt:one" ||
		normalized.ApprovalRefs[0] != "approval:managed_1" ||
		normalized.AllowedStrategyRefs[0] != "strategy:tool_loop" {
		t.Fatalf("clone mutated original: %#v", normalized)
	}

	rawLoaded := normalized
	rawLoaded.RawOutputLoaded = true
	rawLoaded = rawLoaded.Normalize()
	if rawLoaded.Ready ||
		rawLoaded.Status != HostActionReviewRequired ||
		rawLoaded.FailureClass != FailureEvidenceWeak ||
		len(rawLoaded.MissingInputs) != 1 ||
		rawLoaded.MissingInputs[0] != "host:display_safe_refs" {
		t.Fatalf("raw-output downgrade = %#v", rawLoaded)
	}
}

func TestDisplaySafeRefRejectsRawOutputs(t *testing.T) {
	valid := []string{
		"host:approval",
		"follow_up_result:abc_123",
		"control_plane.follow_up_ledger",
		"evidence:metric-1",
	}
	for _, value := range valid {
		if _, ok := NormalizeDisplaySafeRef(value); !ok {
			t.Fatalf("expected display-safe ref %q to be accepted", value)
		}
	}

	invalid := []string{
		"",
		"foo bar",
		"https://example.com/raw",
		"/tmp/agentx.env",
		"service://user:pass@example.invalid/db",
		"api_key=abc123",
		"C:\\Users\\mason\\.env",
	}
	for _, value := range invalid {
		if ref, ok := NormalizeDisplaySafeRef(value); ok {
			t.Fatalf("expected raw ref %q to be rejected, got %q", value, ref)
		}
	}

	refs := DisplaySafeRefs([]string{"host:approval", "https://example.com/raw", "host:approval", "evidence:one"})
	if len(refs) != 2 || refs[0] != "host:approval" || refs[1] != "evidence:one" {
		t.Fatalf("DisplaySafeRefs = %#v", refs)
	}
}

func TestVerifyDisplaySafeOnlyConformance(t *testing.T) {
	ok := VerifyDisplaySafeOnly(false, []string{"host:approval", "evidence:one"})
	if ok.Status != VerificationSatisfied || !ok.Satisfied || ok.FailureClass != FailureNone {
		t.Fatalf("safe verification = %#v", ok)
	}
	if len(ok.EvidenceRefs) != 2 || ok.EvidenceRefs[0].Source != "host:control_plane" {
		t.Fatalf("safe evidence refs = %#v", ok.EvidenceRefs)
	}

	blocked := VerifyDisplaySafeOnly(true, []string{"host:approval"})
	if blocked.Status != VerificationBlocked ||
		blocked.Satisfied ||
		blocked.FailureClass != FailureEvidenceWeak ||
		blocked.NextHostAction != "provide_display_safe_refs" {
		t.Fatalf("raw-output verification = %#v", blocked)
	}
	if len(blocked.Boundaries) != 1 || blocked.Boundaries[0] != "raw_output_not_allowed" {
		t.Fatalf("raw-output boundaries = %#v", blocked.Boundaries)
	}
	if len(blocked.MissingInputs) != 1 || blocked.MissingInputs[0] != "host:display_safe_refs" {
		t.Fatalf("raw-output missing inputs = %#v", blocked.MissingInputs)
	}

	unsafeRef := VerifyDisplaySafeOnly(false, []string{"https://example.com/raw"})
	if unsafeRef.Status != VerificationBlocked ||
		unsafeRef.FailureClass != FailureEvidenceWeak ||
		len(unsafeRef.EvidenceRefs) != 0 {
		t.Fatalf("unsafe ref verification = %#v", unsafeRef)
	}

	manual := VerificationResult{
		Status:          VerificationSatisfied,
		Satisfied:       true,
		FailureClass:    FailureNone,
		RawOutputLoaded: true,
	}.Normalize()
	if manual.Status != VerificationReviewRequired ||
		manual.Satisfied ||
		manual.FailureClass != FailureEvidenceWeak ||
		manual.NextHostAction != "provide_display_safe_refs" {
		t.Fatalf("manual raw-output normalization = %#v", manual)
	}
}

func TestObservationNormalizePreservesMetricShapeWithoutRawLeak(t *testing.T) {
	observation := Observation{
		Kind:       " metric ",
		Source:     "host:metrics",
		Subject:    "system:local",
		Name:       "cpu_usage",
		Value:      "37.2",
		Unit:       "percent",
		Strength:   "strong",
		ObservedAt: "2026-05-31T12:00:00Z",
		EvidenceRefs: []EvidenceRef{
			{Ref: "evidence:cpu", Kind: "metric", Strength: "strong", Source: "host:metrics"},
			{Ref: "/tmp/agentx-raw.txt", Kind: "raw", Strength: "weak"},
		},
		DisplaySafeRefs: []DisplaySafeRef{"metric:cpu", "https://example.com/raw"},
	}

	normalized := observation.Normalize()
	if normalized.Kind != "metric" ||
		normalized.Source != "host:metrics" ||
		normalized.Subject != "system:local" ||
		normalized.Strength != EvidenceStrong {
		t.Fatalf("normalized observation = %#v", normalized)
	}
	if len(normalized.EvidenceRefs) != 1 || normalized.EvidenceRefs[0].Ref != "evidence:cpu" {
		t.Fatalf("observation evidence refs = %#v", normalized.EvidenceRefs)
	}
	if len(normalized.DisplaySafeRefs) != 1 || normalized.DisplaySafeRefs[0] != "metric:cpu" {
		t.Fatalf("observation display refs = %#v", normalized.DisplaySafeRefs)
	}
}
