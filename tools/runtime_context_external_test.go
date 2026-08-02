package tools_test

import (
	"context"
	"testing"

	tools "github.com/wsnacj/agentx-go/tools"
)

func TestToolRuntimeContextPreservesSessionAndNormalizesNetworkGuard(t *testing.T) {
	ctx := tools.WithToolSessionID(context.Background(), "  session-123  ")
	ctx = tools.WithToolRuntimeNetworkGuard(ctx, tools.RuntimeNetworkGuard{
		WebFetchAllowPrivateHosts: tools.OptionalBool{Set: true, Value: true},
		WebFetchAllowCIDRs: tools.OptionalStrings{
			Set:    true,
			Values: []string{" 127.0.0.0/8 ", "127.0.0.0/8"},
		},
	})
	if got := tools.ToolSessionIDFromContext(ctx); got != "session-123" {
		t.Fatalf("session id = %q", got)
	}
	guard, ok := tools.ToolRuntimeNetworkGuardFromContext(ctx)
	if !ok || !guard.WebFetchAllowPrivateHosts.Value {
		t.Fatalf("network guard = %#v, %v", guard, ok)
	}
	if got := guard.WebFetchAllowCIDRs.Values; len(got) != 1 || got[0] != "127.0.0.0/8" {
		t.Fatalf("allow cidrs = %#v", got)
	}
}
