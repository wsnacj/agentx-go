package tools

import (
	"strconv"
	"strings"
)

type BrowserToolMetadataRouteHints struct {
	RuntimeInfo              BrowserRuntimeInfo
	Source                   string
	Endpoint                 string
	DefaultCandidateRoute    BrowserRuntimeInfo
	DefaultCandidateSource   string
	DefaultCandidateEndpoint string
	Surface                  string
	OptInTargets             []string
	HiddenDefaultCandidate   bool
	SubstratePosture         string
	SubstrateStatus          string
	SelectionStrategy        string
	SubstrateReason          string
	RepairAvailable          bool
}

type BrowserToolMetadataRouteHintActions struct {
	DoctorAction string
	RepairAction string
	ReadyAction  string
}

func (h BrowserToolMetadataRouteHints) PreferredRuntimeInfo() BrowserRuntimeInfo {
	info, _, _ := h.preferredRuntimeRoute()
	return info
}

func (h BrowserToolMetadataRouteHints) preferredRuntimeRoute() (BrowserRuntimeInfo, string, string) {
	runtimeInfo := normalizeBrowserRuntimeInfo(h.RuntimeInfo)
	runtimeSource := strings.TrimSpace(h.Source)
	runtimeEndpoint := strings.TrimSpace(h.Endpoint)
	defaultCandidate := normalizeBrowserRuntimeInfo(h.DefaultCandidateRoute)
	defaultCandidateSource := strings.TrimSpace(h.DefaultCandidateSource)
	defaultCandidateEndpoint := strings.TrimSpace(h.DefaultCandidateEndpoint)
	if defaultCandidate == (BrowserRuntimeInfo{}) {
		return runtimeInfo, runtimeSource, runtimeEndpoint
	}
	if runtimeInfo == (BrowserRuntimeInfo{}) {
		return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
	}
	if !h.HiddenDefaultCandidate ||
		BrowserSubstratePosture(runtimeInfo.Backend, runtimeInfo.Target) != BrowserSubstrateLegacySystemHost {
		return runtimeInfo, runtimeSource, runtimeEndpoint
	}
	switch strings.TrimSpace(h.SelectionStrategy) {
	case BrowserSubstrateSelectionPreferNodeOverLegacy, BrowserSubstrateSelectionPreferNodeRuntime:
		if strings.EqualFold(strings.TrimSpace(defaultCandidate.Target), "node") {
			return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
		}
	case BrowserSubstrateSelectionPreferSandboxRuntime:
		if strings.EqualFold(strings.TrimSpace(defaultCandidate.Target), "sandbox") {
			return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
		}
	case BrowserSubstrateSelectionPreferCustomBackend:
		return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
	}
	switch strings.TrimSpace(h.SubstratePosture) {
	case BrowserSubstrateNodeRuntime:
		if strings.EqualFold(strings.TrimSpace(defaultCandidate.Target), "node") {
			return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
		}
	case BrowserSubstrateSandboxRuntime:
		if strings.EqualFold(strings.TrimSpace(defaultCandidate.Target), "sandbox") {
			return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
		}
	case BrowserSubstrateCustomBackend:
		return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
	}
	return runtimeInfo, runtimeSource, runtimeEndpoint
}

func (h BrowserToolMetadataRouteHints) EffectiveDefaultCandidateRoute() BrowserRuntimeInfo {
	info, _, _ := h.effectiveDefaultCandidateRoute()
	return info
}

func (h BrowserToolMetadataRouteHints) effectiveDefaultCandidateRoute() (BrowserRuntimeInfo, string, string) {
	defaultCandidate := normalizeBrowserRuntimeInfo(h.DefaultCandidateRoute)
	defaultCandidateSource := strings.TrimSpace(h.DefaultCandidateSource)
	defaultCandidateEndpoint := strings.TrimSpace(h.DefaultCandidateEndpoint)
	if defaultCandidate != (BrowserRuntimeInfo{}) {
		return defaultCandidate, defaultCandidateSource, defaultCandidateEndpoint
	}
	if !h.HiddenDefaultCandidate {
		return BrowserRuntimeInfo{}, "", ""
	}
	runtimeInfo := normalizeBrowserRuntimeInfo(h.RuntimeInfo)
	runtimeSource := strings.TrimSpace(h.Source)
	runtimeEndpoint := strings.TrimSpace(h.Endpoint)
	if runtimeInfo == (BrowserRuntimeInfo{}) {
		return BrowserRuntimeInfo{}, "", ""
	}
	if BrowserSubstratePosture(runtimeInfo.Backend, runtimeInfo.Target) == BrowserSubstrateLegacySystemHost {
		return BrowserRuntimeInfo{}, "", ""
	}
	return runtimeInfo, runtimeSource, runtimeEndpoint
}

