package contextwindow

import (
	"context"
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestPrepareUsesTokenLimitAndSemanticCompaction(t *testing.T) {
	counter := llm.TokenCounterFunc(func(_ context.Context, request llm.TokenCountRequest) (llm.TokenCount, error) {
		var tokens int64
		for _, message := range request.Messages {
			tokens += int64(len(strings.Fields(message.Content)))
		}
		return llm.TokenCount{Tokens: tokens, Exact: true, Source: "fixture-tokenizer"}, nil
	})
	orchestrator := Orchestrator{
		Policy:       Policy{MaxChars: 10_000, ProtectedTailSegments: 1, SummaryTargetChars: 30, MaxInputTokens: 7},
		TokenCounter: counter,
		Summarizer: SummarizerFunc(func(context.Context, SummaryRequest) (Summary, error) {
			return Summary{Content: "old facts"}, nil
		}),
	}
	result, err := orchestrator.Prepare(context.Background(), Request{Messages: llm.Conversation{
		{Role: "user", Content: "one two three"},
		{Role: "assistant", Content: "four five six"},
		{Role: "user", Content: "latest question"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Report.SemanticSummaryUsed || result.Report.AfterInputTokens > 7 || !result.Report.InputTokenCountExact {
		t.Fatalf("unexpected result: %+v", result.Report)
	}
}

func TestPrepareRequiresCounterForTokenPolicy(t *testing.T) {
	_, err := (Orchestrator{Policy: Policy{MaxChars: 100, SummaryTargetChars: 20, ProtectedTailSegments: 1, MaxInputTokens: 10}}).
		Prepare(context.Background(), Request{})
	typed, ok := AsError(err)
	if !ok || typed.Code != ErrorCodeInvalidPolicy {
		t.Fatalf("unexpected error: %v", err)
	}
}
