package controlcontract

import (
	"fmt"
	"strings"
)

type ObjectiveReplanGraphPatchInput struct {
	PatchRef        DisplaySafeRef          `json:"patch_ref,omitempty"`
	SourceGraphRef  DisplaySafeRef          `json:"source_graph_ref,omitempty"`
	SourceNodeRef   DisplaySafeRef          `json:"source_node_ref,omitempty"`
	Proposal        ObjectiveReplanProposal `json:"proposal,omitempty"`
	Boundaries      []Boundary              `json:"boundaries,omitempty"`
	RawOutputLoaded bool                    `json:"raw_output_loaded"`
}

// ObjectiveReplanGraphPatch projects a replan proposal into host-reviewable
// recovery nodes. It never mutates an ObjectiveGraph or dispatches a runner.
type ObjectiveReplanGraphPatch struct {
	ContractVersion    string                        `json:"contract_version,omitempty"`
	Projected          bool                          `json:"projected"`
	Available          bool                          `json:"available"`
	ReadyForHostReview bool                          `json:"ready_for_host_review"`
	ReadyForGraphApply bool                          `json:"ready_for_graph_apply"`
	Status             VerificationStatus            `json:"status,omitempty"`
	Action             ObjectiveReplanProposalAction `json:"action,omitempty"`
	PatchRef           DisplaySafeRef                `json:"patch_ref,omitempty"`
	SourceGraphRef     DisplaySafeRef                `json:"source_graph_ref,omitempty"`
	SourceNodeRef      DisplaySafeRef                `json:"source_node_ref,omitempty"`
	ProposalRef        DisplaySafeRef                `json:"proposal_ref,omitempty"`
	PatchNodes         []ObjectiveNode               `json:"patch_nodes,omitempty"`
	EvidenceRefs       []EvidenceRef                 `json:"evidence_refs,omitempty"`
	MissingInputs      []MissingInput                `json:"missing_inputs,omitempty"`
	BlockedReasons     []string                      `json:"blocked_reasons,omitempty"`
	Boundaries         []Boundary                    `json:"boundaries,omitempty"`
	NextHostAction     NextHostAction                `json:"next_host_action,omitempty"`
	RunnerEffect       string                        `json:"runner_effect,omitempty"`
	PromptEffect       string                        `json:"prompt_effect,omitempty"`
	RawOutputLoaded    bool                          `json:"raw_output_loaded"`
}

