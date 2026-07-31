package lowering

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

type recordingValidator struct {
	calls []string
	err   error
}

func (v *recordingValidator) ValidateSpec(workflow.Spec) error {
	v.calls = append(v.calls, "spec")
	return v.err
}

func (v *recordingValidator) ValidateNode(node workflow.NodeSpec) error {
	v.calls = append(v.calls, "node:"+node.ID)
	return v.err
}

type recordingMapper struct {
	calls []string
	value MappedCall
	err   error
}

func (m *recordingMapper) MapNode(node workflow.NodeSpec, mode workflow.ExecutionMode) (MappedCall, error) {
	m.calls = append(m.calls, node.ID+":"+string(mode))
	return m.value, m.err
}

func TestLowerSpecPreservesOrderAndBuildsOrchestrationPlan(t *testing.T) {
	validator := &recordingValidator{}
	mapper := &recordingMapper{value: MappedCall{
		Name:      "tool",
		Arguments: map[string]any{"value": "ok"},
	}}
	spec := workflow.Spec{
		ID:        "wf",
		Version:   "v1",
		EntryNode: "start",
		Nodes: []workflow.NodeSpec{{
			ID:     "start",
			Kind:   workflow.NodeTool,
			Config: map[string]any{"tool": "tool"},
		}},
		Edges:       []workflow.EdgeSpec{{From: "start", To: "end"}},
		StateSchema: []workflow.StateSlotSpec{{Name: "result", Type: "string"}},
	}

	plan, err := LowerSpec(spec, Dependencies{Validator: validator, Mapper: mapper})
	if err != nil {
		t.Fatalf("lower spec: %v", err)
	}
	if !reflect.DeepEqual(validator.calls, []string{"spec", "node:start"}) {
		t.Fatalf("unexpected validation order: %#v", validator.calls)
	}
	if !reflect.DeepEqual(mapper.calls, []string{"start:inline"}) {
		t.Fatalf("unexpected mapper calls: %#v", mapper.calls)
	}
	node := plan.Nodes["start"]
	if node.Call.Name != "tool" || node.Call.Arguments != `{"value":"ok"}` {
		t.Fatalf("unexpected lowered node: %#v", node)
	}
	runtimePlan := plan.OrchestrationPlan("runtime-wf")
	if runtimePlan.WorkflowID != "runtime-wf" || runtimePlan.Nodes["start"].Call != node.Call {
		t.Fatalf("unexpected orchestration plan: %#v", runtimePlan)
	}
}

func TestLowerNodePreservesNonEmptyExecutionMode(t *testing.T) {
	validator := &recordingValidator{}
	mapper := &recordingMapper{value: MappedCall{Name: "tool"}}
	node, err := LowerNode(workflow.NodeSpec{
		ID:            "step",
		Kind:          workflow.NodeTool,
		ExecutionMode: workflow.ExecutionMode(" "),
	}, Dependencies{Validator: validator, Mapper: mapper})
	if err != nil {
		t.Fatalf("lower node: %v", err)
	}
	if node.ExecutionMode != workflow.ExecutionMode(" ") {
		t.Fatalf("expected non-empty mode to be preserved, got %q", node.ExecutionMode)
	}
}

func TestLowerSpecWrapsNodeError(t *testing.T) {
	sentinel := errors.New("mapping failed")
	_, err := LowerSpec(workflow.Spec{
		Nodes: []workflow.NodeSpec{{ID: "broken", Kind: workflow.NodeTool}},
	}, Dependencies{
		Validator: &recordingValidator{},
		Mapper:    &recordingMapper{err: sentinel},
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped mapper error, got %v", err)
	}
	if !strings.Contains(err.Error(), `workflow: lower node "broken": mapping failed`) {
		t.Fatalf("unexpected error text: %v", err)
	}
}

func TestLowerNodeRejectsUnmarshalableArguments(t *testing.T) {
	_, err := LowerNode(workflow.NodeSpec{ID: "broken"}, Dependencies{
		Validator: &recordingValidator{},
		Mapper: &recordingMapper{value: MappedCall{
			Name:      "tool",
			Arguments: map[string]any{"bad": make(chan int)},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "marshal tool arguments") {
		t.Fatalf("unexpected error: %v", err)
	}
}
