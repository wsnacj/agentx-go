package llm

import "testing"

func TestModelLimitsNormalize(t *testing.T) {
	got := (ModelLimits{ContextWindowTokens: 128, MaxInputTokens: 256, MaxOutputTokens: -1}).Normalize()
	if got.ContextWindowTokens != 128 || got.MaxInputTokens != 128 || got.MaxOutputTokens != 0 {
		t.Fatalf("unexpected limits: %+v", got)
	}
}