func BuildObjectiveReplanGraphPatch(input ObjectiveReplanGraphPatchInput) ObjectiveReplanGraphPatch {
	rawProposal := input.Proposal
	proposal := rawProposal.Normalize()
	result := ObjectiveReplanGraphPatch{
		ContractVersion: ContractVersion,
		Projected:       true,
		Available:       true,
		Status:          proposal.Status,
		Action:          proposal.Action,
		PatchRef:        firstDisplaySafeRef(input.PatchRef, objectiveReplanGraphPatchDefaultRef(proposal)),
		SourceGraphRef:  normalizeOneDisplaySafeRef(input.SourceGraphRef),
		SourceNodeRef:   normalizeOneDisplaySafeRef(input.SourceNodeRef),
		ProposalRef:     proposal.ProposalRef,
		EvidenceRefs:    cloneEvidenceRefs(proposal.EvidenceRefs),
		MissingInputs:   cloneMissingInputs(proposal.MissingInputs),
		Boundaries: MergeBoundaries(
			[]Boundary{
				"objective_replan_graph_patch",
				"objective_replan_graph_patch_proposal_only",
				"host_must_review_recovery_nodes",
				"host_must_apply_graph_patch",
				"no_graph_mutation_by_core",
				"no_runner_dispatch",
				"no_strategy_dispatch",
				"no_runtime_adapter_execution",
			},
			input.Boundaries,
			proposal.Boundaries,
		),
		NextHostAction:  proposal.NextHostAction,
		RunnerEffect:    "none",
		PromptEffect:    "none",
		RawOutputLoaded: input.RawOutputLoaded || proposal.RawOutputLoaded,
	}
	if objectiveReplanGraphPatchInputUnsafe(input, rawProposal) {
		return objectiveReplanGraphPatchBlock(result, VerificationReviewRequired, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if proposal.Action == ObjectiveReplanProposalActionReviewRefs {
		return objectiveReplanGraphPatchBlock(result, VerificationReviewRequired, "host:display_safe_refs", "provide_display_safe_refs", "raw_output_not_allowed")
	}
	if !ObjectiveReplanProposalRequiresHostDispatchBinding(&proposal) {
		result.Status = VerificationNotApplicable
		result.NextHostAction = firstNextHostAction(proposal.NextHostAction, objectiveReplanProposalNextHostAction(proposal.Action))
		result.BlockedReasons = append(result.BlockedReasons, "replan_action_does_not_require_graph_patch")
		result.Boundaries = AppendBoundaries(result.Boundaries, "objective_replan_graph_patch_not_required")
		return result.Normalize()
	}
	result.PatchNodes = objectiveReplanGraphPatchNodes(proposal)
	if len(result.PatchNodes) == 0 {
		return objectiveReplanGraphPatchBlock(result, VerificationBlocked, "host:objective_replan_graph_patch_node", "review_replan_proposal", "objective_replan_graph_patch_node_missing")
	}
	result.ReadyForHostReview = true
	result.ReadyForGraphApply = false
	result.Status = VerificationPartial
	result.NextHostAction = "review_objective_replan_graph_patch"
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, "host:objective_replan_graph_patch_review")
	result.Boundaries = AppendBoundaries(result.Boundaries, "host_must_bind_recovery_node_before_runtime")
	return result.Normalize()
}

func CloneObjectiveReplanGraphPatch(in ObjectiveReplanGraphPatch) ObjectiveReplanGraphPatch {
	out := in
	out.PatchNodes = cloneObjectiveNodes(in.PatchNodes)
	out.EvidenceRefs = cloneEvidenceRefs(in.EvidenceRefs)
	out.MissingInputs = cloneMissingInputs(in.MissingInputs)
	out.BlockedReasons = cloneStringSlice(in.BlockedReasons)
	out.Boundaries = cloneBoundaries(in.Boundaries)
	return out
}

func (p ObjectiveReplanGraphPatch) Clone() ObjectiveReplanGraphPatch {
	return CloneObjectiveReplanGraphPatch(p)
}

func (p ObjectiveReplanGraphPatch) Normalize() ObjectiveReplanGraphPatch {
	out := p.Clone()
	out.ContractVersion = defaultContractVersion(out.ContractVersion)
	out.Projected = true
	out.Available = true
	out.Status = NormalizeVerificationStatus(string(out.Status))
	if out.Status == VerificationNotEvaluated {
		out.Status = VerificationBlocked
	}
	out.Action = NormalizeObjectiveReplanProposalAction(string(out.Action))
	out.PatchRef = firstDisplaySafeRef(out.PatchRef, "patch:objective_replan_graph")
	out.SourceGraphRef = normalizeOneDisplaySafeRef(out.SourceGraphRef)
	out.SourceNodeRef = normalizeOneDisplaySafeRef(out.SourceNodeRef)
	out.ProposalRef = normalizeOneDisplaySafeRef(out.ProposalRef)
	out.PatchNodes = normalizeObjectiveNodes(out.PatchNodes)
	out.EvidenceRefs = normalizeEvidenceRefs(out.EvidenceRefs)
	out.MissingInputs = normalizeMissingInputs(out.MissingInputs)
	out.BlockedReasons = normalizeStringList(out.BlockedReasons)
	out.Boundaries = normalizeBoundaries(out.Boundaries)
	out.NextHostAction = NormalizeNextHostAction(string(out.NextHostAction))
	out.RunnerEffect = normalizeControlToken(out.RunnerEffect)
	out.PromptEffect = normalizeControlToken(out.PromptEffect)
	if out.RunnerEffect == "" {
		out.RunnerEffect = "none"
	}
	if out.PromptEffect == "" {
		out.PromptEffect = "none"
	}
	if out.RawOutputLoaded {
		out.Status = VerificationReviewRequired
		out.ReadyForHostReview = false
		out.ReadyForGraphApply = false
		out.MissingInputs = AppendMissingInputs(out.MissingInputs, "host:display_safe_refs")
		out.Boundaries = AppendBoundaries(out.Boundaries, "raw_output_not_allowed")
		out.NextHostAction = "provide_display_safe_refs"
	}
	return out
}

func objectiveReplanGraphPatchNodes(proposal ObjectiveReplanProposal) []ObjectiveNode {
	nodes := []ObjectiveNode{}
	for i, step := range proposal.Steps {
		step = step.Normalize()
		if !objectiveReplanGraphPatchStepRequiresNode(step) {
			continue
		}
		node := ObjectiveNode{
			NodeRef:          objectiveReplanGraphPatchNodeRef(step, i),
			Kind:             "objective_replan_recovery_node",
			State:            ObjectiveNodeStatePending,
			Optional:         false,
			CapabilityRef:    objectiveReplanGraphPatchStepCapabilityRef(step),
			StrategyRef:      objectiveReplanGraphPatchStepStrategyRef(step, proposal),
			RequiredEvidence: cloneEvidenceRefs(step.RequiredEvidence),
			AttemptPolicy: ObjectiveNodeAttemptPolicy{
				MaxAttempts:             1,
				RetryableFailureClasses: []FailureClass{FailureEvidenceMissing, FailureEvidenceWeak, FailureVerificationFailed, FailureTimeout},
				NoProgressGate:          true,
				PolicyRefs:              []DisplaySafeRef{"policy:objective_replan_recovery_node"},
			},
			SideEffectClass:  ObjectiveCapabilitySideEffectUnspecified,
			RequiresApproval: false,
			PolicyRefs:       []DisplaySafeRef{"policy:host_review_required"},
			Boundaries: MergeBoundaries(
				[]Boundary{
					"objective_replan_graph_patch_node",
					Boundary("replan_action_" + string(step.Action)),
					"host_must_bind_recovery_node_before_runtime",
					"descriptor_binding_required",
					"no_runtime_adapter_execution",
				},
				step.Boundaries,
			),
			MissingInputs: MergeMissingInputs(
				step.MissingInputs,
				[]MissingInput{"host:capability_descriptor_binding"},
			),
		}
		if node.CapabilityRef == "" {
			node.MissingInputs = AppendMissingInputs(node.MissingInputs, "host:capability_ref")
		}
		if node.StrategyRef == "" {
			node.MissingInputs = AppendMissingInputs(node.MissingInputs, "host:strategy_ref")
		}
		if len(node.RequiredEvidence) == 0 {
			node.MissingInputs = AppendMissingInputs(node.MissingInputs, "host:required_evidence")
		}
		nodes = append(nodes, node.Normalize())
	}
	return normalizeObjectiveNodes(nodes)
}

func objectiveReplanGraphPatchStepRequiresNode(step ObjectiveReplanProposalStep) bool {
	return ObjectiveReplanProposalActionMetadataFor(step.Action).RequiresHostDispatchBinding
}

func objectiveReplanGraphPatchStepCapabilityRef(step ObjectiveReplanProposalStep) DisplaySafeRef {
	if len(step.CapabilityRefs) > 0 {
		return firstDisplaySafeRef(step.CapabilityRefs...)
	}
	if step.NextStrategy != "" && strings.HasPrefix(string(step.NextStrategy), "capability:") {
		return normalizeOneDisplaySafeRef(step.NextStrategy)
	}
	return ""
}

func objectiveReplanGraphPatchStepStrategyRef(step ObjectiveReplanProposalStep, proposal ObjectiveReplanProposal) DisplaySafeRef {
	return firstDisplaySafeRef(step.NextStrategy, proposal.NextStrategyRef, step.CurrentStrategy, proposal.CurrentStrategyRef)
}

func objectiveReplanGraphPatchNodeRef(step ObjectiveReplanProposalStep, index int) DisplaySafeRef {
	token := ""
	if step.StepRef != "" {
		token = strings.TrimPrefix(string(step.StepRef), "replan_step:")
		token = strings.TrimPrefix(token, "step:")
		token = normalizeControlToken(token)
	}
	if token == "" {
		token = normalizeControlToken(string(step.Action))
	}
	if token == "" {
		token = "step"
	}
	if len(token) > 80 {
		token = strings.Trim(token[:80], "_.:-")
	}
	return DisplaySafeRef(fmt.Sprintf("node:replan_%s_%02d", token, index+1))
}

func objectiveReplanGraphPatchDefaultRef(proposal ObjectiveReplanProposal) DisplaySafeRef {
	if proposal.ProposalRef != "" {
		token := strings.TrimPrefix(string(proposal.ProposalRef), "proposal:")
		token = normalizeControlToken(token)
		if token != "" {
			return DisplaySafeRef("patch:" + token)
		}
	}
	action := normalizeControlToken(string(proposal.Action))
	if action == "" {
		action = "proposal"
	}
	return DisplaySafeRef("patch:objective_replan_" + action)
}

func objectiveReplanGraphPatchInputUnsafe(input ObjectiveReplanGraphPatchInput, proposal ObjectiveReplanProposal) bool {
	return input.RawOutputLoaded ||
		displaySafeRefRejected(input.PatchRef) ||
		displaySafeRefRejected(input.SourceGraphRef) ||
		displaySafeRefRejected(input.SourceNodeRef) ||
		proposal.RawOutputLoaded ||
		displaySafeRefRejected(proposal.ProposalRef) ||
		displaySafeRefRejected(proposal.CurrentStrategyRef) ||
		displaySafeRefRejected(proposal.NextStrategyRef) ||
		displaySafeRefSliceRejected(proposal.CapabilityGapRefs) ||
		evidenceRefRejected(proposal.EvidenceRefs) ||
		displaySafeRefSliceRejected(proposal.DecisionBasis) ||
		displaySafeRefSliceRejected(proposal.PolicyRefs) ||
		objectiveReplanGraphPatchStepsUnsafe(proposal.Steps)
}

func objectiveReplanGraphPatchStepsUnsafe(steps []ObjectiveReplanProposalStep) bool {
	for _, step := range steps {
		if displaySafeRefRejected(step.StepRef) ||
			displaySafeRefRejected(step.CurrentStrategy) ||
			displaySafeRefRejected(step.NextStrategy) ||
			displaySafeRefSliceRejected(step.CapabilityRefs) ||
			evidenceRefRejected(step.RequiredEvidence) ||
			evidenceRefRejected(step.EvidenceRefs) ||
			displaySafeRefSliceRejected(step.DecisionBasis) {
			return true
		}
	}
	return false
}

func objectiveReplanGraphPatchBlock(result ObjectiveReplanGraphPatch, status VerificationStatus, missing MissingInput, next NextHostAction, boundary Boundary) ObjectiveReplanGraphPatch {
	result.Status = status
	result.ReadyForHostReview = false
	result.ReadyForGraphApply = false
	if boundary == "raw_output_not_allowed" {
		result.Action = ObjectiveReplanProposalActionReviewRefs
		result.RawOutputLoaded = true
	}
	result.MissingInputs = AppendMissingInputs(result.MissingInputs, missing)
	result.Boundaries = AppendBoundaries(result.Boundaries, boundary)
	result.NextHostAction = next
	return result.Normalize()
}
