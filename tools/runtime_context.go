package tools

import (
	"context"
	"sort"
	"strings"
)

type sessionContextKey struct{}
type runtimeNetworkGuardContextKey struct{}

// OptionalBool distinguishes an explicit false value from no override.
type OptionalBool struct {
	Set   bool
	Value bool
}

// OptionalStrings distinguishes an explicitly empty list from no override.
type OptionalStrings struct {
	Set    bool
	Values []string
}

// OptionalInts distinguishes an explicitly empty list from no override.
type OptionalInts struct {
	Set    bool
	Values []int
}

// RuntimeNetworkGuard carries host-selected per-run network-policy overrides.
// It does not decide defaults and it does not perform network I/O.
type RuntimeNetworkGuard struct {
	WebSearchAllowPrivateHosts    OptionalBool
	WebSearchTrustedEnvProxy      OptionalBool
	WebSearchAllowCIDRs           OptionalStrings
	WebSearchDenyCIDRs            OptionalStrings
	WebSearchAllowPorts           OptionalInts
	WebSearchDenyPorts            OptionalInts
	WebFetchAllowPrivateHosts     OptionalBool
	WebFetchTrustedEnvProxy       OptionalBool
	WebFetchAllowCIDRs            OptionalStrings
	WebFetchDenyCIDRs             OptionalStrings
	WebFetchAllowPorts            OptionalInts
	WebFetchDenyPorts             OptionalInts
	HTTPRequestAllowPrivateHosts  OptionalBool
	HTTPRequestTrustedEnvProxy    OptionalBool
	HTTPRequestAllowCIDRs         OptionalStrings
	HTTPRequestDenyCIDRs          OptionalStrings
	HTTPRequestAllowPorts         OptionalInts
	HTTPRequestDenyPorts          OptionalInts
	BrowserProxyAllowPrivateHosts OptionalBool
	BrowserProxyTrustedEnvProxy   OptionalBool
	BrowserProxyAllowCIDRs        OptionalStrings
	BrowserProxyDenyCIDRs         OptionalStrings
	BrowserProxyAllowPorts        OptionalInts
	BrowserProxyDenyPorts         OptionalInts
	NodesGatewayAllowPrivateHosts OptionalBool
	NodesGatewayTrustedEnvProxy   OptionalBool
	NodesGatewayAllowCIDRs        OptionalStrings
	NodesGatewayDenyCIDRs         OptionalStrings
	NodesGatewayAllowPorts        OptionalInts
	NodesGatewayDenyPorts         OptionalInts
}

// WithToolSessionID returns a context carrying the normalized session
// identity. A nil context or blank identity is preserved unchanged.
func WithToolSessionID(ctx context.Context, sessionID string) context.Context {
	sessionID = strings.TrimSpace(sessionID)
	if ctx == nil || sessionID == "" {
		return ctx
	}
	return context.WithValue(ctx, sessionContextKey{}, sessionID)
}

// ToolSessionIDFromContext returns the normalized session identity, if any.
func ToolSessionIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	sessionID, _ := ctx.Value(sessionContextKey{}).(string)
	return strings.TrimSpace(sessionID)
}

// WithToolRuntimeNetworkGuard returns a context carrying normalized,
// host-selected network overrides. An empty guard is a no-op.
func WithToolRuntimeNetworkGuard(ctx context.Context, guard RuntimeNetworkGuard) context.Context {
	guard = normalizeRuntimeNetworkGuard(guard)
	if ctx == nil {
		return nil
	}
	if runtimeNetworkGuardEmpty(guard) {
		return ctx
	}
	return context.WithValue(ctx, runtimeNetworkGuardContextKey{}, guard)
}

// ToolRuntimeNetworkGuardFromContext returns normalized network overrides.
func ToolRuntimeNetworkGuardFromContext(ctx context.Context) (RuntimeNetworkGuard, bool) {
	if ctx == nil {
		return RuntimeNetworkGuard{}, false
	}
	guard, ok := ctx.Value(runtimeNetworkGuardContextKey{}).(RuntimeNetworkGuard)
	if !ok {
		return RuntimeNetworkGuard{}, false
	}
	guard = normalizeRuntimeNetworkGuard(guard)
	if runtimeNetworkGuardEmpty(guard) {
		return RuntimeNetworkGuard{}, false
	}
	return guard, true
}

