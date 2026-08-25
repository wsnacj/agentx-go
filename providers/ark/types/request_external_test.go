package types_test

import (
	"encoding/json"
	"testing"

	arktypes "github.com/wsnacj/agentx-go/providers/ark/types"
)

func TestResponseRequestParallelToolCallsWireContract(t *testing.T) {
	parallel := false
	request := arktypes.ResponseRequest{
		Model:             "doubao-fixture",
		Input:             arktypes.NewInputText("probe"),
		ParallelToolCalls: &parallel,
	}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	value, ok := payload["parallel_tool_calls"]
	if !ok || value != false {
		t.Fatalf("parallel_tool_calls = %#v, present=%t", value, ok)
	}
	if _, ok := payload["max_tool_calls"]; ok {
		t.Fatalf("unsupported max_tool_calls leaked into payload: %s", data)
	}

	request.ParallelToolCalls = nil
	data, err = json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	payload = make(map[string]any)
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if _, ok := payload["parallel_tool_calls"]; ok {
		t.Fatalf("nil parallel_tool_calls should be omitted: %s", data)
	}
}