func (h BrowserToolMetadataRouteHints) WithPreferredRuntimeInfo() BrowserToolMetadataRouteHints {
	h.RuntimeInfo, h.Source, h.Endpoint = h.preferredRuntimeRoute()
	return h
}

func (h BrowserToolMetadataRouteHints) Canonicalized() BrowserToolMetadataRouteHints {
	h.RuntimeInfo = normalizeBrowserRuntimeInfo(h.RuntimeInfo)
	h.Source = strings.TrimSpace(h.Source)
	h.Endpoint = strings.TrimSpace(h.Endpoint)
	h.DefaultCandidateSource = strings.TrimSpace(h.DefaultCandidateSource)
	h.DefaultCandidateEndpoint = strings.TrimSpace(h.DefaultCandidateEndpoint)
	h.DefaultCandidateRoute, h.DefaultCandidateSource, h.DefaultCandidateEndpoint = h.effectiveDefaultCandidateRoute()
	h.RuntimeInfo, h.Source, h.Endpoint = h.preferredRuntimeRoute()
	return h
}

func (h BrowserToolMetadataRouteHints) DetailFields() []string {
	h = h.Canonicalized()
	parts := make([]string, 0, 13)
	if value := strings.TrimSpace(h.Surface); value != "" {
		parts = append(parts, "surface="+value)
	}
	if value := strings.TrimSpace(h.Source); value != "" {
		parts = append(parts, "source="+value)
	}
	if value := strings.TrimSpace(h.Endpoint); value != "" {
		parts = append(parts, "endpoint="+value)
	}
	if value := strings.TrimSpace(h.RuntimeInfo.Backend); value != "" {
		parts = append(parts, "backend="+value)
	}
	if value := strings.TrimSpace(h.RuntimeInfo.Profile); value != "" {
		parts = append(parts, "profile="+value)
	}
	if value := strings.TrimSpace(h.RuntimeInfo.Target); value != "" {
		parts = append(parts, "target="+value)
	}
	if len(h.OptInTargets) > 0 {
		parts = append(parts, "opt_in_targets="+strings.Join(h.OptInTargets, "+"))
	}
	if value := strings.TrimSpace(h.DefaultCandidateLabel()); value != "" {
		parts = append(parts, "default_candidate="+value)
	}
	if value := strings.TrimSpace(h.DefaultCandidateSource); value != "" {
		parts = append(parts, "default_candidate_source="+value)
	}
	if value := strings.TrimSpace(h.DefaultCandidateEndpoint); value != "" {
		parts = append(parts, "default_candidate_endpoint="+value)
	}
	if value := strings.TrimSpace(h.DefaultCandidateRoute.Backend); value != "" {
		parts = append(parts, "default_candidate_backend="+value)
	}
	if value := strings.TrimSpace(h.DefaultCandidateRoute.Profile); value != "" {
		parts = append(parts, "default_candidate_profile="+value)
	}
	if value := strings.TrimSpace(h.DefaultCandidateRoute.Target); value != "" {
		parts = append(parts, "default_candidate_target="+value)
	}
	if value := strings.TrimSpace(h.SubstratePosture); value != "" {
		parts = append(parts, "substrate_posture="+value)
	}
	if value := strings.TrimSpace(h.SubstrateStatus); value != "" {
		parts = append(parts, "substrate_status="+value)
	}
	if value := strings.TrimSpace(h.SelectionStrategy); value != "" {
		parts = append(parts, "selection_strategy="+value)
	}
	if value := strings.TrimSpace(h.SubstrateReason); value != "" {
		parts = append(parts, "substrate_reason="+strconv.Quote(value))
	}
	if h.RepairAvailable {
		parts = append(parts, "repair_available=true")
	}
	return parts
}

