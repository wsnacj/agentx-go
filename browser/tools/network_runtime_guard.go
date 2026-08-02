package tools

import "context"

func resolveToolRuntimeNetworkBool(ctx context.Context, pick func(ToolRuntimeNetworkGuard) ToolRuntimeBool, fallback bool) bool {
	guard, ok := ToolRuntimeNetworkGuardFromContext(ctx)
	if !ok {
		return fallback
	}
	override := pick(guard)
	if !override.Set {
		return fallback
	}
	return override.Value
}

func resolveToolRuntimeNetworkStrings(ctx context.Context, pick func(ToolRuntimeNetworkGuard) ToolRuntimeStrings) (ToolRuntimeStrings, bool) {
	guard, ok := ToolRuntimeNetworkGuardFromContext(ctx)
	if !ok {
		return ToolRuntimeStrings{}, false
	}
	override := pick(guard)
	if !override.Set {
		return ToolRuntimeStrings{}, false
	}
	return override, true
}

func resolveToolRuntimeNetworkInts(ctx context.Context, pick func(ToolRuntimeNetworkGuard) ToolRuntimeInts) (ToolRuntimeInts, bool) {
	guard, ok := ToolRuntimeNetworkGuardFromContext(ctx)
	if !ok {
		return ToolRuntimeInts{}, false
	}
	override := pick(guard)
	if !override.Set {
		return ToolRuntimeInts{}, false
	}
	return override, true
}

func resolveToolRuntimeNetworkPolicy(ctx context.Context, base outboundNetworkPolicy, pick func(ToolRuntimeNetworkGuard) ToolRuntimeBool) outboundNetworkPolicy {
	effective := base
	effective.allowPrivate = resolveToolRuntimeNetworkBool(ctx, pick, effective.allowPrivate)
	return effective
}

func applyToolRuntimeCIDROverrides(base outboundNetworkPolicy, allow, deny ToolRuntimeStrings) (outboundNetworkPolicy, error) {
	effective := base
	if allow.Set {
		parsed, err := parseCIDRs(allow.Values)
		if err != nil {
			return outboundNetworkPolicy{}, err
		}
		effective.allowCIDRs = parsed
		effective.allowCIDRsSet = true
	}
	if deny.Set {
		parsed, err := parseCIDRs(deny.Values)
		if err != nil {
			return outboundNetworkPolicy{}, err
		}
		effective.denyCIDRs = parsed
	}
	return effective, nil
}

func applyToolRuntimePortOverrides(base outboundNetworkPolicy, allow, deny ToolRuntimeInts) outboundNetworkPolicy {
	effective := base
	if allow.Set {
		effective.allowPorts = buildPortSet(allow.Values)
		effective.allowPortsSet = true
	}
	if deny.Set {
		effective.denyPorts = buildPortSet(deny.Values)
	}
	return effective
}

func resolveToolRuntimeSharedFetchPolicy(ctx context.Context, base outboundNetworkPolicy) (outboundNetworkPolicy, error) {
	effective := resolveToolRuntimeNetworkPolicy(ctx, base, func(guard ToolRuntimeNetworkGuard) ToolRuntimeBool {
		return guard.WebFetchAllowPrivateHosts
	})
	allowCIDRs, _ := resolveToolRuntimeNetworkStrings(ctx, func(guard ToolRuntimeNetworkGuard) ToolRuntimeStrings {
		return guard.WebFetchAllowCIDRs
	})
	denyCIDRs, _ := resolveToolRuntimeNetworkStrings(ctx, func(guard ToolRuntimeNetworkGuard) ToolRuntimeStrings {
		return guard.WebFetchDenyCIDRs
	})
	var err error
	effective, err = applyToolRuntimeCIDROverrides(effective, allowCIDRs, denyCIDRs)
	if err != nil {
		return outboundNetworkPolicy{}, err
	}
	allowPorts, _ := resolveToolRuntimeNetworkInts(ctx, func(guard ToolRuntimeNetworkGuard) ToolRuntimeInts {
		return guard.WebFetchAllowPorts
	})
	denyPorts, _ := resolveToolRuntimeNetworkInts(ctx, func(guard ToolRuntimeNetworkGuard) ToolRuntimeInts {
		return guard.WebFetchDenyPorts
	})
	return applyToolRuntimePortOverrides(effective, allowPorts, denyPorts), nil
}
