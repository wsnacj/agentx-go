package controlcontract

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

type ObjectiveSpecJSONDecodeInput struct {
	RawJSON                   []byte                          `json:"-"`
	SpecRef                   DisplaySafeRef                  `json:"spec_ref,omitempty"`
	RawGoalRef                DisplaySafeRef                  `json:"raw_goal_ref,omitempty"`
	UserGoalDigest            string                          `json:"user_goal_digest,omitempty"`
	ProjectionRef             DisplaySafeRef                  `json:"projection_ref,omitempty"`
	SourceRef                 DisplaySafeRef                  `json:"source_ref,omitempty"`
	AllowedCapabilityRefs     []DisplaySafeRef                `json:"allowed_capability_refs,omitempty"`
	AllowedSideEffectPolicies []ObjectiveSpecSideEffectPolicy `json:"allowed_side_effect_policies,omitempty"`
	Boundaries                []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded           bool                            `json:"raw_output_loaded"`
}

type ObjectiveSpecJSONDecodeReport struct {
	ContractVersion        string                       `json:"contract_version,omitempty"`
	Decoded                bool                         `json:"decoded"`
	Available              bool                         `json:"available"`
	Status                 VerificationStatus           `json:"status,omitempty"`
	Mode                   string                       `json:"mode,omitempty"`
	RunnerEffect           string                       `json:"runner_effect,omitempty"`
	PromptEffect           string                       `json:"prompt_effect,omitempty"`
	SpecRef                DisplaySafeRef               `json:"spec_ref,omitempty"`
	SourceRef              DisplaySafeRef               `json:"source_ref,omitempty"`
	Spec                   ObjectiveSpec                `json:"spec,omitempty"`
	Projection             ObjectiveSpecFrameProjection `json:"projection,omitempty"`
	ReadyForObjectiveFrame bool                         `json:"ready_for_objective_frame"`
	FailureClass           FailureClass                 `json:"failure_class,omitempty"`
	MissingInputs          []MissingInput               `json:"missing_inputs,omitempty"`
	Boundaries             []Boundary                   `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction               `json:"next_host_action,omitempty"`
	RawOutputLoaded        bool                         `json:"raw_output_loaded"`
}

func BuildObjectiveSpecFromJSON(input ObjectiveSpecJSONDecodeInput) ObjectiveSpecJSONDecodeReport {
	result := baseObjectiveSpecJSONDecodeReport(input)
	if objectiveSpecJSONDecodeInputUnsafe(input) {
		return objectiveSpecJSONDecodeReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if input.RawOutputLoaded {
		return objectiveSpecJSONDecodeReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(input.RawJSON)) == 0 {
		return objectiveSpecJSONDecodeReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_spec_json", "provide_objective_spec_json", "objective_spec_json_missing")
	}

	spec, ok, boundary := decodeObjectiveSpecJSON(input.RawJSON)
	if !ok {
		result = objectiveSpecJSONDecodeReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_spec_json", "provide_objective_spec_json", boundary)
		result.Boundaries = AppendBoundaries(result.Boundaries, "deterministic_blocked_fallback", "no_prompt_heuristic_fallback")
		return result.Normalize()
	}
	result.Decoded = true
	spec = objectiveSpecApplyDecodeDefaults(spec, input).Normalize()
	if !objectiveSpecCandidateCapabilitiesAllowed(spec.CandidateCapabilities, input.AllowedCapabilityRefs) {
		result.Spec = spec
		return objectiveSpecJSONDecodeReportBlock(result, VerificationBlocked, FailureCapabilityMissing, "host:allowed_capability_ref", "provide_strategy_scope", "objective_spec_candidate_capability_not_allowed")
	}
	if !objectiveSpecSideEffectPolicyAllowed(spec.SideEffectPolicy, input.AllowedSideEffectPolicies) {
		result.Spec = spec
		return objectiveSpecJSONDecodeReportBlock(result, VerificationBlocked, FailurePolicyBlocked, "host:objective_side_effect_policy", "request_host_approval", "objective_spec_side_effect_policy_not_allowed")
	}

	projection := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec:          spec,
		ProjectionRef: input.ProjectionRef,
		SourceRef:     input.SourceRef,
		Boundaries: MergeBoundaries(
			[]Boundary{"objective_spec_json_validated"},
			input.Boundaries,
		),
	})
	result.Spec = projection.Spec
	result.Projection = projection
	result.Status = projection.Status
	result.FailureClass = projection.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, projection.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, projection.Boundaries)
	result.NextHostAction = projection.NextHostAction
	result.ReadyForObjectiveFrame = projection.Status == VerificationSatisfied && projection.FrameMapped
	if result.ReadyForObjectiveFrame {
		result.FailureClass = FailureNone
	}
	return result.Normalize()
}

func (r ObjectiveSpecJSONDecodeReport) Clone() ObjectiveSpecJSONDecodeReport {
	out := r
	out.Spec = r.Spec.Clone()
	out.Projection = r.Projection.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveSpecJSONDecodeReport) Normalize() ObjectiveSpecJSONDecodeReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_spec_json_decode"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.SpecRef = normalizeOneDisplaySafeRef(out.SpecRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	if !objectiveSpecIsEmpty(out.Spec) {
		out.Spec = out.Spec.Normalize()
	}
	if objectiveSpecFrameProjectionPresent(out.Projection) {
		out.Projection = out.Projection.Normalize()
	}
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Spec.RawOutputLoaded || out.Projection.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForObjectiveFrame = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.ReadyForObjectiveFrame = false
	}
	return out
}

type ObjectiveSpecBuilder interface {
	BuildObjectiveSpec(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error)
}

type ObjectiveSpecBuilderFunc func(context.Context, ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error)

func (f ObjectiveSpecBuilderFunc) BuildObjectiveSpec(ctx context.Context, request ObjectiveSpecBuilderRequest) (ObjectiveSpecBuilderResponse, error) {
	return f(ctx, request)
}

type ObjectiveSpecBuilderRequest struct {
	RequestRef                DisplaySafeRef                  `json:"request_ref,omitempty"`
	RawGoalRef                DisplaySafeRef                  `json:"raw_goal_ref,omitempty"`
	UserGoalDigest            string                          `json:"user_goal_digest,omitempty"`
	CatalogRef                DisplaySafeRef                  `json:"catalog_ref,omitempty"`
	AllowedCapabilityRefs     []DisplaySafeRef                `json:"allowed_capability_refs,omitempty"`
	AllowedSideEffectPolicies []ObjectiveSpecSideEffectPolicy `json:"allowed_side_effect_policies,omitempty"`
	PolicyRefs                []DisplaySafeRef                `json:"policy_refs,omitempty"`
	Boundaries                []Boundary                      `json:"boundaries,omitempty"`
	RawOutputLoaded           bool                            `json:"raw_output_loaded"`
}

func CloneObjectiveSpecBuilderRequest(in ObjectiveSpecBuilderRequest) ObjectiveSpecBuilderRequest {
	out := in
	out.AllowedCapabilityRefs = cloneDisplaySafeRefs(in.AllowedCapabilityRefs)
	out.AllowedSideEffectPolicies = cloneObjectiveSpecSideEffectPolicies(in.AllowedSideEffectPolicies)
	out.PolicyRefs = cloneDisplaySafeRefs(in.PolicyRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveSpecBuilderRequest) Clone() ObjectiveSpecBuilderRequest {
	return CloneObjectiveSpecBuilderRequest(r)
}

func (r ObjectiveSpecBuilderRequest) Normalize() ObjectiveSpecBuilderRequest {
	out := CloneObjectiveSpecBuilderRequest(r)
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.RawGoalRef = normalizeOneDisplaySafeRef(out.RawGoalRef)
	out.UserGoalDigest = normalizeFingerprint(out.UserGoalDigest)
	out.CatalogRef = normalizeOneDisplaySafeRef(out.CatalogRef)
	out.AllowedCapabilityRefs = normalizeDisplaySafeRefs(out.AllowedCapabilityRefs)
	out.AllowedSideEffectPolicies = normalizeObjectiveSpecSideEffectPolicies(out.AllowedSideEffectPolicies)
	out.PolicyRefs = normalizeDisplaySafeRefs(out.PolicyRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSpecBuilderResponse struct {
	ResponseRef     DisplaySafeRef `json:"response_ref,omitempty"`
	Spec            ObjectiveSpec  `json:"spec,omitempty"`
	SpecJSON        []byte         `json:"-"`
	Boundaries      []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded bool           `json:"raw_output_loaded"`
}

func CloneObjectiveSpecBuilderResponse(in ObjectiveSpecBuilderResponse) ObjectiveSpecBuilderResponse {
	out := in
	out.Spec = in.Spec.Clone()
	out.SpecJSON = append([]byte(nil), in.SpecJSON...)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (r ObjectiveSpecBuilderResponse) Clone() ObjectiveSpecBuilderResponse {
	return CloneObjectiveSpecBuilderResponse(r)
}

func (r ObjectiveSpecBuilderResponse) Normalize() ObjectiveSpecBuilderResponse {
	out := CloneObjectiveSpecBuilderResponse(r)
	out.ResponseRef = normalizeOneDisplaySafeRef(out.ResponseRef)
	if !objectiveSpecIsEmpty(out.Spec) {
		out.Spec = out.Spec.Normalize()
	}
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	return out
}

type ObjectiveSpecBuildInput struct {
	Enabled       bool                        `json:"enabled"`
	Builder       ObjectiveSpecBuilder        `json:"-"`
	Request       ObjectiveSpecBuilderRequest `json:"request,omitempty"`
	ProjectionRef DisplaySafeRef              `json:"projection_ref,omitempty"`
	SourceRef     DisplaySafeRef              `json:"source_ref,omitempty"`
	Boundaries    []Boundary                  `json:"boundaries,omitempty"`
}

type ObjectiveSpecBuildReport struct {
	ContractVersion        string                        `json:"contract_version,omitempty"`
	Built                  bool                          `json:"built"`
	Available              bool                          `json:"available"`
	Status                 VerificationStatus            `json:"status,omitempty"`
	Mode                   string                        `json:"mode,omitempty"`
	RunnerEffect           string                        `json:"runner_effect,omitempty"`
	PromptEffect           string                        `json:"prompt_effect,omitempty"`
	RequestRef             DisplaySafeRef                `json:"request_ref,omitempty"`
	SourceRef              DisplaySafeRef                `json:"source_ref,omitempty"`
	BuilderCalled          bool                          `json:"builder_called"`
	DecodeAttempted        bool                          `json:"decode_attempted"`
	Spec                   ObjectiveSpec                 `json:"spec,omitempty"`
	JSONDecode             ObjectiveSpecJSONDecodeReport `json:"json_decode,omitempty"`
	Projection             ObjectiveSpecFrameProjection  `json:"projection,omitempty"`
	ReadyForObjectiveFrame bool                          `json:"ready_for_objective_frame"`
	FailureClass           FailureClass                  `json:"failure_class,omitempty"`
	MissingInputs          []MissingInput                `json:"missing_inputs,omitempty"`
	Boundaries             []Boundary                    `json:"boundaries,omitempty"`
	NextHostAction         NextHostAction                `json:"next_host_action,omitempty"`
	RawOutputLoaded        bool                          `json:"raw_output_loaded"`
}

func BuildObjectiveSpecWithBuilder(ctx context.Context, input ObjectiveSpecBuildInput) ObjectiveSpecBuildReport {
	result := baseObjectiveSpecBuildReport(input)
	if objectiveSpecBuildInputUnsafe(input) {
		return objectiveSpecBuildReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !input.Enabled {
		return objectiveSpecBuildReportBlock(result, VerificationBlocked, FailureInsufficientInformation, "host:objective_spec_builder_enabled", "enable_objective_closure", "objective_spec_builder_disabled")
	}
	if input.Builder == nil {
		return objectiveSpecBuildReportBlock(result, VerificationBlocked, FailureHostAdapterMissing, "host:objective_spec_builder", "provide_objective_spec_builder", "objective_spec_builder_missing")
	}

	request := input.Request.Normalize()
	result.BuilderCalled = true
	response, err := input.Builder.BuildObjectiveSpec(ctx, request)
	if err != nil {
		result = objectiveSpecBuildReportBlock(result, VerificationBlocked, FailureExternalDependencyUnavailable, "host:objective_spec_builder_result", "provide_objective_spec", "objective_spec_builder_failed")
		result.Boundaries = AppendBoundaries(result.Boundaries, "deterministic_blocked_fallback", "no_prompt_heuristic_fallback")
		return result.Normalize()
	}
	response = response.Normalize()
	if response.RawOutputLoaded || displaySafeRefRejected(response.ResponseRef) {
		return objectiveSpecBuildReportBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(response.SpecJSON)) > 0 {
		result.DecodeAttempted = true
		decode := BuildObjectiveSpecFromJSON(ObjectiveSpecJSONDecodeInput{
			RawJSON:                   response.SpecJSON,
			SpecRef:                   response.ResponseRef,
			RawGoalRef:                request.RawGoalRef,
			UserGoalDigest:            request.UserGoalDigest,
			ProjectionRef:             input.ProjectionRef,
			SourceRef:                 firstDisplaySafeRef(input.SourceRef, response.ResponseRef),
			AllowedCapabilityRefs:     request.AllowedCapabilityRefs,
			AllowedSideEffectPolicies: request.AllowedSideEffectPolicies,
			Boundaries: MergeBoundaries(
				input.Boundaries,
				request.Boundaries,
				response.Boundaries,
				[]Boundary{"objective_spec_builder_json_response"},
			),
			RawOutputLoaded: response.RawOutputLoaded,
		})
		return objectiveSpecBuildReportFromDecode(result, decode)
	}
	if objectiveSpecIsEmpty(response.Spec) {
		return objectiveSpecBuildReportBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_spec", "provide_objective_spec", "objective_spec_builder_empty_response")
	}

	spec := objectiveSpecApplyBuilderDefaults(response.Spec, request, response).Normalize()
	if !objectiveSpecCandidateCapabilitiesAllowed(spec.CandidateCapabilities, request.AllowedCapabilityRefs) {
		result.Spec = spec
		return objectiveSpecBuildReportBlock(result, VerificationBlocked, FailureCapabilityMissing, "host:allowed_capability_ref", "provide_strategy_scope", "objective_spec_candidate_capability_not_allowed")
	}
	if !objectiveSpecSideEffectPolicyAllowed(spec.SideEffectPolicy, request.AllowedSideEffectPolicies) {
		result.Spec = spec
		return objectiveSpecBuildReportBlock(result, VerificationBlocked, FailurePolicyBlocked, "host:objective_side_effect_policy", "request_host_approval", "objective_spec_side_effect_policy_not_allowed")
	}
	projection := BuildObjectiveSpecFrameProjection(ObjectiveSpecFrameProjectionInput{
		Spec:          spec,
		ProjectionRef: input.ProjectionRef,
		SourceRef:     firstDisplaySafeRef(input.SourceRef, response.ResponseRef),
		Boundaries: MergeBoundaries(
			input.Boundaries,
			request.Boundaries,
			response.Boundaries,
			[]Boundary{"objective_spec_builder_direct_response"},
		),
	})
	result.Spec = projection.Spec
	result.Projection = projection
	result.Status = projection.Status
	result.FailureClass = projection.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, projection.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, projection.Boundaries)
	result.NextHostAction = projection.NextHostAction
	result.ReadyForObjectiveFrame = projection.Status == VerificationSatisfied && projection.FrameMapped
	result.Built = result.ReadyForObjectiveFrame
	if result.ReadyForObjectiveFrame {
		result.FailureClass = FailureNone
	}
	return result.Normalize()
}

func (r ObjectiveSpecBuildReport) Clone() ObjectiveSpecBuildReport {
	out := r
	out.Spec = r.Spec.Clone()
	out.JSONDecode = r.JSONDecode.Clone()
	out.Projection = r.Projection.Clone()
	out.MissingInputs = cloneMissingInputs(r.MissingInputs)
	out.Boundaries = cloneBoundaries(r.Boundaries)
	return out
}

func (r ObjectiveSpecBuildReport) Normalize() ObjectiveSpecBuildReport {
	out := r.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_spec_builder"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.RequestRef = normalizeOneDisplaySafeRef(out.RequestRef)
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	if !objectiveSpecIsEmpty(out.Spec) {
		out.Spec = out.Spec.Normalize()
	}
	if objectiveSpecJSONDecodeReportPresent(out.JSONDecode) {
		out.JSONDecode = out.JSONDecode.Normalize()
	}
	if objectiveSpecFrameProjectionPresent(out.Projection) {
		out.Projection = out.Projection.Normalize()
	}
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Spec.RawOutputLoaded || out.JSONDecode.RawOutputLoaded || out.Projection.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.Built = false
		out.ReadyForObjectiveFrame = false
		out.FailureClass = firstFailureClass(out.FailureClass, FailureEvidenceWeak)
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	if out.Status != VerificationSatisfied {
		out.Built = false
		out.ReadyForObjectiveFrame = false
	}
	return out
}

func baseObjectiveSpecJSONDecodeReport(input ObjectiveSpecJSONDecodeInput) ObjectiveSpecJSONDecodeReport {
	return ObjectiveSpecJSONDecodeReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_spec_json_decode",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		SpecRef:         normalizeOneDisplaySafeRef(input.SpecRef),
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		FailureClass:    FailureInsufficientInformation,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_spec_json_decoder",
				"strict_json_decoder",
				"disallow_unknown_fields",
				"no_code_fence_stripping",
				"no_llm_call",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			input.Boundaries,
		),
		NextHostAction:  "provide_objective_spec_json",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func baseObjectiveSpecBuildReport(input ObjectiveSpecBuildInput) ObjectiveSpecBuildReport {
	request := input.Request.Normalize()
	return ObjectiveSpecBuildReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_spec_builder",
		RunnerEffect:    "none",
		PromptEffect:    "host_builder_interface_only",
		RequestRef:      request.RequestRef,
		SourceRef:       normalizeOneDisplaySafeRef(input.SourceRef),
		FailureClass:    FailureInsufficientInformation,
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_spec_builder_interface",
				"host_owned_llm_builder",
				"core_validator_only",
				"no_concrete_model_binding",
				"no_prompt_heuristic_fallback",
				"no_runner_dispatch",
				"no_backend_execution",
			},
			input.Boundaries,
			request.Boundaries,
		),
		NextHostAction:  "provide_objective_spec",
		RawOutputLoaded: request.RawOutputLoaded,
	}
}

func objectiveSpecJSONDecodeReportBlock(result ObjectiveSpecJSONDecodeReport, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveSpecJSONDecodeReport {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func objectiveSpecBuildReportBlock(result ObjectiveSpecBuildReport, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveSpecBuildReport {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func decodeObjectiveSpecJSON(raw []byte) (ObjectiveSpec, bool, Boundary) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var spec ObjectiveSpec
	if err := decoder.Decode(&spec); err != nil {
		return ObjectiveSpec{}, false, "objective_spec_json_decode_failed"
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ObjectiveSpec{}, false, "objective_spec_json_trailing_tokens"
	}
	return spec, true, ""
}

func objectiveSpecApplyDecodeDefaults(spec ObjectiveSpec, input ObjectiveSpecJSONDecodeInput) ObjectiveSpec {
	out := spec.Clone()
	if out.SpecRef == "" {
		out.SpecRef = input.SpecRef
	}
	if out.RawGoalRef == "" {
		out.RawGoalRef = input.RawGoalRef
	}
	if out.UserGoalDigest == "" {
		out.UserGoalDigest = input.UserGoalDigest
	}
	out.Boundaries = MergeBoundaries(out.Boundaries, input.Boundaries)
	return out
}

func objectiveSpecApplyBuilderDefaults(spec ObjectiveSpec, request ObjectiveSpecBuilderRequest, response ObjectiveSpecBuilderResponse) ObjectiveSpec {
	out := spec.Clone()
	if out.SpecRef == "" {
		out.SpecRef = response.ResponseRef
	}
	if out.RawGoalRef == "" {
		out.RawGoalRef = request.RawGoalRef
	}
	if out.UserGoalDigest == "" {
		out.UserGoalDigest = request.UserGoalDigest
	}
	out.PolicyRefs = normalizeDisplaySafeRefs(append(out.PolicyRefs, request.PolicyRefs...))
	out.Boundaries = MergeBoundaries(out.Boundaries, request.Boundaries, response.Boundaries)
	return out
}

func objectiveSpecBuildReportFromDecode(result ObjectiveSpecBuildReport, decode ObjectiveSpecJSONDecodeReport) ObjectiveSpecBuildReport {
	result.JSONDecode = decode
	result.Spec = decode.Spec
	result.Projection = decode.Projection
	result.Status = decode.Status
	result.FailureClass = decode.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, decode.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, decode.Boundaries)
	result.NextHostAction = decode.NextHostAction
	result.ReadyForObjectiveFrame = decode.ReadyForObjectiveFrame
	result.Built = decode.ReadyForObjectiveFrame
	result.RawOutputLoaded = result.RawOutputLoaded || decode.RawOutputLoaded
	if result.ReadyForObjectiveFrame {
		result.FailureClass = FailureNone
	}
	return result.Normalize()
}

func objectiveSpecJSONDecodeInputUnsafe(input ObjectiveSpecJSONDecodeInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.SpecRef) ||
		displaySafeRefRejected(input.RawGoalRef) ||
		displaySafeRefRejected(input.ProjectionRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefSliceRejected(input.AllowedCapabilityRefs) ||
		ContainsUnsafeRawOutput(input.UserGoalDigest)
}

func objectiveSpecBuildInputUnsafe(input ObjectiveSpecBuildInput) bool {
	request := input.Request
	return request.RawOutputLoaded ||
		displaySafeRefRejected(input.ProjectionRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefRejected(request.RequestRef) ||
		displaySafeRefRejected(request.RawGoalRef) ||
		displaySafeRefRejected(request.CatalogRef) ||
		displaySafeRefSliceRejected(request.AllowedCapabilityRefs) ||
		displaySafeRefSliceRejected(request.PolicyRefs) ||
		ContainsUnsafeRawOutput(request.UserGoalDigest)
}

func objectiveSpecCandidateCapabilitiesAllowed(candidates, allowed []DisplaySafeRef) bool {
	normalizedCandidates := normalizeDisplaySafeRefs(candidates)
	normalizedAllowed := normalizeDisplaySafeRefs(allowed)
	if len(normalizedCandidates) == 0 || len(normalizedAllowed) == 0 {
		return true
	}
	allowedSet := make(map[DisplaySafeRef]struct{}, len(normalizedAllowed))
	for _, ref := range normalizedAllowed {
		allowedSet[ref] = struct{}{}
	}
	for _, ref := range normalizedCandidates {
		if _, ok := allowedSet[ref]; !ok {
			return false
		}
	}
	return true
}

func objectiveSpecSideEffectPolicyAllowed(policy ObjectiveSpecSideEffectPolicy, allowed []ObjectiveSpecSideEffectPolicy) bool {
	normalizedPolicy := NormalizeObjectiveSpecSideEffectPolicy(string(policy))
	if normalizedPolicy == ObjectiveSpecSideEffectUnspecified {
		return true
	}
	normalizedAllowed := normalizeObjectiveSpecSideEffectPolicies(allowed)
	if len(normalizedAllowed) == 0 {
		normalizedAllowed = []ObjectiveSpecSideEffectPolicy{
			ObjectiveSpecSideEffectReadOnly,
			ObjectiveSpecSideEffectRequiresApproval,
			ObjectiveSpecSideEffectForbidden,
		}
	}
	for _, value := range normalizedAllowed {
		if value == normalizedPolicy {
			return true
		}
	}
	return false
}

func normalizeObjectiveSpecSideEffectPolicies(in []ObjectiveSpecSideEffectPolicy) []ObjectiveSpecSideEffectPolicy {
	out := make([]ObjectiveSpecSideEffectPolicy, 0, len(in))
	seen := map[ObjectiveSpecSideEffectPolicy]struct{}{}
	for _, value := range in {
		normalized := NormalizeObjectiveSpecSideEffectPolicy(string(value))
		if normalized == ObjectiveSpecSideEffectUnspecified {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}
	return out
}

func cloneObjectiveSpecSideEffectPolicies(in []ObjectiveSpecSideEffectPolicy) []ObjectiveSpecSideEffectPolicy {
	if len(in) == 0 {
		return nil
	}
	return append([]ObjectiveSpecSideEffectPolicy(nil), in...)
}

func objectiveSpecIsEmpty(spec ObjectiveSpec) bool {
	return spec.SpecRef == "" &&
		spec.ObjectiveID == "" &&
		spec.UserGoalDigest == "" &&
		spec.RawGoalRef == "" &&
		spec.GoalSummary == "" &&
		spec.ControlMode == "" &&
		spec.Intensity == "" &&
		len(spec.SuccessCriteria) == 0 &&
		len(spec.Constraints) == 0 &&
		len(spec.RequiredEvidence) == 0 &&
		len(spec.CandidateCapabilities) == 0 &&
		len(spec.SourceContext) == 0 &&
		spec.SideEffectPolicy == "" &&
		spec.MissingInfoPolicy == "" &&
		len(spec.AcceptablePartial) == 0 &&
		spec.Budget.BudgetRef == "" &&
		spec.Budget.MaxNodes == 0 &&
		spec.Budget.MaxAttempts == 0 &&
		spec.Budget.MaxDurationSeconds == 0 &&
		spec.Budget.MaxCostUnits == 0 &&
		len(spec.Budget.PolicyRefs) == 0 &&
		len(spec.ApprovalRefs) == 0 &&
		len(spec.PolicyRefs) == 0 &&
		len(spec.Boundaries) == 0 &&
		len(spec.MissingInputs) == 0 &&
		!spec.RawOutputLoaded
}

func objectiveSpecJSONDecodeReportPresent(report ObjectiveSpecJSONDecodeReport) bool {
	return report.Decoded ||
		report.Available ||
		report.Status != "" ||
		report.Mode != "" ||
		report.SpecRef != "" ||
		report.SourceRef != "" ||
		!objectiveSpecIsEmpty(report.Spec) ||
		objectiveSpecFrameProjectionPresent(report.Projection) ||
		len(report.MissingInputs) > 0 ||
		len(report.Boundaries) > 0 ||
		report.NextHostAction != "" ||
		report.RawOutputLoaded
}

func objectiveSpecFrameProjectionPresent(projection ObjectiveSpecFrameProjection) bool {
	return projection.Projected ||
		projection.Available ||
		projection.Status != "" ||
		projection.Mode != "" ||
		projection.ProjectionRef != "" ||
		projection.SourceRef != "" ||
		!objectiveSpecIsEmpty(projection.Spec) ||
		projection.Frame.ID != "" ||
		len(projection.MappingWarnings) > 0 ||
		len(projection.MissingInputs) > 0 ||
		len(projection.Boundaries) > 0 ||
		projection.NextHostAction != "" ||
		projection.RawOutputLoaded
}
