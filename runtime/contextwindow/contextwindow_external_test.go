package contextwindow_test

import (
	"context"
	"errors"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
	"github.com/wsnacj/agentx-go/runtime/contextwindow"
)

func TestExternalContract(t *testing.T) {
	orchestrator := contextwindow.Orchestrator{
		Policy: contextwindow.Policy{
			MaxChars: 64, ProtectedTailSegments: 1, SummaryTargetChars: 24,
		},
	}
	_, err := orchestrator.Prepare(context.Background(), contextwindow.Request{
		Messages: llm.Conversation{{Role: "user", Content: "a message that cannot fit in the deliberately tiny context window without a host summarizer"}},
	})
	var typed *contextwindow.Error
	if !errors.As(err, &typed) {
		t.Fatalf("error type = %T", err)
	}
	if typed.Code != contextwindow.ErrorCodeSummarizerUnavailable && typed.Code != contextwindow.ErrorCodeLimitUnresolved {
		t.Fatalf("error = %#v", typed)
	}
}
