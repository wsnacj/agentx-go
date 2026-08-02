package main

import (
	"context"
	"testing"
)

func TestFixedVersionProviderCohort(t *testing.T) {
	value, err := run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if value.AnthropicContent != "anthropic-ready" || value.AnthropicTokens != 5 || value.CodexContent != "codex-ready" || value.CodexTool != "lookup" || value.CodexTokens != 6 || value.Authorization != "Bearer codex-fixture" || value.AccountID != "acct-fixture" {
		t.Fatalf("result = %#v", value)
	}
}
