package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/wsnacj/agentx-go/runtime/memory"
)

func TestExternalContractUsesTypedDisplaySafeErrors(t *testing.T) {
	coordinator := memory.Coordinator{
		Policy: memory.Policy{MaxRecallLimit: 4, MaxContentBytes: 1024, MaxReferenceCount: 8},
		Backend: memory.BackendFuncs{RecallFunc: func(context.Context, memory.RecallRequest) (memory.RecallResult, error) {
			return memory.RecallResult{}, errors.New("private backend details")
		}},
	}
	_, err := coordinator.Recall(context.Background(), memory.RecallRequest{ScopeRef: "user:1", Query: "preference"})
	var typed *memory.Error
	if !errors.As(err, &typed) || typed.Code != memory.ErrorCodeRecallFailed {
		t.Fatalf("error = %#v", err)
	}
	if err.Error() != "memory recall failed" {
		t.Fatalf("display error = %q", err.Error())
	}
}