func BrowserToolMetadataRouteHintsDetailText(h BrowserToolMetadataRouteHints) string {
	return strings.Join(h.DetailFields(), " ")
}

func BrowserToolMetadataHiddenDefaultCandidateText(
	h BrowserToolMetadataRouteHints,
	actions BrowserToolMetadataRouteHintActions,
) string {
	h = h.Canonicalized()
	actions = BrowserToolMetadataRouteHintActions{
		DoctorAction: strings.TrimSpace(actions.DoctorAction),
		RepairAction: strings.TrimSpace(actions.RepairAction),
		ReadyAction:  strings.TrimSpace(actions.ReadyAction),
	}
	if !h.HiddenDefaultCandidate {
		return ""
	}
	routeInfo := normalizeBrowserRuntimeInfo(h.RuntimeInfo)
	if routeInfo == (BrowserRuntimeInfo{}) {
		routeInfo = h.EffectiveDefaultCandidateRoute()
	}
	routeParts := make([]string, 0, 3)
	if value := strings.TrimSpace(routeInfo.Backend); value != "" {
		routeParts = append(routeParts, value)
	}
	if value := strings.TrimSpace(routeInfo.Profile); value != "" {
		routeParts = append(routeParts, value)
	}
	if value := strings.TrimSpace(routeInfo.Target); value != "" {
		routeParts = append(routeParts, value)
	}
	line := "Browser route note: current tool metadata indicates a hidden managed browserd default candidate"
	if len(routeParts) > 0 {
		line += " on `" + strings.Join(routeParts, "/") + "`"
	}
	if value := strings.TrimSpace(h.SelectionStrategy); value != "" {
		line += " (`selection_strategy=" + value + "`)"
	}
	if value := strings.TrimSpace(h.SubstrateReason); value != "" {
		line += "; current route note: " + value
	}
	actionParts := make([]string, 0, 3)
	if actions.DoctorAction != "" {
		actionParts = append(actionParts, "prefer `"+actions.DoctorAction+"`")
	}
	if h.RepairAvailable && actions.RepairAction != "" {
		actionParts = append(actionParts, "run `"+actions.RepairAction+"` if browserd bootstrap is blocked")
	}
	if actions.ReadyAction != "" {
		actionParts = append(actionParts, "then use `"+actions.ReadyAction+"` before falling back to host-only browser flows")
	}
	if len(actionParts) > 0 {
		line += "; " + strings.Join(actionParts, ", ")
	}
	if strings.HasSuffix(line, ".") {
		return line
	}
	return line + "."
}

func BrowserToolMetadataRouteHintsDisplayText(h BrowserToolMetadataRouteHints) string {
	if summary := BrowserToolMetadataHiddenDefaultCandidateText(h, BrowserToolMetadataRouteHintActions{}); summary != "" {
		return summary
	}
	return BrowserToolMetadataRouteHintsDetailText(h)
}

func (h BrowserToolMetadataRouteHints) Empty() bool {
	return normalizeBrowserRuntimeInfo(h.RuntimeInfo) == (BrowserRuntimeInfo{}) &&
		strings.TrimSpace(h.Source) == "" &&
		strings.TrimSpace(h.Endpoint) == "" &&
		normalizeBrowserRuntimeInfo(h.DefaultCandidateRoute) == (BrowserRuntimeInfo{}) &&
		strings.TrimSpace(h.DefaultCandidateSource) == "" &&
		strings.TrimSpace(h.DefaultCandidateEndpoint) == "" &&
		strings.TrimSpace(h.Surface) == "" &&
		len(mergeToolMetadataStrings(nil, h.OptInTargets)) == 0 &&
		!h.HiddenDefaultCandidate &&
		strings.TrimSpace(h.SubstratePosture) == "" &&
		strings.TrimSpace(h.SubstrateStatus) == "" &&
		strings.TrimSpace(h.SelectionStrategy) == "" &&
		strings.TrimSpace(h.SubstrateReason) == "" &&
		!h.RepairAvailable
}

