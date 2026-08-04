package transcript

import (
	"strings"
	"testing"

	llm "github.com/wsnacj/agentx-go/components/llm"
)

func TestEstimateAndEvaluate(t *testing.T) {
	messages := llm.Conversation{{Role: "assistant", Content: "hi", ToolCalls: []llm.FunctionCall{{Name: "lookup", Arguments: `{"q":"x"}`}}}}
	input := EstimateInput{SystemPrompt: "sys", Messages: messages, RoleAware: true}
	want := len("sys") + len("hi") + len("lookup") + len(`{"q":"x"}`)
	if got := EstimateChars(input); got != want {
		t.Fatalf("EstimateChars() = %d, want %d", got, want)
	}
	decision := Evaluate(input, GuardPolicy{WarnChars: want, MaxChars: want})
	if !decision.Warn || !decision.Overflow || decision.EstimatedChars != want {
		t.Fatalf("Evaluate() = %#v", decision)
	}
	if got := OverflowMessage(want, want); got != "context limit exceeded for this model (estimated 20 chars, limit 20)" {
		t.Fatalf("OverflowMessage() = %q", got)
	}
}

func TestCompactPreservesHostAnchorWithoutOwningItsPolicy(t *testing.T) {
	evidence := `{"payload":"` + strings.Repeat("x", 600) + `","citations":["p1"]}`
	messages := llm.Conversation{
		{Role: "tool", Content: evidence},
		{Role: "user", Content: strings.Repeat("q", 300)},
	}
	got, diagnostic := Compact(messages, CompactionPolicy{
		MaxChars: 700,
		ToolOutputAnchor: func(content string) int {
			return strings.LastIndex(content, `"citations"`)
		},
	})
	if diagnostic.CompactedToolOutputs == 0 {
		t.Fatalf("expected tool compaction, got %#v", diagnostic)
	}
	if !strings.Contains(got[0].Content, `"citations"`) {
		t.Fatalf("expected host-selected suffix to survive, got %q", got[0].Content)
	}
	if messages[0].Content != evidence {
		t.Fatal("Compact mutated caller conversation")
	}
}

func TestSanitizeStrictProtocolAndReasoning(t *testing.T) {
	got, diagnostic := Sanitize(llm.Conversation{
		{Role: "assistant", Content: "<thinking>private</thinking>visible", ToolCalls: []llm.FunctionCall{{Name: "lookup", Arguments: `{}`}}},
		{Role: "tool", Content: "result"},
	}, SanitizePolicy{StrictToolProtocol: true, StripInternalReasoning: true})
	if len(got) != 2 || got[0].ToolCalls[0].ID != "agentx_call_1" || got[1].ToolCallID != "agentx_call_1" {
		t.Fatalf("Sanitize() = %#v", got)
	}
	if strings.Contains(got[0].Content, "private") || !strings.Contains(got[0].Content, "visible") {
		t.Fatalf("reasoning projection = %q", got[0].Content)
	}
	if diagnostic.SynthesizedToolCallIDs != 1 || diagnostic.RecoveredToolResults != 1 || diagnostic.StrippedReasoningMsgs != 1 {
		t.Fatalf("diagnostic = %#v", diagnostic)
	}
}

func TestPruneKeepsNewestWholeProtocolSegments(t *testing.T) {
	messages := llm.Conversation{
		{Role: "system", Content: "summary"},
		{Role: "assistant", ToolCalls: []llm.FunctionCall{{ID: "call_1", Name: "lookup"}}},
		{Role: "tool", ToolCallID: "call_1", Content: "result"},
		{Role: "assistant", Content: "recent"},
	}
	got, drops := PruneTailPreservingSystemPrefix(messages, HistoryPolicy{MaxEvents: 2, StrictToolProtocol: true})
	if len(got) != 2 || got[0].Role != "system" || got[1].Content != "recent" || drops != 2 {
		t.Fatalf("PruneTailPreservingSystemPrefix() = %#v, drops=%d", got, drops)
	}
}
