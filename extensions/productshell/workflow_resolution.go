package productshell

import (
	"encoding/json"
	"fmt"
	"strings"

	agentxworkflow "github.com/wsnacj/agentx-go/runtime/workflow"

	"github.com/wsnacj/agentx-go/extensions/pack"
)

// WorkflowResolutionRuntime binds explicit or routed workflows while leaving
// registry ownership and materialization to the host.
type WorkflowResolutionRuntime struct {
	HasRegisteredPackFn             func(string) bool
	ResolvePackWorkflowFn           func(WorkflowBinding) (ResolvedWorkflow, error)
	ResolveExplicitPackBindingFn    func(string, string, string) (pack.Binding, bool, error)
	MaterializeRegisteredWorkflowFn func(pack.Binding) (agentxworkflow.Spec, error)
}

func (rt WorkflowResolutionRuntime) ResolveWorkflow(input Input) (ResolvedWorkflow, error) {
	if input.WorkflowSpec != nil {
		return rt.ResolveExplicitWorkflow(input, *input.WorkflowSpec)
	}
	for _, key := range []string{"workflow_spec", "workflowSpec", "workflow"} {
		raw, ok := input.Options[key]
		if !ok || raw == nil {
			continue
		}
		spec, err := DecodeWorkflowSpec(raw)
		if err != nil {
			return ResolvedWorkflow{}, fmt.Errorf("agentx: decode %s: %w", key, err)
		}
		return rt.ResolveExplicitWorkflow(input, *spec)
	}
	binding, hasBinding, err := ResolveWorkflowBinding(input)
	if err != nil {
		return ResolvedWorkflow{}, err
	}
	if !hasBinding {
		return ResolvedWorkflow{}, nil
	}
	if rt.ResolvePackWorkflowFn == nil {
		return ResolvedWorkflow{}, fmt.Errorf("agentx: pack registry is not configured")
	}
	return rt.ResolvePackWorkflowFn(binding)
}

func (rt WorkflowResolutionRuntime) ResolveExplicitWorkflow(input Input, spec agentxworkflow.Spec) (ResolvedWorkflow, error) {
	packID := strings.TrimSpace(spec.Pack)
	if packID == "" {
		if !ExplicitRawWorkflowOptIn(input) {
			return ResolvedWorkflow{}, fmt.Errorf("agentx: explicit raw workflow requires raw-workflow opt-in")
		}
		return ResolvedWorkflow{Spec: &spec}, nil
	}
	caseType := ExplicitWorkflowBindingCaseType(input, spec)
	if caseType == "" {
		return ResolvedWorkflow{}, fmt.Errorf("agentx: workflow pack %q requires case type to resolve pack-bound execution", packID)
	}
	if rt.HasRegisteredPackFn == nil || !rt.HasRegisteredPackFn(packID) || rt.ResolveExplicitPackBindingFn == nil {
		return ResolvedWorkflow{}, fmt.Errorf("agentx: workflow pack %q case type %q requires registered pack runtime for pack-bound execution", packID, caseType)
	}
	binding, ok, err := rt.ResolveExplicitPackBindingFn(packID, caseType, strings.TrimSpace(spec.ID))
	if err != nil {
		return ResolvedWorkflow{}, err
	}
	if !ok {
		return ResolvedWorkflow{}, fmt.Errorf("agentx: workflow pack %q case type %q workflow %q must resolve to a registered pack binding", packID, caseType, strings.TrimSpace(spec.ID))
	}
	if rt.MaterializeRegisteredWorkflowFn != nil {
		registeredSpec, err := rt.MaterializeRegisteredWorkflowFn(binding)
		if err != nil {
			return ResolvedWorkflow{}, err
		}
		equivalent, err := explicitWorkflowPackBindingEquivalent(spec, registeredSpec)
		if err != nil {
			return ResolvedWorkflow{}, fmt.Errorf("agentx: normalize explicit workflow pack binding: %w", err)
		}
		if !equivalent {
			return ResolvedWorkflow{}, fmt.Errorf("agentx: explicit workflow pack %q case type %q workflow %q must match registered pack workflow execution semantics", strings.TrimSpace(binding.PackID), strings.TrimSpace(binding.CaseType), strings.TrimSpace(binding.WorkflowID))
		}
	}
	spec = binding.Workflow
	return ResolvedWorkflow{Spec: &spec, PackBinding: &binding}, nil
}

func explicitWorkflowPackBindingEquivalent(left, right agentxworkflow.Spec) (bool, error) {
	leftPayload, err := json.Marshal(normalizeExplicitWorkflowPackBindingSpec(left))
	if err != nil {
		return false, err
	}
	rightPayload, err := json.Marshal(normalizeExplicitWorkflowPackBindingSpec(right))
	if err != nil {
		return false, err
	}
	return string(leftPayload) == string(rightPayload), nil
}

func normalizeExplicitWorkflowPackBindingSpec(spec agentxworkflow.Spec) agentxworkflow.Spec {
	normalized := agentxworkflow.Spec{
		ID: strings.TrimSpace(spec.ID), Version: strings.TrimSpace(spec.Version), Pack: strings.TrimSpace(spec.Pack),
		CaseTypes: append([]string(nil), spec.CaseTypes...), PlanningMode: spec.PlanningMode,
		EntryNode: strings.TrimSpace(spec.EntryNode), Edges: append([]agentxworkflow.EdgeSpec(nil), spec.Edges...),
		StateSchema:     append([]agentxworkflow.StateSlotSpec(nil), spec.StateSchema...),
		ArtifactSchema:  append([]agentxworkflow.ArtifactTypeRef(nil), spec.ArtifactSchema...),
		EvaluatorSchema: append([]agentxworkflow.EvaluatorRef(nil), spec.EvaluatorSchema...),
		DefaultContract: strings.TrimSpace(spec.DefaultContract),
	}
	for _, node := range spec.Nodes {
		node.ID = strings.TrimSpace(node.ID)
		node.Title = ""
		node.Description = ""
		node.ContractRef = strings.TrimSpace(node.ContractRef)
		normalized.Nodes = append(normalized.Nodes, node)
	}
	return normalized
}
