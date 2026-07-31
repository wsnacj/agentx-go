package lowering_test

import (
	"errors"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	lowering "github.com/wsnacj/agentx-go/runtime/workflow/lowering"
)

func TestDependenciesAreRequired(t *testing.T) {
	if _, err := lowering.LowerSpec(workflow.Spec{}, lowering.Dependencies{}); err == nil ||
		err.Error() != "workflow lowering: validator is required" {
		t.Fatalf("LowerSpec() error = %v", err)
	}
	if _, err := lowering.LowerSpec(workflow.Spec{}, lowering.Dependencies{
		Validator: acceptingValidator{},
	}); err == nil || err.Error() != "workflow lowering: mapper is required" {
		t.Fatalf("LowerSpec() error = %v", err)
	}
}

func TestExternalConsumerCanInjectPolicyAndBuildPlan(t *testing.T) {
	plan, err := lowering.LowerSpec(workflow.Spec{
		ID:        "external",
		Version:   "v1",
		EntryNode: "start",
		Nodes: []workflow.NodeSpec{{
			ID:   "start",
			Kind: workflow.NodeTool,
		}},
	}, lowering.Dependencies{
		Validator: acceptingValidator{},
		Mapper:    fixedMapper{},
	})
	if err != nil {
		t.Fatalf("LowerSpec(): %v", err)
	}
	if got := plan.Nodes["start"].Call; got.Name != "portable_tool" ||
		got.Arguments != `{"input":"value"}` {
		t.Fatalf("unexpected call: %#v", got)
	}
	if got := plan.OrchestrationPlan("runtime-id"); got.WorkflowID != "runtime-id" ||
		got.EntryNode != "start" {
		t.Fatalf("unexpected orchestration plan: %#v", got)
	}
}

func TestValidatorAndMapperErrorIdentityIsPreserved(t *testing.T) {
	sentinel := errors.New("host rejected")
	if _, err := lowering.LowerSpec(workflow.Spec{}, lowering.Dependencies{
		Validator: failingValidator{err: sentinel},
		Mapper:    fixedMapper{},
	}); !errors.Is(err, sentinel) {
		t.Fatalf("validator error identity lost: %v", err)
	}
	if _, err := lowering.LowerNode(workflow.NodeSpec{ID: "node"}, lowering.Dependencies{
		Validator: acceptingValidator{},
		Mapper:    failingMapper{err: sentinel},
	}); !errors.Is(err, sentinel) {
		t.Fatalf("mapper error identity lost: %v", err)
	}
}

type acceptingValidator struct{}

func (acceptingValidator) ValidateSpec(workflow.Spec) error {
	return nil
}

func (acceptingValidator) ValidateNode(workflow.NodeSpec) error {
	return nil
}

type failingValidator struct {
	err error
}

func (v failingValidator) ValidateSpec(workflow.Spec) error {
	return v.err
}

func (v failingValidator) ValidateNode(workflow.NodeSpec) error {
	return v.err
}

type fixedMapper struct{}

func (fixedMapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (lowering.MappedCall, error) {
	return lowering.MappedCall{
		Name:      "portable_tool",
		Arguments: map[string]any{"input": "value"},
	}, nil
}

type failingMapper struct {
	err error
}

func (m failingMapper) MapNode(workflow.NodeSpec, workflow.ExecutionMode) (lowering.MappedCall, error) {
	return lowering.MappedCall{}, m.err
}
