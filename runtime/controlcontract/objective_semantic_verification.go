package controlcontract

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

type ObjectiveSemanticCoverageStatus string

const (
	ObjectiveSemanticCoverageUnspecified   ObjectiveSemanticCoverageStatus = "unspecified"
	ObjectiveSemanticCoverageCovered       ObjectiveSemanticCoverageStatus = "covered"
	ObjectiveSemanticCoverageMissing       ObjectiveSemanticCoverageStatus = "missing"
	ObjectiveSemanticCoverageUncertain     ObjectiveSemanticCoverageStatus = "uncertain"
	ObjectiveSemanticCoverageNotApplicable ObjectiveSemanticCoverageStatus = "not_applicable"
)

func NormalizeObjectiveSemanticCoverageStatus(raw string) ObjectiveSemanticCoverageStatus {
	switch normalizeEnumToken(raw) {
	case "covered", "satisfied", "met":
		return ObjectiveSemanticCoverageCovered
	case "missing", "not_covered", "unmet":
		return ObjectiveSemanticCoverageMissing
	case "uncertain", "needs_review", "weak":
		return ObjectiveSemanticCoverageUncertain
	case "not_applicable", "not_app", "na", "n_a":
		return ObjectiveSemanticCoverageNotApplicable
	default:
		return ObjectiveSemanticCoverageUnspecified
	}
}

type ObjectiveSemanticCoverageAssessment struct {
	ContractVersion     string                          `json:"contract_version,omitempty"`
	CriterionRef        DisplaySafeRef                  `json:"criterion_ref,omitempty"`
	Status              ObjectiveSemanticCoverageStatus `json:"status,omitempty"`
	RequiredEvidence    []EvidenceRef                   `json:"required_evidence,omitempty"`
	CoveredEvidenceRefs []EvidenceRef                   `json:"covered_evidence_refs,omitempty"`
	MissingEvidence     []EvidenceRef                   `json:"missing_evidence,omitempty"`
	Findings            []string                        `json:"findings,omitempty"`
	MissingInputs       []MissingInput                  `json:"missing_inputs,omitempty"`
	Boundaries          []Boundary                      `json:"boundaries,omitempty"`
}

