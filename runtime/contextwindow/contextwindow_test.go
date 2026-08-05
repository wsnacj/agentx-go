package contextwindow

import (
	"context"
	"errors"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestPrepareSemanticCompactionPreservesProtocolAndLatestUser(t *testing.T) {
	messages := llm.Conversation{
		{Role: "system", Content: "base"},
		{Role: "user", Content: "first request"},
		{Role: "assistant", Content: strings.Repeat("old answer ", 20)},
		{Role: "user", Content: "latest request"},
		{Role: "assistant", ToolCalls: []llm.FunctionCall{{ID: "call-1", Name: "lookup", Arguments: `{}`}}},
		{Role: "tool", ToolCallID: "call-1", Content: strings.Repeat("tool result ", 30)},
		{Role: "assistant", Content: "working"},
	}
	var summarized llm.Conversation
	orchestrator := Orchestrator{
		Policy: Policy{
			MaxChars: 300, ProtectedHeadSegments: 1, ProtectedTailSegments: 1,
			SummaryTargetChars: 80, StrictToolProtocol: true,
		},
		Summarizer: SummarizerFunc(func(_ context.Context, request SummaryRequest) (Summary, error) {
			summarized = cloneConversation(request.Messages)
			if request.PreviousSummary != "previous" || request.TargetChars != 80 {
				t.Fatalf("summary request = %#v", request)
			}
			return Summary{Content: "old work and tool evidence"}, nil
		}),
	}
	result, err := orchestrator.Prepare(context.Background(), Request{
		Messages: messages, PreviousSummary: "previous",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Report.SemanticSummaryUsed || result.Summary != "old work and tool evidence" {
		t.Fatalf("result = %#v", result)
	}
	if len(summarized) == 0 {
		t.Fatal("summarizer did not receive a middle window")
	}
	if !containsMessage(result.Messages, "user", "latest request") {
		t.Fatalf("latest user message missing: %#v", result.Messages)
	}
	if !containsToolProtocolPair(result.Messages, "call-1") {
		t.Fatalf("tool protocol pair was split: %#v", result.Messages)
	}
	if messages[2].Content != strings.Repeat("old answer ", 20) {
		t.Fatal("Prepare mutated caller conversation")
	}
}

func TestPrepareSummaryFailureReturnsOriginal(t *testing.T) {
	messages := llm.Conversation{
		{Role: "user", Content: strings.Repeat("old ", 80)},
		{Role: "assistant", Content: strings.Repeat("answer ", 80)},
		{Role: "user", Content: "latest"},
	}
	cause := errors.New("provider detail that must not be displayed")
	orchestrator := Orchestrator{
		Policy: Policy{MaxChars: 100, ProtectedTailSegments: 1, SummaryTargetChars: 40},
		Summarizer: SummarizerFunc(func(context.Context, SummaryRequest) (Summary, error) {
			return Summary{}, cause
		}),
	}
	result, err := orchestrator.Prepare(context.Background(), Request{Messages: messages})
	typed, ok := AsError(err)
	if !ok || typed.Code != ErrorCodeSummarizationFailed || !errors.Is(err, cause) {
		t.Fatalf("error = %#v", err)
	}
	if strings.Contains(err.Error(), "provider detail") {
		t.Fatalf("display error leaked cause: %q", err)
	}
	if !equalConversation(result.Messages, messages) {
		t.Fatalf("failure result = %#v, want original %#v", result.Messages, messages)
	}
}

func TestPrepareCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	orchestrator := Orchestrator{Policy: Policy{MaxChars: 10, ProtectedTailSegments: 1, SummaryTargetChars: 5}}
	_, err := orchestrator.Prepare(ctx, Request{Messages: llm.Conversation{{Role: "user", Content: "hello"}}})
	typed, ok := AsError(err)
	if !ok || typed.Code != ErrorCodeCanceled || !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %#v", err)
	}
}

func TestPrepareRejectsInvalidPolicy(t *testing.T) {
	_, err := (Orchestrator{}).Prepare(context.Background(), Request{})
	typed, ok := AsError(err)
	if !ok || typed.Code != ErrorCodeInvalidPolicy {
		t.Fatalf("error = %#v", err)
	}
}

func containsMessage(messages llm.Conversation, role, content string) bool {
	for _, message := range messages {
		if message.Role == role && message.Content == content {
			return true
		}
	}
	return false
}

func containsToolProtocolPair(messages llm.Conversation, id string) bool {
	call := -1
	result := -1
	for index, message := range messages {
		for _, toolCall := range message.ToolCalls {
			if toolCall.ID == id {
				call = index
			}
		}
		if message.ToolCallID == id {
			result = index
		}
	}
	return call >= 0 && result == call+1
}

func equalConversation(a, b llm.Conversation) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index].Role != b[index].Role || a[index].Content != b[index].Content || a[index].ToolCallID != b[index].ToolCallID || len(a[index].ToolCalls) != len(b[index].ToolCalls) {
			return false
		}
	}
	return true
}
