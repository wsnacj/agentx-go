package tool_test

import (
	"context"
	"testing"

	"github.com/wsnacj/agentx-go/components/tool"
)

func TestContractsRemainWireCompatible(t *testing.T) {
	definition := tool.Definition{Type: "function", Function: tool.Function{Name: "echo"}}
	handler := tool.Handler(func(_ context.Context, call tool.Call) (tool.Result, error) {
		return call.Arguments, nil
	})
	result, err := handler(context.Background(), tool.Call{Name: definition.Function.Name, Arguments: "ready"})
	if err != nil || result != "ready" {
		t.Fatalf("result=%q err=%v", result, err)
	}
}
