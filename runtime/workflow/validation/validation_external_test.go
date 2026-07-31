package validation_test

import (
	"errors"
	"reflect"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	workflowvalidation "github.com/wsnacj/agentx-go/runtime/workflow/validation"
)

func TestPolicyIsRequired(t *testing.T) {
	if err := workflowvalidation.ValidateSpec(workflow.Spec{}, nil); err == nil ||
		err.Error() != "workflow validation: policy is required" {
		t.Fatalf("ValidateSpec(nil policy) error = %v", err)
	}
	if err := workflowvalidation.ValidateNodeSpec(workflow.NodeSpec{}, nil); err == nil ||
		err.Error() != "workflow validation: policy is required" {
		t.Fatalf("ValidateNodeSpec(nil policy) error = %v", err)
	}
}

func TestStructuralValidationWithExplicitPolicy(t *testing.T) {
	policy := &recordingPolicy{}
	spec := workflow.Spec{
		ID:              "portable-structure",
		Pack:            "host-pack",
		DefaultContract: "host-contract",
		EntryNode:       "parallel",
		Nodes: []workflow.NodeSpec{
			{
				ID:   "parallel",
				Kind: workflow.NodeParallel,
				Outputs: []workflow.BindingSpec{{
					From: "result",
					To:   "state.parallel",
				}},
			},
			{
				ID:   "final",
				Kind: workflow.NodeTool,
				Inputs: []workflow.BindingSpec{{
					From: "node.parallel.output",
					To:   "args.input",
				}},
			},
		},
		Edges: []workflow.EdgeSpec{{
			From:      "parallel",
			To:        "final",
			Condition: "host-owned",
		}},
	}
	if err := workflowvalidation.ValidateSpec(spec, policy); err != nil {
		t.Fatalf("ValidateSpec(): %v", err)
	}
	want := []string{
		"pack-contract",
		"pack-metadata",
		"node-runtime:parallel",
		"binding-target:output",
		"node-config:parallel",
		"node-runtime:final",
		"binding-target:input",
		"node-config:final",
		"edge-runtime:parallel->final",
		"edge-determinism",
		"binding-source:output",
		"binding-source:input",
	}
	if !reflect.DeepEqual(policy.calls, want) {
		t.Fatalf("policy calls = %#v, want %#v", policy.calls, want)
	}
}

func TestStructuralErrorAndPolicyIdentity(t *testing.T) {
	sentinel := errors.New("host policy rejected node")
	policy := &recordingPolicy{nodeRuntimeError: sentinel}
	err := workflowvalidation.ValidateSpec(workflow.Spec{
		ID:        "policy-error",
		EntryNode: "node",
		Nodes: []workflow.NodeSpec{{
			ID:   "node",
			Kind: workflow.NodeTool,
		}},
	}, policy)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ValidateSpec() error = %v, want sentinel identity", err)
	}
	if got, want := err.Error(), `workflow: node "node" invalid: host policy rejected node`; got != want {
		t.Fatalf("ValidateSpec() error = %q, want %q", got, want)
	}

	err = workflowvalidation.ValidateSpec(workflow.Spec{
		ID:        "duplicate",
		EntryNode: "node",
		Nodes: []workflow.NodeSpec{
			{ID: "node", Kind: workflow.NodeTool},
			{ID: "node", Kind: workflow.NodeTool},
		},
	}, &recordingPolicy{})
	if got, want := err.Error(), `workflow: duplicate node id "node"`; got != want {
		t.Fatalf("duplicate error = %q, want %q", got, want)
	}
}

func TestFieldValidationContract(t *testing.T) {
	if err := workflowvalidation.ValidateTrimmedField(" value ", "field"); err == nil ||
		err.Error() != `workflow: field " value " must not include surrounding whitespace` {
		t.Fatalf("ValidateTrimmedField() error = %v", err)
	}
	if err := workflowvalidation.ValidateOptionalField("  ", "field"); err == nil ||
		err.Error() != `workflow: field "  " must not be whitespace-only` {
		t.Fatalf("ValidateOptionalField() error = %v", err)
	}
	if err := workflowvalidation.ValidateOptionalField("", "field"); err != nil {
		t.Fatalf("ValidateOptionalField(empty): %v", err)
	}
}

type recordingPolicy struct {
	calls            []string
	nodeRuntimeError error
}

func (p *recordingPolicy) ValidatePackScopedContractUsage(workflow.Spec) error {
	p.calls = append(p.calls, "pack-contract")
	return nil
}

func (p *recordingPolicy) ValidatePackScopedWorkflowMetadataUsage(workflow.Spec) error {
	p.calls = append(p.calls, "pack-metadata")
	return nil
}

func (p *recordingPolicy) ValidateNodeRuntimeCapabilities(node workflow.NodeSpec) error {
	p.calls = append(p.calls, "node-runtime:"+node.ID)
	return p.nodeRuntimeError
}

func (p *recordingPolicy) ValidateNodeConfig(node workflow.NodeSpec) error {
	p.calls = append(p.calls, "node-config:"+node.ID)
	return nil
}

func (p *recordingPolicy) ValidateEdgeRuntimeCapabilities(_ workflow.EdgeSpec, from string, to string) error {
	p.calls = append(p.calls, "edge-runtime:"+from+"->"+to)
	return nil
}

func (p *recordingPolicy) ValidateLinearRuntimeEdgeDeterminism([]workflow.EdgeSpec) error {
	p.calls = append(p.calls, "edge-determinism")
	return nil
}

func (p *recordingPolicy) ValidateReachableCycleRuntimeCapability(nodeID string) error {
	p.calls = append(p.calls, "cycle:"+nodeID)
	return nil
}

func (p *recordingPolicy) ValidateBindingTargetShape(_ string, label string) error {
	p.calls = append(p.calls, "binding-target:"+label)
	return nil
}

func (p *recordingPolicy) ValidateBindingSourceShape(_ string, label string) error {
	p.calls = append(p.calls, "binding-source:"+label)
	return nil
}