func CloneObjectiveSemanticCoverageAssessment(in ObjectiveSemanticCoverageAssessment) ObjectiveSemanticCoverageAssessment {
	out := in
	out.RequiredEvidence = cloneEvidenceRefs(in.RequiredEvidence)
	out.CoveredEvidenceRefs = cloneEvidenceRefs(in.CoveredEvidenceRefs)
	out.MissingEvidence = cloneEvidenceRefs(in.MissingEvidence)
	out.Findings = cloneStringSlice(in.Findings)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (a ObjectiveSemanticCoverageAssessment) Clone() ObjectiveSemanticCoverageAssessment {
	return CloneObjectiveSemanticCoverageAssessment(a)
}

func (a ObjectiveSemanticCoverageAssessment) Normalize() ObjectiveSemanticCoverageAssessment {
	out := a.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.CriterionRef = normalizeOneDisplaySafeRef(out.CriterionRef)
	out.Status = NormalizeObjectiveSemanticCoverageStatus(string(out.Status))
	out.RequiredEvidence = normalizeEvidenceRefs(out.RequiredEvidence)
	out.CoveredEvidenceRefs = normalizeEvidenceRefs(out.CoveredEvidenceRefs)
	out.MissingEvidence = normalizeEvidenceRefs(out.MissingEvidence)
	out.Findings = normalizeStringList(out.Findings)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSemanticVerificationAdvice struct {
	ContractVersion string                                `json:"contract_version,omitempty"`
	AdviceRef       DisplaySafeRef                        `json:"advice_ref,omitempty"`
	SuggestedStatus VerificationStatus                    `json:"suggested_status,omitempty"`
	Coverage        []ObjectiveSemanticCoverageAssessment `json:"coverage,omitempty"`
	EvidenceRefs    []EvidenceRef                         `json:"evidence_refs,omitempty"`
	MissingInputs   []MissingInput                        `json:"missing_inputs,omitempty"`
	Findings        []string                              `json:"findings,omitempty"`
	Boundaries      []Boundary                            `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction                        `json:"next_host_action,omitempty"`
	RawOutputLoaded bool                                  `json:"raw_output_loaded"`
}

func CloneObjectiveSemanticVerificationAdvice(in ObjectiveSemanticVerificationAdvice) ObjectiveSemanticVerificationAdvice {
	out := in
	out.Coverage = cloneObjectiveSemanticCoverageAssessments(in.Coverage)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.Findings = cloneStringSlice(in.Findings)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (a ObjectiveSemanticVerificationAdvice) Clone() ObjectiveSemanticVerificationAdvice {
	return CloneObjectiveSemanticVerificationAdvice(a)
}

func (a ObjectiveSemanticVerificationAdvice) Normalize() ObjectiveSemanticVerificationAdvice {
	out := a.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.AdviceRef = normalizeOneDisplaySafeRef(out.AdviceRef)
	out.SuggestedStatus = NormalizeVerificationStatus(string(out.SuggestedStatus))
	if out.SuggestedStatus == VerificationNotEvaluated {
		out.SuggestedStatus = VerificationReviewRequired
	}
	out.Coverage = normalizeObjectiveSemanticCoverageAssessments(out.Coverage)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Findings = normalizeStringList(out.Findings)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	return out
}

type ObjectiveSemanticVerificationJSONDecodeInput struct {
	RawJSON          []byte         `json:"-"`
	AdviceRef        DisplaySafeRef `json:"advice_ref,omitempty"`
	SourceRef        DisplaySafeRef `json:"source_ref,omitempty"`
	RequiredEvidence []EvidenceRef  `json:"required_evidence,omitempty"`
	Boundaries       []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded  bool           `json:"raw_output_loaded"`
}

type ObjectiveSemanticVerificationJSONDecodeReport struct {
	ContractVersion string                              `json:"contract_version,omitempty"`
	Decoded         bool                                `json:"decoded"`
	Available       bool                                `json:"available"`
	Status          VerificationStatus                  `json:"status,omitempty"`
	Mode            string                              `json:"mode,omitempty"`
	RunnerEffect    string                              `json:"runner_effect,omitempty"`
	PromptEffect    string                              `json:"prompt_effect,omitempty"`
	AdviceRef       DisplaySafeRef                      `json:"advice_ref,omitempty"`
	SourceRef       DisplaySafeRef                      `json:"source_ref,omitempty"`
	Advice          ObjectiveSemanticVerificationAdvice `json:"advice,omitempty"`
	FailureClass    FailureClass                        `json:"failure_class,omitempty"`
	MissingInputs   []MissingInput                      `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                          `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction                      `json:"next_host_action,omitempty"`
	RawOutputLoaded bool                                `json:"raw_output_loaded"`
}

func BuildObjectiveSemanticVerificationFromJSON(input ObjectiveSemanticVerificationJSONDecodeInput) ObjectiveSemanticVerificationJSONDecodeReport {
	result := baseObjectiveSemanticVerificationJSONDecodeReport(input)
	if objectiveSemanticVerificationJSONDecodeInputUnsafe(input) || input.RawOutputLoaded {
		return objectiveSemanticVerificationJSONDecodeReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(input.RawJSON)) == 0 {
		return objectiveSemanticVerificationJSONDecodeReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_semantic_verification_json", "provide_objective_semantic_verification_json", "objective_semantic_verification_json_missing")
	}
	advice, ok, boundary := decodeObjectiveSemanticVerificationJSON(input.RawJSON)
	if !ok {
		result = objectiveSemanticVerificationJSONDecodeReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_semantic_verification_json", "provide_objective_semantic_verification_json", boundary)
		result.Boundaries = AppendBoundaries(result.Boundaries, "semantic_verification_deterministic_blocked_fallback")
		return result.Normalize()
	}
	advice = objectiveSemanticVerificationApplyAdviceDefaults(advice, input).Normalize()
	result.Decoded = true
	result.Advice = advice
	result.Status = objectiveSemanticVerificationReportStatus(advice)
	result.FailureClass = objectiveSemanticVerificationFailureClass(advice)
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, advice.MissingInputs, objectiveSemanticCoverageMissingInputs(advice.Coverage))
	result.Boundaries = MergeBoundaries(result.Boundaries, advice.Boundaries, objectiveSemanticCoverageBoundaries(advice.Coverage))
	result.NextHostAction = firstNextHostAction(advice.NextHostAction, objectiveSemanticVerificationNextAction(advice))
	if result.FailureClass == FailureNone {
		result.NextHostAction = firstNextHostAction(result.NextHostAction, "run_deterministic_verification_gate")
	}
	return result.Normalize()
}

func (r ObjectiveSemanticVerificationJSONDecodeReport) Clone() ObjectiveSemanticVerificationJSONDecodeReport {
	out := r
	out.Advice = r.Advice.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveSemanticVerificationJSONDecodeReport) Normalize() ObjectiveSemanticVerificationJSONDecodeReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_semantic_verification_json_decode"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.AdviceRef = normalizeOneDisplaySafeRef(out.AdviceRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	if !objectiveSemanticVerificationAdviceIsEmpty(out.Advice) {
		out.Advice = out.Advice.Normalize()
	}
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Advice.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

type ObjectiveSemanticVerifier interface {
	VerifyObjectiveSemantics(context.Context, ObjectiveSemanticVerifierRequest) (ObjectiveSemanticVerifierResponse, error)
}

type ObjectiveSemanticVerifierFunc func(context.Context, ObjectiveSemanticVerifierRequest) (ObjectiveSemanticVerifierResponse, error)

func (f ObjectiveSemanticVerifierFunc) VerifyObjectiveSemantics(ctx context.Context, request ObjectiveSemanticVerifierRequest) (ObjectiveSemanticVerifierResponse, error) {
	return f(ctx, request)
}

type ObjectiveSemanticVerifierRequest struct {
	RequestRef        DisplaySafeRef                  `json:"request_ref,omitempty"`
	Spec              ObjectiveSpec                   `json:"spec,omitempty"`
	Graph             ObjectiveGraph                  `json:"graph,omitempty"`
	Verification      ObjectiveVerificationGateResult `json:"verification,omitempty"`
	AttemptLedgerRefs []DisplaySafeRef                `json:"attempt_ledger_refs,omitempty"`
	EvidenceRefs      []EvidenceRef                   `json:"evidence_refs,omitempty"`
	PolicyRefs        []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Boundaries        []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded   bool                            `json:"raw_output_loaded"`
}

func CloneObjectiveSemanticVerifierRequest(in ObjectiveSemanticVerifierRequest) ObjectiveSemanticVerifierRequest {
	out := in
	out.Spec = in.Spec.Clone()
	out.Graph = in.Graph.Clone()
	out.Verification = in.Verification.Clone()
	out.AttemptLedgerRefs = cloneDisplaySafeRefs(in.AttemptLedgerRefs)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveSemanticVerifierRequest) Clone() ObjectiveSemanticVerifierRequest {
	return CloneObjectiveSemanticVerifierRequest(r)
}

func (r ObjectiveSemanticVerifierRequest) Normalize() ObjectiveSemanticVerifierRequest {
	out := r.Clone()
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.Spec = out.Spec.Normalize()
	out.Graph = out.Graph.Normalize()
	out.Verification = out.Verification.Normalize()
	out.AttemptLedgerRefs = normalizeDisplaySafeRefs(out.AttemptLedgerRefs)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSemanticVerifierResponse struct {
	ResponseRef     DisplaySafeRef                      `json:"response_ref,omitempty"`
	Advice          ObjectiveSemanticVerificationAdvice `json:"advice,omitempty"`
	AdviceJSON      []byte                              `json:"-"`
	Boundaries      []Boundary                          `json:"boundaries,omitempty"`
	RawOutputLoaded bool                                `json:"raw_output_loaded"`
}

func CloneObjectiveSemanticVerifierResponse(in ObjectiveSemanticVerifierResponse) ObjectiveSemanticVerifierResponse {
	out := in
	out.Advice = in.Advice.Clone()
	out.AdviceJSON = append([]byte(nil), in.AdviceJSON...)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveSemanticVerifierResponse) Clone() ObjectiveSemanticVerifierResponse {
	return CloneObjectiveSemanticVerifierResponse(r)
}

func (r ObjectiveSemanticVerifierResponse) Normalize() ObjectiveSemanticVerifierResponse {
	out := r.Clone()
	out.ResponseRef = normalizeOneDisplaySafeRef(out.ResponseRef)
	if !objectiveSemanticVerificationAdviceIsEmpty(out.Advice) {
		out.Advice = out.Advice.Normalize()
	}
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSemanticVerificationInput struct {
	Enabled    bool                             `json:"enabled"`
	Verifier   ObjectiveSemanticVerifier        `json:"-"`
	Request    ObjectiveSemanticVerifierRequest `json:"request,omitempty"`
	AdviceRef  DisplaySafeRef                   `json:"advice_ref,omitempty"`
	SourceRef  DisplaySafeRef                   `json:"source_ref,omitempty"`
	Boundaries []Boundary                       `json:"boundaries,omitempty"`
}

type ObjectiveSemanticVerificationReport struct {
	ContractVersion             string                                        `json:"contract_version,omitempty"`
	Evaluated                   bool                                          `json:"evaluated"`
	Available                   bool                                          `json:"available"`
	Status                      VerificationStatus                            `json:"status,omitempty"`
	Mode                        string                                        `json:"mode,omitempty"`
	RunnerEffect                string                                        `json:"runner_effect,omitempty"`
	PromptEffect                string                                        `json:"prompt_effect,omitempty"`
	RequestRef                  DisplaySafeRef                                `json:"request_ref,omitempty"`
	SourceRef                   DisplaySafeRef                                `json:"source_ref,omitempty"`
	VerifierCalled              bool                                          `json:"verifier_called"`
	DecodeAttempted             bool                                          `json:"decode_attempted"`
	SuggestedStatus             VerificationStatus                            `json:"suggested_status,omitempty"`
	Advice                      ObjectiveSemanticVerificationAdvice           `json:"advice,omitempty"`
	JSONDecode                  ObjectiveSemanticVerificationJSONDecodeReport `json:"json_decode,omitempty"`
	Coverage                    []ObjectiveSemanticCoverageAssessment         `json:"coverage,omitempty"`
	EvidenceRefs                []EvidenceRef                                 `json:"evidence_refs,omitempty"`
	MissingInputs               []MissingInput                                `json:"missing_inputs,omitempty"`
	Findings                    []string                                      `json:"findings,omitempty"`
	FailureClass                FailureClass                                  `json:"failure_class,omitempty"`
	Boundaries                  []Boundary                                    `json:"boundaries,omitempty"`
	NextHostAction              NextHostAction                                `json:"next_host_action,omitempty"`
	ReadyForDeterministicReview bool                                          `json:"ready_for_deterministic_review"`
	RawOutputLoaded             bool                                          `json:"raw_output_loaded"`
}

func BuildObjectiveSemanticVerification(ctx context.Context, input ObjectiveSemanticVerificationInput) ObjectiveSemanticVerificationReport {
	result := baseObjectiveSemanticVerificationReport(input)
	if objectiveSemanticVerificationInputUnsafe(input) {
		return objectiveSemanticVerificationReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !input.Enabled {
		return objectiveSemanticVerificationReportBlock(result, VerificationNotApplicable, FailureNone, "host:objective_semantic_verifier_enabled", "continue_objective_runtime_loop", "objective_semantic_verifier_disabled")
	}
	if input.Verifier == nil {
		return objectiveSemanticVerificationReportBlock(result, VerificationBlocked, FailureHostAdapterMissing, "host:objective_semantic_verifier", "provide_objective_semantic_verifier", "objective_semantic_verifier_missing")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	request := input.Request.Normalize()
	result.VerifierCalled = true
	response, err := input.Verifier.VerifyObjectiveSemantics(ctx, request)
	if err != nil {
		result = objectiveSemanticVerificationReportBlock(result, VerificationBlocked, FailureExternalDependencyUnavailable, "host:objective_semantic_verifier_result", "review_semantic_verification", "objective_semantic_verifier_failed")
		result.Boundaries = AppendBoundaries(result.Boundaries, "semantic_verification_deterministic_blocked_fallback")
		return result.Normalize()
	}
	response = response.Normalize()
	if response.RawOutputLoaded || displaySafeRefRejected(response.ResponseRef) {
		return objectiveSemanticVerificationReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(response.AdviceJSON)) > 0 {
		result.DecodeAttempted = true
		decode := BuildObjectiveSemanticVerificationFromJSON(ObjectiveSemanticVerificationJSONDecodeInput{
			RawJSON:          response.AdviceJSON,
			AdviceRef:        firstDisplaySafeRef(input.AdviceRef, response.ResponseRef),
			SourceRef:        firstDisplaySafeRef(input.SourceRef, response.ResponseRef),
			RequiredEvidence: objectiveSemanticRequestRequiredEvidence(request),
			Boundaries: MergeBoundaries(
				input.Boundaries,
				request.Boundaries,
				response.Boundaries,
				[]Boundary{"objective_semantic_verifier_json_response"},
			),
			RawOutputLoaded: response.RawOutputLoaded,
		})
		return objectiveSemanticVerificationReportFromDecode(result, decode)
	}
	if objectiveSemanticVerificationAdviceIsEmpty(response.Advice) {
		return objectiveSemanticVerificationReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_semantic_verification_advice", "provide_objective_semantic_verification_advice", "objective_semantic_verifier_empty_response")
	}
	advice := objectiveSemanticVerificationApplyResponseDefaults(response.Advice, request, response, input).Normalize()
	return objectiveSemanticVerificationReportFromAdvice(result, advice).Normalize()
}

func (r ObjectiveSemanticVerificationReport) Clone() ObjectiveSemanticVerificationReport {
	out := r
	out.Advice = r.Advice.Clone()
	out.JSONDecode = r.JSONDecode.Clone()
	out.Coverage = cloneObjectiveSemanticCoverageAssessments(r.Coverage)
	out.EvidenceRefs = cloneEvidenceRefs(r.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.Findings = cloneStringSlice(r.Findings)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveSemanticVerificationReport) Normalize() ObjectiveSemanticVerificationReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_semantic_verification"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "semantic_verification_advice_only"
	}
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.SuggestedStatus = NormalizeVerificationStatus(string(out.SuggestedStatus))
	if !objectiveSemanticVerificationAdviceIsEmpty(out.Advice) {
		out.Advice = out.Advice.Normalize()
	}
	if objectiveSemanticVerificationJSONDecodeReportPresent(out.JSONDecode) {
		out.JSONDecode = out.JSONDecode.Normalize()
	}
	out.Coverage = normalizeObjectiveSemanticCoverageAssessments(out.Coverage)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Findings = normalizeStringList(out.Findings)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Advice.RawOutputLoaded || out.JSONDecode.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForDeterministicReview = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status == VerificationSatisfied {
		out.Status = VerificationReviewRequired
		out.Boundaries = AppendBoundaries(out.Boundaries, "semantic_verification_cannot_mark_satisfied")
		out.NextHostAction = firstNextHostAction(out.NextHostAction, "run_deterministic_verification_gate")
	}
	return out
}

func cloneObjectiveSemanticCoverageAssessments(values []ObjectiveSemanticCoverageAssessment) []ObjectiveSemanticCoverageAssessment {
	if len(values) == 0 {
		return nil
	}
	out := make([]ObjectiveSemanticCoverageAssessment, 0, len(values))
	for _, value := range values {
		out = append(out, value.Clone())
	}
	return out
}

func normalizeObjectiveSemanticCoverageAssessments(values []ObjectiveSemanticCoverageAssessment) []ObjectiveSemanticCoverageAssessment {
	if len(values) == 0 {
		return nil
	}
	out := make([]ObjectiveSemanticCoverageAssessment, 0, len(values))
	seen := map[DisplaySafeRef]bool{}
	for _, value := range values {
		normalized := value.Normalize()
		if normalized.CriterionRef == "" {
			continue
		}
		if seen[normalized.CriterionRef] {
			continue
		}
		seen[normalized.CriterionRef] = true
		out = append(out, normalized)
	}
	return out
}

func decodeObjectiveSemanticVerificationJSON(raw []byte) (ObjectiveSemanticVerificationAdvice, bool, Boundary) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var advice ObjectiveSemanticVerificationAdvice
	if err := dec.Decode(&advice); err != nil {
		return ObjectiveSemanticVerificationAdvice{}, false, "objective_semantic_verification_json_invalid"
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return ObjectiveSemanticVerificationAdvice{}, false, "objective_semantic_verification_json_trailing_token"
	}
	return advice, true, ""
}

func baseObjectiveSemanticVerificationJSONDecodeReport(input ObjectiveSemanticVerificationJSONDecodeInput) ObjectiveSemanticVerificationJSONDecodeReport {
	return ObjectiveSemanticVerificationJSONDecodeReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_semantic_verification_json_decode",
		RunnerEffect:    "none",
		PromptEffect:    "semantic_verification_advice_only",
		AdviceRef:       normalizeOneDisplaySafeRef(input.AdviceRef),
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_semantic_verification",
				"semantic_verification_advice_only",
				"does_not_mark_objective_satisfied",
				"strict_json_decode",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"no_strategy_dispatch",
			},
			input.Boundaries...,
		),
		NextHostAction:  "review_semantic_verification",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func objectiveSemanticVerificationJSONDecodeReportBlock(report ObjectiveSemanticVerificationJSONDecodeReport, status VerificationStatus, failure FailureClass, missing MissingInput, action NextHostAction, boundary Boundary) ObjectiveSemanticVerificationJSONDecodeReport {
	report.Status = status
	report.FailureClass = failure
	report.MissingInputs = AppendMissingInputs(report.MissingInputs, missing)
	report.Boundaries = AppendBoundaries(report.Boundaries, boundary)
	report.NextHostAction = action
	return report.Normalize()
}

func objectiveSemanticVerificationApplyAdviceDefaults(advice ObjectiveSemanticVerificationAdvice, input ObjectiveSemanticVerificationJSONDecodeInput) ObjectiveSemanticVerificationAdvice {
	advice = advice.Normalize()
	advice.AdviceRef = firstDisplaySafeRef(advice.AdviceRef, input.AdviceRef, "advice:objective_semantic_verification")
	advice.Boundaries = MergeBoundaries(advice.Boundaries, input.Boundaries)
	return advice
}

func objectiveSemanticVerificationApplyResponseDefaults(advice ObjectiveSemanticVerificationAdvice, request ObjectiveSemanticVerifierRequest, response ObjectiveSemanticVerifierResponse, input ObjectiveSemanticVerificationInput) ObjectiveSemanticVerificationAdvice {
	advice = advice.Normalize()
	advice.AdviceRef = firstDisplaySafeRef(advice.AdviceRef, input.AdviceRef, response.ResponseRef, "advice:objective_semantic_verification")
	advice.EvidenceRefs = MergeEvidenceRefs(advice.EvidenceRefs, response.Advice.EvidenceRefs, request.EvidenceRefs, request.Verification.EvidenceRefs)
	advice.Boundaries = MergeBoundaries(advice.Boundaries, input.Boundaries, request.Boundaries, response.Boundaries)
	advice.RawOutputLoaded = advice.RawOutputLoaded || response.RawOutputLoaded || request.RawOutputLoaded
	return advice
}

func baseObjectiveSemanticVerificationReport(input ObjectiveSemanticVerificationInput) ObjectiveSemanticVerificationReport {
	request := input.Request.Normalize()
	return ObjectiveSemanticVerificationReport{
		ContractVersion: ContractVersion,
		Available:       input.Verifier != nil,
		Status:          VerificationBlocked,
		Mode:            "objective_semantic_verification",
		RunnerEffect:    "none",
		PromptEffect:    "semantic_verification_advice_only",
		RequestRef:      request.RequestRef,
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		EvidenceRefs:    MergeEvidenceRefs(request.EvidenceRefs, request.Verification.EvidenceRefs),
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_semantic_verification",
				"semantic_verification_advice_only",
				"does_not_mark_objective_satisfied",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"no_strategy_dispatch",
			},
			input.Boundaries...,
		),
		NextHostAction:  "review_semantic_verification",
		RawOutputLoaded: request.RawOutputLoaded,
	}
}

func objectiveSemanticVerificationReportBlock(report ObjectiveSemanticVerificationReport, status VerificationStatus, failure FailureClass, missing MissingInput, action NextHostAction, boundary Boundary) ObjectiveSemanticVerificationReport {
	report.Status = status
	report.FailureClass = failure
	report.MissingInputs = AppendMissingInputs(report.MissingInputs, missing)
	report.Boundaries = AppendBoundaries(report.Boundaries, boundary)
	report.NextHostAction = action
	return report.Normalize()
}

func objectiveSemanticVerificationReportFromDecode(base ObjectiveSemanticVerificationReport, decode ObjectiveSemanticVerificationJSONDecodeReport) ObjectiveSemanticVerificationReport {
	base.Evaluated = decode.Decoded
	base.DecodeAttempted = true
	base.JSONDecode = decode
	base.Advice = decode.Advice
	base.Coverage = decode.Advice.Coverage
	base.SuggestedStatus = decode.Advice.SuggestedStatus
	base.Status = decode.Status
	base.FailureClass = decode.FailureClass
	base.EvidenceRefs = MergeEvidenceRefs(base.EvidenceRefs, decode.Advice.EvidenceRefs)
	base.MissingInputs = MergeMissingInputs(base.MissingInputs, decode.MissingInputs)
	base.Findings = normalizeStringList(append(cloneStringSlice(base.Findings), decode.Advice.Findings...))
	base.Boundaries = MergeBoundaries(base.Boundaries, decode.Boundaries)
	base.NextHostAction = decode.NextHostAction
	base.RawOutputLoaded = base.RawOutputLoaded || decode.RawOutputLoaded
	base.ReadyForDeterministicReview = objectiveSemanticVerificationReadyForDeterministicReview(decode.Advice)
	return base.Normalize()
}

func objectiveSemanticVerificationReportFromAdvice(base ObjectiveSemanticVerificationReport, advice ObjectiveSemanticVerificationAdvice) ObjectiveSemanticVerificationReport {
	base.Evaluated = true
	base.Advice = advice
	base.Coverage = advice.Coverage
	base.SuggestedStatus = advice.SuggestedStatus
	base.Status = objectiveSemanticVerificationReportStatus(advice)
	base.FailureClass = objectiveSemanticVerificationFailureClass(advice)
	base.EvidenceRefs = MergeEvidenceRefs(base.EvidenceRefs, advice.EvidenceRefs, objectiveSemanticCoverageEvidenceRefs(advice.Coverage))
	base.MissingInputs = MergeMissingInputs(base.MissingInputs, advice.MissingInputs, objectiveSemanticCoverageMissingInputs(advice.Coverage))
	base.Findings = normalizeStringList(append(cloneStringSlice(base.Findings), advice.Findings...))
	base.Boundaries = MergeBoundaries(base.Boundaries, advice.Boundaries, objectiveSemanticCoverageBoundaries(advice.Coverage))
	base.NextHostAction = firstNextHostAction(advice.NextHostAction, objectiveSemanticVerificationNextAction(advice))
	base.ReadyForDeterministicReview = objectiveSemanticVerificationReadyForDeterministicReview(advice)
	base.RawOutputLoaded = base.RawOutputLoaded || advice.RawOutputLoaded
	return base
}

func objectiveSemanticVerificationReportStatus(advice ObjectiveSemanticVerificationAdvice) VerificationStatus {
	advice = advice.Normalize()
	if advice.RawOutputLoaded {
		return VerificationReviewRequired
	}
	hasMissing := false
	hasUncertain := false
	for _, coverage := range advice.Coverage {
		switch coverage.Normalize().Status {
		case ObjectiveSemanticCoverageMissing:
			hasMissing = true
		case ObjectiveSemanticCoverageUncertain, ObjectiveSemanticCoverageUnspecified:
			hasUncertain = true
		}
	}
	if hasMissing {
		return VerificationPartial
	}
	if hasUncertain || len(advice.Coverage) == 0 {
		return VerificationReviewRequired
	}
	return VerificationReviewRequired
}

func objectiveSemanticVerificationFailureClass(advice ObjectiveSemanticVerificationAdvice) FailureClass {
	advice = advice.Normalize()
	if advice.RawOutputLoaded {
		return FailureEvidenceWeak
	}
	for _, coverage := range advice.Coverage {
		switch coverage.Normalize().Status {
		case ObjectiveSemanticCoverageMissing:
			return FailureEvidenceMissing
		case ObjectiveSemanticCoverageUncertain, ObjectiveSemanticCoverageUnspecified:
			return FailureVerificationFailed
		}
	}
	if len(advice.Coverage) == 0 {
		return FailureInvalidInput
	}
	return FailureNone
}

func objectiveSemanticVerificationNextAction(advice ObjectiveSemanticVerificationAdvice) NextHostAction {
	switch objectiveSemanticVerificationFailureClass(advice) {
	case FailureNone:
		return "run_deterministic_verification_gate"
	case FailureEvidenceMissing:
		return "add_evidence_node"
	case FailureInvalidInput:
		return "provide_objective_semantic_verification_advice"
	default:
		return "review_semantic_verification"
	}
}

func objectiveSemanticVerificationReadyForDeterministicReview(advice ObjectiveSemanticVerificationAdvice) bool {
	return objectiveSemanticVerificationFailureClass(advice) == FailureNone && len(advice.Normalize().EvidenceRefs) > 0
}

func objectiveSemanticRequestRequiredEvidence(request ObjectiveSemanticVerifierRequest) []EvidenceRef {
	request = request.Normalize()
	return MergeEvidenceRefs(request.Spec.RequiredEvidence, request.Graph.RequiredEvidence, request.Verification.Frame.RequiredEvidence)
}

func objectiveSemanticCoverageEvidenceRefs(values []ObjectiveSemanticCoverageAssessment) []EvidenceRef {
	var out []EvidenceRef
	for _, value := range values {
		coverage := value.Normalize()
		out = MergeEvidenceRefs(out, coverage.CoveredEvidenceRefs)
	}
	return out
}

func objectiveSemanticCoverageMissingInputs(values []ObjectiveSemanticCoverageAssessment) []MissingInput {
	var out []MissingInput
	for _, value := range values {
		out = MergeMissingInputs(out, value.Normalize().MissingInputs)
	}
	return out
}

func objectiveSemanticCoverageBoundaries(values []ObjectiveSemanticCoverageAssessment) []Boundary {
	var out []Boundary
	for _, value := range values {
		out = MergeBoundaries(out, value.Normalize().Boundaries)
	}
	return out
}

func objectiveSemanticVerificationInputUnsafe(input ObjectiveSemanticVerificationInput) bool {
	if input.Request.RawOutputLoaded {
		return true
	}
	return displaySafeRefRejected(input.AdviceRef) || displaySafeRefRejected(input.SourceRef)
}

func objectiveSemanticVerificationJSONDecodeInputUnsafe(input ObjectiveSemanticVerificationJSONDecodeInput) bool {
	if input.RawOutputLoaded {
		return true
	}
	return displaySafeRefRejected(input.AdviceRef) || displaySafeRefRejected(input.SourceRef)
}

func objectiveSemanticVerificationAdviceIsEmpty(advice ObjectiveSemanticVerificationAdvice) bool {
	advice = advice.Normalize()
	return advice.AdviceRef == "" &&
		advice.SuggestedStatus == VerificationReviewRequired &&
		len(advice.Coverage) == 0 &&
		len(advice.EvidenceRefs) == 0 &&
		len(advice.MissingInputs) == 0 &&
		len(advice.Findings) == 0 &&
		len(advice.Boundaries) == 0
}

func objectiveSemanticVerificationJSONDecodeReportPresent(report ObjectiveSemanticVerificationJSONDecodeReport) bool {
	return report.Decoded || report.AdviceRef != "" || report.Status != ""
}
