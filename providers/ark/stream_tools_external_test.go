package ark_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/wsnacj/agentx-go/providers/ark"
	arktypes "github.com/wsnacj/agentx-go/providers/ark/types"
)

func TestCollectFunctionCallsFromStream(t *testing.T) {
	events := make(chan arktypes.Event, 3)
	errs := make(chan error)
	responseRaw := json.RawMessage(`{"type":"response.created","response":{"id":"resp_1","model":"doubao"}}`)
	itemRaw := json.RawMessage(`{"type":"response.output_item.added","item":{"id":"item_1","type":"function_call","name":"lookup","call_id":"call_1"}}`)
	argsRaw := json.RawMessage(`{"type":"response.function_call_arguments.done","item_id":"item_1","arguments":"{\"q\":\"agentx\"}"}`)
	events <- arktypes.Event{Type: arktypes.EventTypeResponseCreated, Raw: responseRaw}
	events <- arktypes.Event{Type: arktypes.EventTypeResponseOutputItemAdded, Raw: itemRaw}
	events <- arktypes.Event{Type: arktypes.EventTypeResponseFunctionCallArgsDone, Raw: argsRaw}
	close(events)
	close(errs)

	result, err := ark.CollectFunctionCallsFromStream(context.Background(), &arktypes.StreamResult{Ch: events, Err: errs, Cancel: func() {}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ResponseID != "resp_1" || result.Model != "doubao" || len(result.Calls) != 1 {
		t.Fatalf("result = %#v", result)
	}
	if call := result.Calls[0]; call.Name != "lookup" || call.CallID != "call_1" || call.Arguments != `{"q":"agentx"}` {
		t.Fatalf("call = %#v", call)
	}
}

func TestBuildToolOutputRequestContract(t *testing.T) {
	request, err := ark.BuildToolOutputRequest(
		&arktypes.Response{ID: "resp_1", Model: "doubao"},
		[]ark.FunctionResult{{Call: ark.FunctionCall{CallID: "call_1"}, Output: "ok"}},
		ark.ToolFollowupOptions{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if request.PreviousResponseID == nil || *request.PreviousResponseID != "resp_1" || request.Model != "doubao" || len(request.Input.Items) != 1 {
		t.Fatalf("request = %#v", request)
	}
	if item := request.Input.Items[0]; item.Type != "function_call_output" || item.CallID != "call_1" || item.Output != "ok" {
		t.Fatalf("item = %#v", item)
	}
}
