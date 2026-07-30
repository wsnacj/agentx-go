package bindingstate_test

import (
	"testing"

	workflow "github.com/wsnacj/agentx-go/runtime/workflow"
	bindingstate "github.com/wsnacj/agentx-go/runtime/workflow/bindingstate"
)

func TestExternalPackageConsumer(t *testing.T) {
	runtime := bindingstate.New(bindingstate.Inputs{
		InitialState: map[string]any{"request": "review"},
	})
	args, err := runtime.MaterializeArguments("collect", `{}`, []workflow.BindingSpec{
		{From: "state.request", To: "args.query"},
	})
	if err != nil {
		t.Fatalf("MaterializeArguments(): %v", err)
	}
	if args != `{"query":"review"}` {
		t.Fatalf("arguments = %s", args)
	}
	if err := runtime.ApplyNodeOutputs(
		"collect",
		[]workflow.BindingSpec{{From: "result.report", To: "state.report"}},
		bindingstate.NewNodeResult("completed", `{"report":"ready"}`, ""),
	); err != nil {
		t.Fatalf("ApplyNodeOutputs(): %v", err)
	}
	if err := runtime.ValidateRequiredSlots([]workflow.StateSlotSpec{
		{Name: "report", Required: true},
	}); err != nil {
		t.Fatalf("ValidateRequiredSlots(): %v", err)
	}
	if got := runtime.State()["report"]; got != "ready" {
		t.Fatalf("report = %#v, want ready", got)
	}
}