func normalizeRuntimeNetworkGuard(in RuntimeNetworkGuard) RuntimeNetworkGuard {
	in.WebSearchAllowCIDRs = normalizeOptionalStrings(in.WebSearchAllowCIDRs)
	in.WebSearchDenyCIDRs = normalizeOptionalStrings(in.WebSearchDenyCIDRs)
	in.WebSearchAllowPorts = normalizeOptionalInts(in.WebSearchAllowPorts)
	in.WebSearchDenyPorts = normalizeOptionalInts(in.WebSearchDenyPorts)
	in.WebFetchAllowCIDRs = normalizeOptionalStrings(in.WebFetchAllowCIDRs)
	in.WebFetchDenyCIDRs = normalizeOptionalStrings(in.WebFetchDenyCIDRs)
	in.WebFetchAllowPorts = normalizeOptionalInts(in.WebFetchAllowPorts)
	in.WebFetchDenyPorts = normalizeOptionalInts(in.WebFetchDenyPorts)
	in.HTTPRequestAllowCIDRs = normalizeOptionalStrings(in.HTTPRequestAllowCIDRs)
	in.HTTPRequestDenyCIDRs = normalizeOptionalStrings(in.HTTPRequestDenyCIDRs)
	in.HTTPRequestAllowPorts = normalizeOptionalInts(in.HTTPRequestAllowPorts)
	in.HTTPRequestDenyPorts = normalizeOptionalInts(in.HTTPRequestDenyPorts)
	in.BrowserProxyAllowCIDRs = normalizeOptionalStrings(in.BrowserProxyAllowCIDRs)
	in.BrowserProxyDenyCIDRs = normalizeOptionalStrings(in.BrowserProxyDenyCIDRs)
	in.BrowserProxyAllowPorts = normalizeOptionalInts(in.BrowserProxyAllowPorts)
	in.BrowserProxyDenyPorts = normalizeOptionalInts(in.BrowserProxyDenyPorts)
	in.NodesGatewayAllowCIDRs = normalizeOptionalStrings(in.NodesGatewayAllowCIDRs)
	in.NodesGatewayDenyCIDRs = normalizeOptionalStrings(in.NodesGatewayDenyCIDRs)
	in.NodesGatewayAllowPorts = normalizeOptionalInts(in.NodesGatewayAllowPorts)
	in.NodesGatewayDenyPorts = normalizeOptionalInts(in.NodesGatewayDenyPorts)
	return in
}

func runtimeNetworkGuardEmpty(in RuntimeNetworkGuard) bool {
	return !in.WebSearchAllowPrivateHosts.Set &&
		!in.WebSearchTrustedEnvProxy.Set &&
		!in.WebSearchAllowCIDRs.Set &&
		!in.WebSearchDenyCIDRs.Set &&
		!in.WebSearchAllowPorts.Set &&
		!in.WebSearchDenyPorts.Set &&
		!in.WebFetchAllowPrivateHosts.Set &&
		!in.WebFetchTrustedEnvProxy.Set &&
		!in.WebFetchAllowCIDRs.Set &&
		!in.WebFetchDenyCIDRs.Set &&
		!in.WebFetchAllowPorts.Set &&
		!in.WebFetchDenyPorts.Set &&
		!in.HTTPRequestAllowPrivateHosts.Set &&
		!in.HTTPRequestTrustedEnvProxy.Set &&
		!in.HTTPRequestAllowCIDRs.Set &&
		!in.HTTPRequestDenyCIDRs.Set &&
		!in.HTTPRequestAllowPorts.Set &&
		!in.HTTPRequestDenyPorts.Set &&
		!in.BrowserProxyAllowPrivateHosts.Set &&
		!in.BrowserProxyTrustedEnvProxy.Set &&
		!in.BrowserProxyAllowCIDRs.Set &&
		!in.BrowserProxyDenyCIDRs.Set &&
		!in.BrowserProxyAllowPorts.Set &&
		!in.BrowserProxyDenyPorts.Set &&
		!in.NodesGatewayAllowPrivateHosts.Set &&
		!in.NodesGatewayTrustedEnvProxy.Set &&
		!in.NodesGatewayAllowCIDRs.Set &&
		!in.NodesGatewayDenyCIDRs.Set &&
		!in.NodesGatewayAllowPorts.Set &&
		!in.NodesGatewayDenyPorts.Set
}

func normalizeOptionalStrings(in OptionalStrings) OptionalStrings {
	if !in.Set {
		return OptionalStrings{}
	}
	return OptionalStrings{Set: true, Values: normalizeRuntimeNetworkStringValues(in.Values)}
}

func normalizeOptionalInts(in OptionalInts) OptionalInts {
	if !in.Set {
		return OptionalInts{}
	}
	return OptionalInts{Set: true, Values: normalizeRuntimeNetworkIntValues(in.Values)}
}

func normalizeRuntimeNetworkStringValues(values []string) []string {
	if values == nil {
		return nil
	}
	if len(values) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		out = append(out, normalized)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return []string{}
	}
	return out
}

func normalizeRuntimeNetworkIntValues(values []int) []int {
	if values == nil {
		return nil
	}
	if len(values) == 0 {
		return []int{}
	}
	seen := map[int]bool{}
	out := make([]int, 0, len(values))
	for _, value := range values {
		if value <= 0 || value > 65535 || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Ints(out)
	if len(out) == 0 {
		return []int{}
	}
	return out
}
