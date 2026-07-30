package bindingstate

import (
	"strings"
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
)

func TestRuntimeBindingStateTransitionContract(t *testing.T) {
	runtime := New(Inputs{
		SessionInput: map[string]any{"request": "risk-alert"},
		CaseInput:    map[string]any{"ticket_id": "case-42"},
	})
	args, err := runtime.MaterializeArguments("prepare", `{}`, []workflow.BindingSpec{
		{From: "session.input.request", To: "args.input"},
	})
	if err != nil || args != `{"input":"risk-alert"}` {
		t.Fatalf("materialize prepare arguments: args=%s err=%v", args, err)
	}
	if err := runtime.ApplyNodeOutputs("prepare", []workflow.BindingSpec{
		{From: "result.summary", To: "state.summary"},
	}, NewNodeResult("completed", `{"summary":"ready"}`, "")); err != nil {
		t.Fatalf("apply prepare outputs: %v", err)
	}
	args, err = runtime.MaterializeArguments("submit", `{}`, []workflow.BindingSpec{
		{From: "state.summary", To: "args.input"},
		{From: "case.input.ticket_id", To: "args.context.ticket_id"},
		{From: "case.input.optional", To: "args.optional", Optional: true},
	})
	if err != nil || args != `{"context":{"ticket_id":"case-42"},"input":"ready"}` {
		t.Fatalf("materialize submit arguments: args=%s err=%v", args, err)
	}
	if err := runtime.ApplyNodeOutputs("submit", []workflow.BindingSpec{
		{From: "status", To: "state.final_status"},
		{From: "result.optional", To: "state.optional", Optional: true},
	}, NewNodeResult("completed", `{"status":"ok"}`, "")); err != nil {
		t.Fatalf("apply submit outputs: %v", err)
	}
	if err := runtime.ValidateRequiredSlots([]workflow.StateSlotSpec{
		{Name: "summary", Required: true},
		{Name: "final_status", Required: true},
	}); err != nil {
		t.Fatalf("validate required slots: %v", err)
	}
	state := runtime.State()
	if state["summary"] != "ready" || state["final_status"] != "completed" {
		t.Fatalf("unexpected state: %#v", state)
	}
	state["summary"] = "mutated"
	if runtime.State()["summary"] != "ready" {
		t.Fatalf("State must return an isolated snapshot")
	}
}

func TestRuntimeBindingErrorTextContract(t *testing.T) {
	tests := []struct {
		name string
		run  func() error
		want string
	}{
		{
			name: "arguments must be object",
			run: func() error {
				r := New(Inputs{})
				_, err := r.MaterializeArguments("node", `[]`, []workflow.BindingSpec{{From: "state.value", To: "args.value"}})
				return err
			},
			want: `workflow: node "node" arguments must be a JSON object for input bindings`,
		},
		{
			name: "missing required input",
			run: func() error {
				r := New(Inputs{})
				_, err := r.MaterializeArguments("node", `{}`, []workflow.BindingSpec{{From: "state.value", To: "args.value"}})
				return err
			},
			want: `workflow: node "node" input binding "state.value" -> "args.value": value "value" not found`,
		},
		{
			name: "scalar node field",
			run: func() error {
				r := New(Inputs{})
				if err := r.ApplyNodeOutputs("prepare", nil, NewNodeResult("completed", `{}`, "")); err != nil {
					return err
				}
				_, err := r.MaterializeArguments("submit", `{}`, []workflow.BindingSpec{{From: "node.prepare.status.code", To: "args.value"}})
				return err
			},
			want: `workflow: node "submit" input binding "node.prepare.status.code" -> "args.value": node "prepare" status is scalar and cannot be dereferenced`,
		},
		{
			name: "output target empty segment",
			run: func() error {
				r := New(Inputs{})
				return r.ApplyNodeOutputs("node", []workflow.BindingSpec{{From: "result.summary", To: "state..summary"}}, NewNodeResult("completed", `{"summary":"ready"}`, ""))
			},
			want: `workflow: node "node" output binding "result.summary" -> "state..summary": output binding target "state..summary" contains empty path segment`,
		},
		{
			name: "required slot missing",
			run: func() error {
				return New(Inputs{}).ValidateRequiredSlots([]workflow.StateSlotSpec{{Name: "summary", Required: true}})
			},
			want: `workflow: required state slot "summary" was not populated`,
		},
		{
			name: "slot has state prefix",
			run: func() error {
				return New(Inputs{}).ValidateRequiredSlots([]workflow.StateSlotSpec{{Name: "state.summary", Required: true}})
			},
			want: `workflow: state slot "state.summary" must not include "state." prefix; state_schema names are already rooted at workflow state`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			if err == nil || err.Error() != test.want {
				t.Fatalf("error mismatch:\n got: %v\nwant: %s", err, test.want)
			}
		})
	}
}

func TestRuntimeBindingRawJSONAndPathContract(t *testing.T) {
	runtime := New(Inputs{})
	for _, raw := range []string{" ", ` {"input":"ready"} `} {
		if _, err := runtime.MaterializeArguments("node", raw, []workflow.BindingSpec{{From: "case.input.value", To: "args.value", Optional: true}}); err == nil {
			t.Fatalf("expected padded arguments %q to fail", raw)
		}
	}
	if err := runtime.ApplyNodeOutputs("prepare", nil, NewNodeResult("completed", ` {"summary":"ready"} `, "")); err != nil {
		t.Fatalf("record padded output: %v", err)
	}
	_, err := runtime.MaterializeArguments("submit", `{}`, []workflow.BindingSpec{{From: "node.prepare.result.summary", To: "args.value"}})
	if err == nil || !strings.Contains(err.Error(), "not structured") {
		t.Fatalf("expected padded output to stay unstructured, got %v", err)
	}
}

func TestApplyNodeOutputsPreservesWriteOrderOnError(t *testing.T) {
	runtime := New(Inputs{})
	err := runtime.ApplyNodeOutputs("prepare", []workflow.BindingSpec{
		{From: "result.summary", To: "state.summary"},
		{From: "result.missing", To: "state.missing"},
	}, NewNodeResult("completed", `{"summary":"ready"}`, ""))
	if err == nil {
		t.Fatal("expected second binding to fail")
	}
	if got := runtime.State()["summary"]; got != "ready" {
		t.Fatalf("first binding write = %#v, want ready", got)
	}
	args, err := runtime.MaterializeArguments("submit", `{}`, []workflow.BindingSpec{
		{From: "node.prepare.result.summary", To: "args.summary"},
	})
	if err != nil || args != `{"summary":"ready"}` {
		t.Fatalf("recorded node result after error: args=%s err=%v", args, err)
	}
}