func (h BrowserToolMetadataRouteHints) DynamicGroupSelectors() []string {
	groups := make([]string, 0, 4)
	if value := strings.TrimSpace(h.SubstratePosture); value != "" {
		groups = append(groups, "group:browser-substrate:"+value)
	}
	if value := strings.TrimSpace(h.SubstrateStatus); value != "" {
		groups = append(groups, "group:browser-substrate-status:"+value)
	}
	if value := strings.TrimSpace(h.SelectionStrategy); value != "" {
		groups = append(groups, "group:browser-default-selection:"+value)
	}
	if h.HiddenDefaultCandidate {
		groups = append(groups, "group:browser-default-candidate:hidden_managed")
	}
	if h.RepairAvailable {
		groups = append(groups, "group:browser-bootstrap-repairable")
	}
	return normalizeToolList(groups)
}

func (h BrowserToolMetadataRouteHints) DefaultCandidateLabel() string {
	if h.HiddenDefaultCandidate {
		return "hidden_managed"
	}
	return ""
}

func BrowserToolMetadataRouteHintsForMetadata(meta ToolMetadata) BrowserToolMetadataRouteHints {
	hints := BrowserToolMetadataRouteHints{}
	for _, item := range meta.Capabilities {
		value := strings.ToLower(strings.TrimSpace(item))
		switch {
		case strings.HasPrefix(value, "browser_source:"):
			hints.Source = strings.TrimSpace(strings.TrimPrefix(value, "browser_source:"))
		case strings.HasPrefix(value, "browser_endpoint:"):
			hints.Endpoint = strings.TrimSpace(strings.TrimPrefix(value, "browser_endpoint:"))
		case strings.HasPrefix(value, "browser_backend:"):
			hints.RuntimeInfo.Backend = strings.TrimPrefix(value, "browser_backend:")
		case strings.HasPrefix(value, "browser_profile:"):
			hints.RuntimeInfo.Profile = strings.TrimPrefix(value, "browser_profile:")
		case strings.HasPrefix(value, "browser_target:"):
			hints.RuntimeInfo.Target = strings.TrimPrefix(value, "browser_target:")
		case strings.HasPrefix(value, "browser_default_candidate_source:"):
			hints.DefaultCandidateSource = strings.TrimSpace(strings.TrimPrefix(value, "browser_default_candidate_source:"))
		case strings.HasPrefix(value, "browser_default_candidate_endpoint:"):
			hints.DefaultCandidateEndpoint = strings.TrimSpace(strings.TrimPrefix(value, "browser_default_candidate_endpoint:"))
		case strings.HasPrefix(value, "browser_default_candidate_backend:"):
			hints.DefaultCandidateRoute.Backend = strings.TrimPrefix(value, "browser_default_candidate_backend:")
		case strings.HasPrefix(value, "browser_default_candidate_profile:"):
			hints.DefaultCandidateRoute.Profile = strings.TrimPrefix(value, "browser_default_candidate_profile:")
		case strings.HasPrefix(value, "browser_default_candidate_target:"):
			hints.DefaultCandidateRoute.Target = strings.TrimPrefix(value, "browser_default_candidate_target:")
		case strings.HasPrefix(value, "browser_substrate_posture:"):
			hints.SubstratePosture = strings.TrimPrefix(value, "browser_substrate_posture:")
		case strings.HasPrefix(value, "browser_substrate_status:"):
			hints.SubstrateStatus = strings.TrimPrefix(value, "browser_substrate_status:")
		case strings.HasPrefix(value, "browser_substrate_selection_strategy:"):
			hints.SelectionStrategy = strings.TrimPrefix(value, "browser_substrate_selection_strategy:")
		case strings.HasPrefix(value, "browser_substrate_reason:"):
			hints.SubstrateReason = strings.TrimPrefix(value, "browser_substrate_reason:")
		case value == "browser_surface:explicit_managed_opt_in":
			hints.Surface = "explicit_managed_opt_in"
		case value == "browser_default_candidate:hidden_managed":
			hints.HiddenDefaultCandidate = true
		case value == "browser_bootstrap_repair_available":
			hints.RepairAvailable = true
		case strings.HasPrefix(value, "browser_opt_in_target:"):
			if target := strings.TrimPrefix(value, "browser_opt_in_target:"); target != "" {
				hints.OptInTargets = append(hints.OptInTargets, target)
			}
		}
	}
	hints.RuntimeInfo = normalizeBrowserRuntimeInfo(hints.RuntimeInfo)
	hints.DefaultCandidateRoute = normalizeBrowserRuntimeInfo(hints.DefaultCandidateRoute)
	if hints.DefaultCandidateRoute != (BrowserRuntimeInfo{}) {
		hints.HiddenDefaultCandidate = true
	}
	hints.OptInTargets = mergeToolMetadataStrings(nil, hints.OptInTargets)
	return hints
}

