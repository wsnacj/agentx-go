package controlcontract

import (
	"bytes"
	"encoding/json"
	"strings"
)

type ObjectiveRecoveryTarget struct {
	ContractVersion         string           `json:"contract_version,omitempty"`
	TargetRef               DisplaySafeRef   `json:"target_ref,omitempty"`
	SubjectRef              DisplaySafeRef   `json:"subject_ref,omitempty"`
	MissingInput            MissingInput     `json:"missing_input,omitempty"`
	MissingEvidence         EvidenceRef      `json:"missing_evidence,omitempty"`
	FailureClass            FailureClass     `json:"failure_class,omitempty"`
	FailureReason           string           `json:"failure_reason,omitempty"`
	SuggestedStrategyRefs   []DisplaySafeRef `json:"suggested_strategy_refs,omitempty"`
	SuggestedCapabilityRefs []DisplaySafeRef `json:"suggested_capability_refs,omitempty"`
	SuggestedToolRefs       []DisplaySafeRef `json:"suggested_tool_refs,omitempty"`
	Boundaries              []Boundary       `json:"boundaries,omitempty"`
	RawOutputLoaded         bool             `json:"raw_output_loaded"`
}

func CloneObjectiveRecoveryTarget(in ObjectiveRecoveryTarget) ObjectiveRecoveryTarget {
	out := in
	out.MissingEvidence = in.MissingEvidence.Normalize()
	out.SuggestedStrategyRefs = cloneDisplaySafeRefs(in.SuggestedStrategyRefs)
	out.SuggestedCapabilityRefs = cloneDisplaySafeRefs(in.SuggestedCapabilityRefs)
	out.SuggestedToolRefs = cloneDisplaySafeRefs(in.SuggestedToolRefs)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (t ObjectiveRecoveryTarget) Clone() ObjectiveRecoveryTarget {
	return CloneObjectiveRecoveryTarget(t)
}

func (t ObjectiveRecoveryTarget) Normalize() ObjectiveRecoveryTarget {
	out := t.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.TargetRef = normalizeOneDisplaySafeRef(out.TargetRef)
	out.SubjectRef = normalizeOneDisplaySafeRef(out.SubjectRef)
	out.MissingInput = firstMissingInput(out.MissingInput, "")
	out.MissingEvidence = out.MissingEvidence.Normalize()
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = managedObjectiveReplannerSafeReason(out.FailureReason)
	out.SuggestedStrategyRefs = normalizeDisplaySafeRefs(out.SuggestedStrategyRefs)
	out.SuggestedCapabilityRefs = normalizeDisplaySafeRefs(out.SuggestedCapabilityRefs)
	out.SuggestedToolRefs = normalizeDisplaySafeRefs(out.SuggestedToolRefs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	if out.RawOutputLoaded {
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
	}
	return out
}

type ObjectiveRecoveryContractInput struct {
	ContractRef             DisplaySafeRef            `json:"contract_ref,omitempty"`
	SourceRef               DisplaySafeRef            `json:"source_ref,omitempty"`
	Producer                DisplaySafeRef            `json:"producer,omitempty"`
	ObjectiveID             string                    `json:"objective_id,omitempty"`
	CurrentStrategyRef      DisplaySafeRef            `json:"current_strategy_ref,omitempty"`
	RecoveryRecommended     bool                      `json:"recovery_recommended"`
	FinalAnswerRecommended  bool                      `json:"final_answer_recommended"`
	FailureClass            FailureClass              `json:"failure_class,omitempty"`
	FailureReason           string                    `json:"failure_reason,omitempty"`
	Targets                 []ObjectiveRecoveryTarget `json:"targets,omitempty"`
	SuggestedStrategyRefs   []DisplaySafeRef          `json:"suggested_strategy_refs,omitempty"`
	SuggestedCapabilityRefs []DisplaySafeRef          `json:"suggested_capability_refs,omitempty"`
	SuggestedToolRefs       []DisplaySafeRef          `json:"suggested_tool_refs,omitempty"`
	EvidenceRefs            []EvidenceRef             `json:"evidence_refs,omitempty"`
	MissingInputs           []MissingInput            `json:"missing_inputs,omitempty"`
	Boundaries              []Boundary                `json:"boundaries,omitempty"`
	RawOutputLoaded         bool                      `json:"raw_output_loaded"`
}

type ObjectiveRecoveryContract struct {
	ContractVersion         string                    `json:"contract_version,omitempty"`
	Projected               bool                      `json:"projected"`
	Available               bool                      `json:"available"`
	Recommended             bool                      `json:"recommended"`
	FinalAnswerRecommended  bool                      `json:"final_answer_recommended"`
	Status                  VerificationStatus        `json:"status,omitempty"`
	Mode                    string                    `json:"mode,omitempty"`
	RunnerEffect            string                    `json:"runner_effect,omitempty"`
	PromptEffect            string                    `json:"prompt_effect,omitempty"`
	ContractRef             DisplaySafeRef            `json:"contract_ref,omitempty"`
	SourceRef               DisplaySafeRef            `json:"source_ref,omitempty"`
	Producer                DisplaySafeRef            `json:"producer,omitempty"`
	ObjectiveID             string                    `json:"objective_id,omitempty"`
	CurrentStrategyRef      DisplaySafeRef            `json:"current_strategy_ref,omitempty"`
	TargetCount             int                       `json:"target_count,omitempty"`
	Targets                 []ObjectiveRecoveryTarget `json:"targets,omitempty"`
	SuggestedStrategyRefs   []DisplaySafeRef          `json:"suggested_strategy_refs,omitempty"`
	SuggestedCapabilityRefs []DisplaySafeRef          `json:"suggested_capability_refs,omitempty"`
	SuggestedToolRefs       []DisplaySafeRef          `json:"suggested_tool_refs,omitempty"`
	ReplannerSource         ReplannerSourceProjection `json:"replanner_source,omitempty"`
	ReplanProposal          ObjectiveReplanProposal   `json:"replan_proposal,omitempty"`
	EvidenceRefs            []EvidenceRef             `json:"evidence_refs,omitempty"`
	MissingInputs           []MissingInput            `json:"missing_inputs,omitempty"`
	FailureClass            FailureClass              `json:"failure_class,omitempty"`
	FailureReason           string                    `json:"failure_reason,omitempty"`
	Boundaries              []Boundary                `json:"boundaries,omitempty"`
	NextHostAction          NextHostAction            `json:"next_host_action,omitempty"`
	RawOutputLoaded         bool                      `json:"raw_output_loaded"`
}

func BuildObjectiveRecoveryContract(input ObjectiveRecoveryContractInput) ObjectiveRecoveryContract {
	result := baseObjectiveRecoveryContract(input)
	if objectiveRecoveryContractInputUnsafe(input) {
		return objectiveRecoveryContractBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	result.Targets = objectiveRecoveryTargetsWithFallbacks(input.Targets, input)
	result.TargetCount = len(result.Targets)
	result.SuggestedStrategyRefs = normalizeDisplaySafeRefs(input.SuggestedStrategyRefs)
	result.SuggestedCapabilityRefs = normalizeDisplaySafeRefs(input.SuggestedCapabilityRefs)
	result.SuggestedToolRefs = normalizeDisplaySafeRefs(input.SuggestedToolRefs)
	for _, target := range result.Targets {
		result.SuggestedStrategyRefs = mergeRecoveryDisplayRefs(result.SuggestedStrategyRefs, target.SuggestedStrategyRefs)
		result.SuggestedCapabilityRefs = mergeRecoveryDisplayRefs(result.SuggestedCapabilityRefs, target.SuggestedCapabilityRefs)
		result.SuggestedToolRefs = mergeRecoveryDisplayRefs(result.SuggestedToolRefs, target.SuggestedToolRefs)
		result.EvidenceRefs = MergeEvidenceRefs(result.EvidenceRefs, []EvidenceRef{target.MissingEvidence})
		if target.MissingInput != "" {
			result.MissingInputs = AppendMissingInputs(result.MissingInputs, target.MissingInput)
		}
	}
	if !result.Recommended || result.FinalAnswerRecommended {
		result.Status = VerificationNotApplicable
		result.FailureClass = FailureNone
		result.NextHostAction = "return_current_result"
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_recovery_not_recommended")
		return result.Normalize()
	}
	if result.TargetCount == 0 {
		return objectiveRecoveryContractBlock(result, VerificationBlocked, FailureEvidenceMissing, "host:recovery_target", "provide_objective_recovery_contract", "objective_recovery_target_missing")
	}
	if len(result.SuggestedStrategyRefs) == 0 && len(result.SuggestedCapabilityRefs) == 0 && len(result.SuggestedToolRefs) == 0 {
		return objectiveRecoveryContractBlock(result, VerificationBlocked, FailureCapabilityMissing, "host:recovery_capability_ref", "provide_objective_recovery_contract", "objective_recovery_capability_missing")
	}
	result.Status = VerificationPartial
	result.FailureClass = firstFailureClass(result.FailureClass, FailureEvidenceMissing)
	result.NextHostAction = "host_may_add_evidence_node"
	result.Boundaries = AppendBoundaries(result.Boundaries, "objective_recovery_ready_for_replan")
	result.ReplannerSource = objectiveRecoveryReplannerSource(result)
	result.ReplanProposal = objectiveRecoveryReplanProposal(result)
	return result.Normalize()
}

func (c ObjectiveRecoveryContract) Normalize() ObjectiveRecoveryContract {
	out := c
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Available = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_recovery_contract"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.ContractRef = firstDisplaySafeRef(out.ContractRef, "contract:objective_recovery")
	out.SourceRef = normalizeOneDisplaySafeRef(out.SourceRef)
	out.Producer = firstDisplaySafeRef(out.Producer, "host:objective_recovery_contract")
	out.ObjectiveID = objectiveGraphSafeID(out.ObjectiveID)
	out.CurrentStrategyRef = normalizeOneDisplaySafeRef(out.CurrentStrategyRef)
	out.Targets = normalizeObjectiveRecoveryTargets(out.Targets)
	out.TargetCount = len(out.Targets)
	out.SuggestedStrategyRefs = normalizeDisplaySafeRefs(out.SuggestedStrategyRefs)
	out.SuggestedCapabilityRefs = normalizeDisplaySafeRefs(out.SuggestedCapabilityRefs)
	out.SuggestedToolRefs = normalizeDisplaySafeRefs(out.SuggestedToolRefs)
	out.ReplannerSource = out.ReplannerSource.Normalize()
	out.ReplanProposal = out.ReplanProposal.Normalize()
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.FailureReason = managedObjectiveReplannerSafeReason(out.FailureReason)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

type ObjectiveRecoveryContractJSONDecodeInput struct {
	RawJSON            []byte         `json:"-"`
	ContractRef        DisplaySafeRef `json:"contract_ref,omitempty"`
	SourceRef          DisplaySafeRef `json:"source_ref,omitempty"`
	Producer           DisplaySafeRef `json:"producer,omitempty"`
	ObjectiveID        string         `json:"objective_id,omitempty"`
	CurrentStrategyRef DisplaySafeRef `json:"current_strategy_ref,omitempty"`
	Boundaries         []Boundary     `json:"boundaries,omitempty"`
	RawOutputLoaded    bool           `json:"raw_output_loaded"`
}

type ObjectiveRecoveryContractJSONDecodeReport struct {
	ContractVersion string                    `json:"contract_version,omitempty"`
	Decoded         bool                      `json:"decoded"`
	Available       bool                      `json:"available"`
	Status          VerificationStatus        `json:"status,omitempty"`
	Mode            string                    `json:"mode,omitempty"`
	RunnerEffect    string                    `json:"runner_effect,omitempty"`
	PromptEffect    string                    `json:"prompt_effect,omitempty"`
	Contract        ObjectiveRecoveryContract `json:"contract,omitempty"`
	FailureClass    FailureClass              `json:"failure_class,omitempty"`
	MissingInputs   []MissingInput            `json:"missing_inputs,omitempty"`
	Boundaries      []Boundary                `json:"boundaries,omitempty"`
	NextHostAction  NextHostAction            `json:"next_host_action,omitempty"`
	RawOutputLoaded bool                      `json:"raw_output_loaded"`
}

func BuildObjectiveRecoveryContractFromJSON(input ObjectiveRecoveryContractJSONDecodeInput) ObjectiveRecoveryContractJSONDecodeReport {
	result := baseObjectiveRecoveryContractJSONDecodeReport(input)
	if objectiveRecoveryContractJSONDecodeInputUnsafe(input) {
		return objectiveRecoveryContractJSONDecodeBlock(result, VerificationReviewRequired, FailureEvidenceWeak, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if len(bytes.TrimSpace(input.RawJSON)) == 0 {
		return objectiveRecoveryContractJSONDecodeBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_recovery_contract_json", "provide_objective_recovery_contract", "objective_recovery_contract_json_missing")
	}
	decoded, ok := decodeObjectiveRecoveryJSON(input.RawJSON)
	if !ok {
		return objectiveRecoveryContractJSONDecodeBlock(result, VerificationBlocked, FailureInvalidInput, "host:objective_recovery_contract_json", "provide_objective_recovery_contract", "objective_recovery_contract_json_invalid")
	}
	contract := BuildObjectiveRecoveryContract(objectiveRecoveryContractInputFromJSON(decoded, input))
	result.Decoded = true
	result.Contract = contract
	result.Status = contract.Status
	result.FailureClass = contract.FailureClass
	result.MissingInputs = MergeMissingInputs(result.MissingInputs, contract.MissingInputs)
	result.Boundaries = MergeBoundaries(result.Boundaries, contract.Boundaries)
	result.NextHostAction = contract.NextHostAction
	result.RawOutputLoaded = input.RawOutputLoaded || contract.RawOutputLoaded
	return result.Normalize()
}

func (r ObjectiveRecoveryContractJSONDecodeReport) Normalize() ObjectiveRecoveryContractJSONDecodeReport {
	out := r
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Available = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Mode = normalizeControlToken(out.Mode)
	if out.Mode == "" {
		out.Mode = "objective_recovery_contract_json_decode"
	}
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	out.Contract = out.Contract.Normalize()
	out.FailureClass = NormalizeFailureClass(string(out.FailureClass))
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	if out.RawOutputLoaded || out.Contract.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		if out.FailureClass == FailureNone {
			out.FailureClass = FailureEvidenceWeak
		}
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

type objectiveRecoveryJSONContract struct {
	FinalAnswerRecommended  objectiveRecoveryJSONBool     `json:"final_answer_recommended"`
	RecoveryRecommended     objectiveRecoveryJSONBool     `json:"recovery_recommended"`
	Reason                  string                        `json:"reason,omitempty"`
	RecoveryReason          string                        `json:"recovery_reason,omitempty"`
	FailureClass            string                        `json:"failure_class,omitempty"`
	SuggestedRecoveryTools  []string                      `json:"suggested_recovery_tools,omitempty"`
	SuggestedToolRefs       []string                      `json:"suggested_tool_refs,omitempty"`
	SuggestedCapabilityRefs []string                      `json:"suggested_capability_refs,omitempty"`
	SuggestedStrategyRefs   []string                      `json:"suggested_strategy_refs,omitempty"`
	RecoveryTargets         []objectiveRecoveryJSONTarget `json:"recovery_targets,omitempty"`
	Targets                 []objectiveRecoveryJSONTarget `json:"targets,omitempty"`
}

type objectiveRecoveryJSONTarget struct {
	TargetRef               string   `json:"target_ref,omitempty"`
	SubjectRef              string   `json:"subject_ref,omitempty"`
	MissingInput            string   `json:"missing_input,omitempty"`
	MissingDimension        string   `json:"missing_dimension,omitempty"`
	MissingEvidenceRef      string   `json:"missing_evidence_ref,omitempty"`
	MissingEvidenceKind     string   `json:"missing_evidence_kind,omitempty"`
	FailureClass            string   `json:"failure_class,omitempty"`
	FailureCode             string   `json:"failure_code,omitempty"`
	SuggestedTools          []string `json:"suggested_tools,omitempty"`
	SuggestedToolRefs       []string `json:"suggested_tool_refs,omitempty"`
	SuggestedCapabilityRefs []string `json:"suggested_capability_refs,omitempty"`
	SuggestedStrategyRefs   []string `json:"suggested_strategy_refs,omitempty"`
}

type objectiveRecoveryJSONBool struct {
	Set   bool
	Value bool
}

func (b *objectiveRecoveryJSONBool) UnmarshalJSON(raw []byte) error {
	b.Set = true
	var boolValue bool
	if err := json.Unmarshal(raw, &boolValue); err == nil {
		b.Value = boolValue
		return nil
	}
	var stringValue string
	if err := json.Unmarshal(raw, &stringValue); err == nil {
		switch normalizeEnumToken(stringValue) {
		case "1", "true", "t", "yes", "y", "on":
			b.Value = true
		default:
			b.Value = false
		}
		return nil
	}
	var numberValue float64
	if err := json.Unmarshal(raw, &numberValue); err == nil {
		b.Value = numberValue != 0
		return nil
	}
	b.Value = false
	return nil
}

func baseObjectiveRecoveryContract(input ObjectiveRecoveryContractInput) ObjectiveRecoveryContract {
	return ObjectiveRecoveryContract{
		ContractVersion:        ContractVersion,
		Projected:              true,
		Available:              true,
		Recommended:            input.RecoveryRecommended,
		FinalAnswerRecommended: input.FinalAnswerRecommended,
		Status:                 VerificationBlocked,
		Mode:                   "objective_recovery_contract",
		RunnerEffect:           "none",
		PromptEffect:           "none",
		ContractRef:            firstDisplaySafeRef(input.ContractRef, "contract:objective_recovery"),
		SourceRef:              normalizeOneDisplaySafeRef(input.SourceRef),
		Producer:               firstDisplaySafeRef(input.Producer, "host:objective_recovery_contract"),
		ObjectiveID:            objectiveGraphSafeID(input.ObjectiveID),
		CurrentStrategyRef:     normalizeOneDisplaySafeRef(input.CurrentStrategyRef),
		EvidenceRefs:           normalizeEvidenceRefs(input.EvidenceRefs),
		MissingInputs:          normalizeMissingInputs(input.MissingInputs),
		FailureClass:           NormalizeFailureClass(string(input.FailureClass)),
		FailureReason:          managedObjectiveReplannerSafeReason(input.FailureReason),
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_recovery_contract",
				"recovery_projection_only",
				"display_safe_refs_only",
				"no_runner_dispatch",
				"no_runtime_adapter_execution",
				"host_must_apply_recovery_replan",
			},
			input.Boundaries...,
		),
		NextHostAction:  "provide_objective_recovery_contract",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func baseObjectiveRecoveryContractJSONDecodeReport(input ObjectiveRecoveryContractJSONDecodeInput) ObjectiveRecoveryContractJSONDecodeReport {
	return ObjectiveRecoveryContractJSONDecodeReport{
		ContractVersion: ContractVersion,
		Available:       true,
		Status:          VerificationBlocked,
		Mode:            "objective_recovery_contract_json_decode",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		Boundaries: AppendBoundaries(
			[]Boundary{
				"objective_recovery_contract_json_decode",
				"recovery_projection_only",
				"no_runner_dispatch",
			},
			input.Boundaries...,
		),
		NextHostAction:  "provide_objective_recovery_contract",
		RawOutputLoaded: input.RawOutputLoaded,
	}
}

func objectiveRecoveryContractBlock(result ObjectiveRecoveryContract, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveRecoveryContract {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func objectiveRecoveryContractJSONDecodeBlock(result ObjectiveRecoveryContractJSONDecodeReport, status VerificationStatus, failure FailureClass, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveRecoveryContractJSONDecodeReport {
	result.Status = status
	result.FailureClass = failure
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.NextHostAction = next
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	return result.Normalize()
}

func objectiveRecoveryTargetsWithFallbacks(targets []ObjectiveRecoveryTarget, input ObjectiveRecoveryContractInput) []ObjectiveRecoveryTarget {
	out := make([]ObjectiveRecoveryTarget, 0, len(targets))
	for _, target := range targets {
		normalized := target.Normalize()
		if len(normalized.SuggestedStrategyRefs) == 0 {
			normalized.SuggestedStrategyRefs = normalizeDisplaySafeRefs(input.SuggestedStrategyRefs)
		}
		if len(normalized.SuggestedCapabilityRefs) == 0 {
			normalized.SuggestedCapabilityRefs = normalizeDisplaySafeRefs(input.SuggestedCapabilityRefs)
		}
		if len(normalized.SuggestedToolRefs) == 0 {
			normalized.SuggestedToolRefs = normalizeDisplaySafeRefs(input.SuggestedToolRefs)
		}
		if normalized.MissingEvidence.Ref == "" && normalized.MissingInput != "" {
			if ref, ok := NormalizeDisplaySafeRef(string(normalized.MissingInput)); ok {
				normalized.MissingEvidence = EvidenceRef{
					Ref:      ref,
					Kind:     objectiveRecoveryEvidenceKindFromMissingInput(normalized.MissingInput),
					Strength: EvidenceMissing,
					Source:   firstDisplaySafeRef(input.SourceRef, input.Producer, "host:objective_recovery_contract"),
				}.Normalize()
			}
		}
		if objectiveRecoveryTargetEmpty(normalized) {
			continue
		}
		out = append(out, normalized)
	}
	return normalizeObjectiveRecoveryTargets(out)
}

func normalizeObjectiveRecoveryTargets(in []ObjectiveRecoveryTarget) []ObjectiveRecoveryTarget {
	out := make([]ObjectiveRecoveryTarget, 0, len(in))
	for _, target := range in {
		normalized := target.Normalize()
		if objectiveRecoveryTargetEmpty(normalized) {
			continue
		}
		out = append(out, normalized)
	}
	return out
}

func objectiveRecoveryTargetEmpty(target ObjectiveRecoveryTarget) bool {
	return target.TargetRef == "" &&
		target.SubjectRef == "" &&
		target.MissingInput == "" &&
		target.MissingEvidence.Ref == "" &&
		len(target.SuggestedStrategyRefs) == 0 &&
		len(target.SuggestedCapabilityRefs) == 0 &&
		len(target.SuggestedToolRefs) == 0
}

func objectiveRecoveryReplannerSource(contract ObjectiveRecoveryContract) ReplannerSourceProjection {
	candidate := firstObjectiveRecoveryCandidateRef(contract)
	return BuildReplannerSourceProjection(ReplannerSourceInput{
		SourceKind:           ReplannerSourceRecovery,
		SourceRef:            firstDisplaySafeRef(contract.SourceRef, contract.ContractRef),
		Producer:             contract.Producer,
		Status:               contract.Status,
		FailureClass:         firstFailureClass(contract.FailureClass, FailureEvidenceMissing),
		FailureReason:        contract.FailureReason,
		CandidateStrategyRef: candidate,
		CapabilityRefs: mergeRecoveryDisplayRefs(
			mergeRecoveryDisplayRefs(contract.SuggestedCapabilityRefs, contract.SuggestedToolRefs),
			contract.SuggestedStrategyRefs,
		),
		EvidenceRefs:   contract.EvidenceRefs,
		SourceRefs:     []DisplaySafeRef{contract.ContractRef},
		ProposalRefs:   []DisplaySafeRef{firstDisplaySafeRef(contract.ReplanProposal.ProposalRef, "proposal:objective_recovery")},
		MissingInputs:  contract.MissingInputs,
		Boundaries:     AppendBoundaries(contract.Boundaries, "objective_recovery_replanner_source"),
		NextHostAction: "request_host_replanner_decision",
	})
}

func objectiveRecoveryReplanProposal(contract ObjectiveRecoveryContract) ObjectiveReplanProposal {
	steps := make([]ObjectiveReplanProposalStep, 0, len(contract.Targets))
	for index, target := range contract.Targets {
		step := ObjectiveReplanProposalStep{
			ContractVersion:  ContractVersion,
			StepRef:          objectiveRecoveryStepRef(contract, target, index+1),
			Action:           ObjectiveReplanProposalActionAddEvidenceNode,
			Owner:            "host",
			CurrentStrategy:  contract.CurrentStrategyRef,
			NextStrategy:     firstObjectiveRecoveryCandidateRef(contract),
			CapabilityRefs:   mergeRecoveryDisplayRefs(mergeRecoveryDisplayRefs(target.SuggestedCapabilityRefs, target.SuggestedToolRefs), target.SuggestedStrategyRefs),
			RequiredEvidence: normalizeEvidenceRefs([]EvidenceRef{target.MissingEvidence}),
			MissingInputs:    normalizeMissingInputs([]MissingInput{target.MissingInput}),
			EvidenceRefs:     contract.EvidenceRefs,
			DecisionBasis: normalizeDisplaySafeRefs([]DisplaySafeRef{
				"objective_recovery_contract",
				"recovery_target:" + DisplaySafeRef(itoaSmall(index+1)),
			}),
			Boundaries:     AppendBoundaries(contract.Boundaries, target.Boundaries...),
			NextHostAction: "host_may_add_evidence_node",
		}.Normalize()
		steps = append(steps, step)
	}
	if len(steps) == 0 {
		steps = append(steps, ObjectiveReplanProposalStep{
			ContractVersion: ContractVersion,
			StepRef:         "replan_step:objective_recovery_add_evidence",
			Action:          ObjectiveReplanProposalActionAddEvidenceNode,
			Owner:           "host",
			CurrentStrategy: contract.CurrentStrategyRef,
			NextStrategy:    firstObjectiveRecoveryCandidateRef(contract),
			CapabilityRefs: mergeRecoveryDisplayRefs(
				mergeRecoveryDisplayRefs(contract.SuggestedCapabilityRefs, contract.SuggestedToolRefs),
				contract.SuggestedStrategyRefs,
			),
			RequiredEvidence: contract.EvidenceRefs,
			MissingInputs:    contract.MissingInputs,
			EvidenceRefs:     contract.EvidenceRefs,
			DecisionBasis:    []DisplaySafeRef{"objective_recovery_contract"},
			Boundaries:       contract.Boundaries,
			NextHostAction:   "host_may_add_evidence_node",
		}.Normalize())
	}
	return ObjectiveReplanProposal{
		ContractVersion:    ContractVersion,
		Projected:          true,
		ProposalRef:        firstDisplaySafeRef(contract.ContractRef, "proposal:objective_recovery"),
		Status:             contract.Status,
		Action:             ObjectiveReplanProposalActionAddEvidenceNode,
		ObjectiveID:        contract.ObjectiveID,
		CurrentStrategyRef: contract.CurrentStrategyRef,
		NextStrategyRef:    firstObjectiveRecoveryCandidateRef(contract),
		NextOwner:          "host",
		Steps:              steps,
		EvidenceRefs:       contract.EvidenceRefs,
		MissingInputs:      contract.MissingInputs,
		DecisionBasis: []DisplaySafeRef{
			"objective_recovery_contract",
			"objective_recovery_replan_proposal",
		},
		Boundaries:      AppendBoundaries(contract.Boundaries, "objective_recovery_replan_proposal"),
		NextHostAction:  "host_may_add_evidence_node",
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: contract.RawOutputLoaded,
	}.Normalize()
}

func objectiveRecoveryStepRef(contract ObjectiveRecoveryContract, target ObjectiveRecoveryTarget, index int) DisplaySafeRef {
	if target.TargetRef != "" {
		return DisplaySafeRef("replan_step:recovery_" + string(target.TargetRef))
	}
	if target.MissingEvidence.Ref != "" {
		return DisplaySafeRef("replan_step:recovery_" + string(target.MissingEvidence.Ref))
	}
	if contract.ObjectiveID != "" {
		return DisplaySafeRef("replan_step:" + objectiveGraphSafeID(contract.ObjectiveID) + "_recovery_" + itoaSmall(index))
	}
	return DisplaySafeRef("replan_step:objective_recovery_" + itoaSmall(index))
}

func firstObjectiveRecoveryCandidateRef(contract ObjectiveRecoveryContract) DisplaySafeRef {
	for _, group := range [][]DisplaySafeRef{contract.SuggestedStrategyRefs, contract.SuggestedToolRefs, contract.SuggestedCapabilityRefs} {
		for _, ref := range normalizeDisplaySafeRefs(group) {
			return ref
		}
	}
	for _, target := range contract.Targets {
		for _, group := range [][]DisplaySafeRef{target.SuggestedStrategyRefs, target.SuggestedToolRefs, target.SuggestedCapabilityRefs} {
			for _, ref := range normalizeDisplaySafeRefs(group) {
				return ref
			}
		}
	}
	return ""
}

func mergeRecoveryDisplayRefs(groups ...[]DisplaySafeRef) []DisplaySafeRef {
	out := []DisplaySafeRef{}
	for _, group := range groups {
		out = append(out, group...)
	}
	return normalizeDisplaySafeRefs(out)
}

func objectiveRecoveryEvidenceKindFromMissingInput(missing MissingInput) string {
	token := normalizeControlToken(strings.TrimPrefix(string(missing), "evidence:"))
	if token == "" {
		return "recovery_evidence"
	}
	return token
}

func decodeObjectiveRecoveryJSON(raw []byte) (objectiveRecoveryJSONContract, bool) {
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return objectiveRecoveryJSONContract{}, false
	}
	for _, key := range []string{"answer_contract", "objective_recovery_contract", "recovery_contract"} {
		if nested, ok := envelope[key]; ok && len(bytes.TrimSpace(nested)) > 0 {
			var decoded objectiveRecoveryJSONContract
			if err := json.Unmarshal(nested, &decoded); err == nil {
				return decoded, true
			}
			return objectiveRecoveryJSONContract{}, false
		}
	}
	var decoded objectiveRecoveryJSONContract
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return objectiveRecoveryJSONContract{}, false
	}
	return decoded, true
}

func objectiveRecoveryContractInputFromJSON(decoded objectiveRecoveryJSONContract, input ObjectiveRecoveryContractJSONDecodeInput) ObjectiveRecoveryContractInput {
	targets := append([]objectiveRecoveryJSONTarget{}, decoded.RecoveryTargets...)
	targets = append(targets, decoded.Targets...)
	out := ObjectiveRecoveryContractInput{
		ContractRef:             firstDisplaySafeRef(input.ContractRef, "contract:objective_recovery"),
		SourceRef:               input.SourceRef,
		Producer:                input.Producer,
		ObjectiveID:             input.ObjectiveID,
		CurrentStrategyRef:      input.CurrentStrategyRef,
		RecoveryRecommended:     decoded.RecoveryRecommended.Value,
		FinalAnswerRecommended:  decoded.FinalAnswerRecommended.Value,
		FailureClass:            NormalizeFailureClass(decoded.FailureClass),
		FailureReason:           firstNonEmptyContractString(decoded.RecoveryReason, decoded.Reason),
		SuggestedStrategyRefs:   DisplaySafeRefs(decoded.SuggestedStrategyRefs),
		SuggestedCapabilityRefs: DisplaySafeRefs(decoded.SuggestedCapabilityRefs),
		SuggestedToolRefs:       DisplaySafeRefs(append(decoded.SuggestedToolRefs, decoded.SuggestedRecoveryTools...)),
		Boundaries:              input.Boundaries,
		RawOutputLoaded:         input.RawOutputLoaded,
	}
	for _, target := range targets {
		out.Targets = append(out.Targets, objectiveRecoveryTargetFromJSON(target, out))
	}
	return out
}

func objectiveRecoveryTargetFromJSON(target objectiveRecoveryJSONTarget, input ObjectiveRecoveryContractInput) ObjectiveRecoveryTarget {
	missing := objectiveRecoveryMissingInputFromJSON(target)
	return ObjectiveRecoveryTarget{
		ContractVersion:         ContractVersion,
		TargetRef:               DisplaySafeRef(target.TargetRef),
		SubjectRef:              DisplaySafeRef(target.SubjectRef),
		MissingInput:            missing,
		MissingEvidence:         objectiveRecoveryMissingEvidenceFromJSON(target, missing, input),
		FailureClass:            NormalizeFailureClass(target.FailureClass),
		FailureReason:           firstNonEmptyContractString(target.FailureCode),
		SuggestedStrategyRefs:   DisplaySafeRefs(target.SuggestedStrategyRefs),
		SuggestedCapabilityRefs: DisplaySafeRefs(target.SuggestedCapabilityRefs),
		SuggestedToolRefs:       DisplaySafeRefs(append(target.SuggestedToolRefs, target.SuggestedTools...)),
		Boundaries:              []Boundary{"objective_recovery_target"},
	}
}

func objectiveRecoveryMissingInputFromJSON(target objectiveRecoveryJSONTarget) MissingInput {
	if missing := firstMissingInput(MissingInput(target.MissingInput), ""); missing != "" {
		return missing
	}
	if ref, ok := NormalizeDisplaySafeRef(target.MissingEvidenceRef); ok {
		return MissingInput(ref)
	}
	if dimension := normalizeControlToken(target.MissingDimension); dimension != "" {
		return MissingInput("evidence:" + dimension)
	}
	return ""
}

func objectiveRecoveryMissingEvidenceFromJSON(target objectiveRecoveryJSONTarget, missing MissingInput, input ObjectiveRecoveryContractInput) EvidenceRef {
	ref := DisplaySafeRef("")
	if target.MissingEvidenceRef != "" {
		ref, _ = NormalizeDisplaySafeRef(target.MissingEvidenceRef)
	}
	if ref == "" && missing != "" {
		ref, _ = NormalizeDisplaySafeRef(string(missing))
	}
	kind := firstNonEmptyContractString(target.MissingEvidenceKind, strings.TrimPrefix(string(missing), "evidence:"), "recovery_evidence")
	return EvidenceRef{
		Ref:      ref,
		Kind:     kind,
		Strength: EvidenceMissing,
		Source:   firstDisplaySafeRef(input.SourceRef, input.Producer, "host:objective_recovery_contract"),
	}.Normalize()
}

func objectiveRecoveryContractInputUnsafe(input ObjectiveRecoveryContractInput) bool {
	if input.RawOutputLoaded ||
		displaySafeRefRejected(input.ContractRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefRejected(input.Producer) ||
		displaySafeRefRejected(input.CurrentStrategyRef) ||
		displaySafeRefSliceRejected(input.SuggestedStrategyRefs) ||
		displaySafeRefSliceRejected(input.SuggestedCapabilityRefs) ||
		displaySafeRefSliceRejected(input.SuggestedToolRefs) ||
		evidenceRefRejected(input.EvidenceRefs) {
		return true
	}
	for _, target := range input.Targets {
		if objectiveRecoveryTargetUnsafe(target) {
			return true
		}
	}
	return false
}

func objectiveRecoveryTargetUnsafe(target ObjectiveRecoveryTarget) bool {
	return target.RawOutputLoaded ||
		displaySafeRefRejected(target.TargetRef) ||
		displaySafeRefRejected(target.SubjectRef) ||
		evidenceRefRejected([]EvidenceRef{target.MissingEvidence}) ||
		displaySafeRefSliceRejected(target.SuggestedStrategyRefs) ||
		displaySafeRefSliceRejected(target.SuggestedCapabilityRefs) ||
		displaySafeRefSliceRejected(target.SuggestedToolRefs)
}

func objectiveRecoveryContractJSONDecodeInputUnsafe(input ObjectiveRecoveryContractJSONDecodeInput) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.ContractRef) ||
		displaySafeRefRejected(input.SourceRef) ||
		displaySafeRefRejected(input.Producer) ||
		displaySafeRefRejected(input.CurrentStrategyRef)
}
