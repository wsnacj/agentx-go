package tools

import (
	"context"

	agentxtools "github.com/wsnacj/agentx-go/tools"
)

// ToolMetadata aliases the provider-neutral catalog metadata contract owned by
// the canonical tools module.
type ToolMetadata = agentxtools.ToolMetadata

const ToolSourceBuiltin = agentxtools.ToolSourceBuiltin

type ToolRuntimeBool = agentxtools.OptionalBool
type ToolRuntimeStrings = agentxtools.OptionalStrings
type ToolRuntimeInts = agentxtools.OptionalInts
type ToolRuntimeNetworkGuard = agentxtools.RuntimeNetworkGuard

func NormalizeToolName(name string) string {
	return agentxtools.NormalizeToolName(name)
}

func ToolSessionIDFromContext(ctx context.Context) string {
	return agentxtools.ToolSessionIDFromContext(ctx)
}

func WithToolSessionID(ctx context.Context, sessionID string) context.Context {
	return agentxtools.WithToolSessionID(ctx, sessionID)
}

func WithToolRuntimeNetworkGuard(ctx context.Context, guard ToolRuntimeNetworkGuard) context.Context {
	return agentxtools.WithToolRuntimeNetworkGuard(ctx, guard)
}

func ToolRuntimeNetworkGuardFromContext(ctx context.Context) (ToolRuntimeNetworkGuard, bool) {
	return agentxtools.ToolRuntimeNetworkGuardFromContext(ctx)
}

func builtinToolMetadataBoolPtr(value bool) *bool {
	cloned := value
	return &cloned
}

func isBrowserBuiltinToolName(name string) bool {
	switch NormalizeToolName(name) {
	case "browser", "browser_runtime", "browser_act":
		return true
	default:
		return false
	}
}