func ResolveBrowserToolMetadataRouteHints(toolNames []string, metadata map[string]ToolMetadata) BrowserToolMetadataRouteHints {
	if len(toolNames) == 0 || len(metadata) == 0 {
		return BrowserToolMetadataRouteHints{}
	}
	names := append([]string(nil), toolNames...)
	hasBrowserAct := false
	for _, item := range names {
		if NormalizeToolName(item) == "browser_act" {
			hasBrowserAct = true
			break
		}
	}
	if !hasBrowserAct {
		names = append(names, "browser_act")
	}
	hints := BrowserToolMetadataRouteHints{}
	for _, rawName := range names {
		name := NormalizeToolName(rawName)
		if name == "" {
			continue
		}
		meta, ok := metadata[name]
		if !ok || len(meta.Capabilities) == 0 {
			continue
		}
		next := BrowserToolMetadataRouteHintsForMetadata(meta)
		if hints.Surface == "" && next.Surface != "" {
			hints.Surface = next.Surface
		}
		hints.OptInTargets = mergeToolMetadataStrings(hints.OptInTargets, next.OptInTargets)
		if hints.Source == "" && next.Source != "" {
			hints.Source = next.Source
		}
		if hints.Endpoint == "" && next.Endpoint != "" {
			hints.Endpoint = next.Endpoint
		}
		if hints.RuntimeInfo == (BrowserRuntimeInfo{}) && next.RuntimeInfo != (BrowserRuntimeInfo{}) {
			hints.RuntimeInfo = next.RuntimeInfo
		}
		if hints.DefaultCandidateSource == "" && next.DefaultCandidateSource != "" {
			hints.DefaultCandidateSource = next.DefaultCandidateSource
		}
		if hints.DefaultCandidateEndpoint == "" && next.DefaultCandidateEndpoint != "" {
			hints.DefaultCandidateEndpoint = next.DefaultCandidateEndpoint
		}
		if hints.DefaultCandidateRoute == (BrowserRuntimeInfo{}) && next.DefaultCandidateRoute != (BrowserRuntimeInfo{}) {
			hints.DefaultCandidateRoute = next.DefaultCandidateRoute
		}
		if hints.SubstratePosture == "" && next.SubstratePosture != "" {
			hints.SubstratePosture = next.SubstratePosture
		}
		if hints.SubstrateStatus == "" && next.SubstrateStatus != "" {
			hints.SubstrateStatus = next.SubstrateStatus
		}
		if hints.SelectionStrategy == "" && next.SelectionStrategy != "" {
			hints.SelectionStrategy = next.SelectionStrategy
		}
		if hints.SubstrateReason == "" && next.SubstrateReason != "" {
			hints.SubstrateReason = next.SubstrateReason
		}
		hints.HiddenDefaultCandidate = hints.HiddenDefaultCandidate || next.HiddenDefaultCandidate
		hints.RepairAvailable = hints.RepairAvailable || next.RepairAvailable
	}
	if hints.RuntimeInfo.Target == "" && len(hints.OptInTargets) == 1 {
		hints.RuntimeInfo.Target = hints.OptInTargets[0]
	}
	return hints.Canonicalized()
}

// ApplyBrowserToolMetadataRouteHints overlays structured browser route hints onto
// tool metadata using the shared capability contract.
func ApplyBrowserToolMetadataRouteHints(meta ToolMetadata, hints BrowserToolMetadataRouteHints) ToolMetadata {
	hints.RuntimeInfo = normalizeBrowserRuntimeInfo(hints.RuntimeInfo)
	hints.Source = strings.TrimSpace(hints.Source)
	hints.Endpoint = strings.TrimSpace(hints.Endpoint)
	hints.DefaultCandidateSource = strings.TrimSpace(hints.DefaultCandidateSource)
	hints.DefaultCandidateEndpoint = strings.TrimSpace(hints.DefaultCandidateEndpoint)
	hints.DefaultCandidateRoute = normalizeBrowserRuntimeInfo(hints.DefaultCandidateRoute)
	hints.Surface = strings.ToLower(strings.TrimSpace(hints.Surface))
	hints.OptInTargets = mergeToolMetadataStrings(nil, hints.OptInTargets)
	if hints.DefaultCandidateRoute != (BrowserRuntimeInfo{}) {
		hints.HiddenDefaultCandidate = true
	}

	capabilities := append([]string(nil), meta.Capabilities...)
	if hints.RuntimeInfo.Backend != "" {
		capabilities = append(capabilities, "browser_backend:"+hints.RuntimeInfo.Backend)
	}
	if hints.Source != "" {
		capabilities = append(capabilities, "browser_source:"+hints.Source)
	}
	if hints.Endpoint != "" {
		capabilities = append(capabilities, "browser_endpoint:"+hints.Endpoint)
	}
	if hints.RuntimeInfo.Profile != "" {
		capabilities = append(capabilities, "browser_profile:"+hints.RuntimeInfo.Profile)
	}
	if hints.RuntimeInfo.Target != "" {
		capabilities = append(capabilities, "browser_target:"+hints.RuntimeInfo.Target)
	}
	if hints.DefaultCandidateSource != "" {
		capabilities = append(capabilities, "browser_default_candidate_source:"+hints.DefaultCandidateSource)
	}
	if hints.DefaultCandidateEndpoint != "" {
		capabilities = append(capabilities, "browser_default_candidate_endpoint:"+hints.DefaultCandidateEndpoint)
	}
	if hints.DefaultCandidateRoute.Backend != "" {
		capabilities = append(capabilities, "browser_default_candidate_backend:"+hints.DefaultCandidateRoute.Backend)
	}
	if hints.DefaultCandidateRoute.Profile != "" {
		capabilities = append(capabilities, "browser_default_candidate_profile:"+hints.DefaultCandidateRoute.Profile)
	}
	if hints.DefaultCandidateRoute.Target != "" {
		capabilities = append(capabilities, "browser_default_candidate_target:"+hints.DefaultCandidateRoute.Target)
	}
	if hints.HiddenDefaultCandidate {
		capabilities = append(capabilities, "browser_default_candidate:hidden_managed")
	}
	if hints.SubstratePosture != "" {
		capabilities = append(capabilities, "browser_substrate_posture:"+strings.ToLower(strings.TrimSpace(hints.SubstratePosture)))
	}
	if hints.SubstrateStatus != "" {
		capabilities = append(capabilities, "browser_substrate_status:"+strings.ToLower(strings.TrimSpace(hints.SubstrateStatus)))
	}
	if hints.SelectionStrategy != "" {
		capabilities = append(capabilities, "browser_substrate_selection_strategy:"+strings.ToLower(strings.TrimSpace(hints.SelectionStrategy)))
	}
	if hints.SubstrateReason != "" {
		capabilities = append(capabilities, "browser_substrate_reason:"+strings.ToLower(strings.TrimSpace(hints.SubstrateReason)))
	}
	if hints.RepairAvailable {
		capabilities = append(capabilities, "browser_bootstrap_repair_available")
	}
	if hints.Surface == "explicit_managed_opt_in" {
		capabilities = append(capabilities, "browser_surface:explicit_managed_opt_in")
		for _, target := range hints.OptInTargets {
			capabilities = append(capabilities, "browser_opt_in_target:"+target)
		}
	}
	meta.Capabilities = mergeToolMetadataStrings(nil, capabilities)
	return meta
}
