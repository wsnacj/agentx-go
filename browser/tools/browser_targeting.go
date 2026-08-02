package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"regexp"
	"strconv"
	"strings"

	agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"
)

type browserToolTarget struct {
	Value    string
	TabIndex int
	TargetID string
	Explicit bool
}

type browserElementTarget struct {
	Selector string
	Ref      string
	Payload  browserElementRefPayload
}

type browserElementRefPayload struct {
	Selector      string `json:"selector,omitempty"`
	SelectorIndex int    `json:"selector_index,omitempty"`
	FramePath     string `json:"frame_path,omitempty"`
	NativeRef     string `json:"native_ref,omitempty"`
	Role          string `json:"role,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Label         string `json:"label,omitempty"`
	Type          string `json:"type,omitempty"`
	Href          string `json:"href,omitempty"`
	Placeholder   string `json:"placeholder,omitempty"`
	PageURL       string `json:"page_url,omitempty"`
	PageOrigin    string `json:"page_origin,omitempty"`
	PagePath      string `json:"page_path,omitempty"`
	PageTitle     string `json:"page_title,omitempty"`
	TabIndex      int    `json:"tab_index,omitempty"`
}

var browserElementRefStringFieldNames = map[string][]string{
	"selector":    {"selector"},
	"native_ref":  {"native_ref"},
	"role":        {"role"},
	"tag":         {"tag"},
	"label":       {"label", "Label"},
	"type":        {"type"},
	"href":        {"href"},
	"placeholder": {"placeholder"},
	"page_url":    {"page_url"},
	"page_origin": {"page_origin"},
	"page_path":   {"page_path"},
	"page_title":  {"page_title"},
}

var browserLooseContainsSelectorPattern = regexp.MustCompile(`^\s*([A-Za-z*][A-Za-z0-9_-]*)\s*:\s*contains\((.+)\)\s*$`)

func browserElementRefPayloadHasActionableLocator(payload browserElementRefPayload) bool {
	return strings.TrimSpace(payload.Selector) != "" ||
		strings.TrimSpace(payload.NativeRef) != "" ||
		strings.TrimSpace(payload.Label) != "" ||
		strings.TrimSpace(payload.Href) != "" ||
		strings.TrimSpace(payload.Placeholder) != "" ||
		(strings.TrimSpace(payload.Tag) != "" && strings.TrimSpace(payload.Type) != "")
}

func prefersSafari(browserApp string) bool {
	app := strings.ToLower(strings.TrimSpace(browserApp))
	return app == "" || app == "safari"
}

func parseSafariTabsPayload(raw string, activeIndex int) []BrowserTab {
	lines := strings.Split(strings.TrimSpace(raw), "<<<TAB>>>")
	out := make([]BrowserTab, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "<<<FIELD>>>", 3)
		if len(parts) < 3 {
			continue
		}
		index := 0
		fmt.Sscanf(strings.TrimSpace(parts[0]), "%d", &index)
		out = append(out, BrowserTab{
			Index:  index,
			Title:  strings.TrimSpace(parts[1]),
			URL:    strings.TrimSpace(parts[2]),
			Active: index == activeIndex,
		})
	}
	return out
}

func resolveBrowserToolTarget(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, browserApp string, params map[string]any) (browserToolTarget, error) {
	explicitRuntimeTarget := browserHasExplicitRuntimeTarget(params)
	requiresExplicitHostRuntimeTarget := hiddenImplicitHostDefaultBase && !explicitRuntimeTarget
	requireExplicitHostForTabTarget := requiresExplicitHostRuntimeTarget &&
		(strings.TrimSpace(runtimeInfo.Target) == "" || strings.EqualFold(strings.TrimSpace(runtimeInfo.Target), "host"))
	resolveCurrentTarget := func() (browserToolTarget, error) {
		sessionID := ToolSessionIDFromContext(ctx)
		if registry == nil || sessionID == "" {
			return browserToolTarget{}, nil
		}
		route := browserSessionRoute(runtimeInfo, browserApp, "")
		allowDefaultFallback := strings.TrimSpace(runtimeInfo.Profile) == "" &&
			!hiddenImplicitHostDefaultBase
		current, ok := agentxbrowserruntime.ResolveSharedSessionBrowserCurrentTarget(
			registry,
			sessionID,
			route,
			allowDefaultFallback,
		)
		if !ok || strings.TrimSpace(current.ID) == "" {
			return browserToolTarget{}, nil
		}
		if hiddenImplicitHostDefaultBase {
			currentTarget := strings.ToLower(strings.TrimSpace(current.Target))
			if currentTarget == "" || currentTarget == "host" {
				return browserToolTarget{}, nil
			}
		}
		return browserToolTarget{
			Value:    "target:" + strings.TrimSpace(current.ID),
			TabIndex: current.TabIndex,
			TargetID: strings.TrimSpace(current.ID),
			Explicit: true,
		}, nil
	}

	rawTarget := strings.TrimSpace(firstString(params, "target"))
	tabIndex := firstInt(params, "tab_index", "index")
	if rawTarget == "" {
		if tabIndex <= 0 {
			return resolveCurrentTarget()
		}
		if requireExplicitHostForTabTarget {
			return browserToolTarget{}, browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(browserToolTarget{
				Value:    fmt.Sprintf("tab:%d", tabIndex),
				TabIndex: tabIndex,
				Explicit: true,
			})
		}
		return browserToolTarget{
			Value:    fmt.Sprintf("tab:%d", tabIndex),
			TabIndex: tabIndex,
			Explicit: true,
		}, nil
	}
	target, err := parseBrowserToolTarget(rawTarget)
	if err != nil {
		return browserToolTarget{}, err
	}
	if tabIndex > 0 && target.TabIndex != tabIndex {
		return browserToolTarget{}, fmt.Errorf("target conflicts with tab_index")
	}
	if tabIndex > 0 {
		target.TabIndex = tabIndex
		target.Value = fmt.Sprintf("tab:%d", tabIndex)
		target.TargetID = ""
	}
	if target.Value == "current" && target.TargetID == "" && target.TabIndex <= 0 {
		resolved, err := resolveCurrentTarget()
		if err != nil {
			return browserToolTarget{}, err
		}
		if strings.TrimSpace(resolved.TargetID) != "" || resolved.TabIndex > 0 {
			return resolved, nil
		}
		if requiresExplicitHostRuntimeTarget {
			return browserToolTarget{}, browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target)
		}
	}
	if target.TargetID != "" {
		sessionID := ToolSessionIDFromContext(ctx)
		if sessionID == "" {
			return browserToolTarget{}, fmt.Errorf("target handle requires a session-scoped tool context")
		}
		if registry == nil {
			return browserToolTarget{}, fmt.Errorf("target handle requires browser session tracking")
		}
		tracked, ok := agentxbrowserruntime.ResolveSharedSessionBrowserTarget(
			registry,
			sessionID,
			agentxbrowserruntime.BrowserSessionRoute{},
			target.TargetID,
			0,
			false,
		)
		if !ok {
			return browserToolTarget{}, fmt.Errorf("target handle not found: %s", target.TargetID)
		}
		if requiresExplicitHostRuntimeTarget {
			trackedTarget := strings.ToLower(strings.TrimSpace(tracked.Target))
			if trackedTarget == "" || trackedTarget == "host" {
				return browserToolTarget{}, browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target)
			}
		}
		target.TabIndex = tracked.TabIndex
		target.Value = "target:" + target.TargetID
	} else if target.TabIndex > 0 {
		if requireExplicitHostForTabTarget {
			return browserToolTarget{}, browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target)
		}
		target.TargetID = browserTargetIDForTab(ctx, registry, runtimeInfo, browserApp, target.TabIndex)
	}
	target.Explicit = true
	return target, nil
}

func browserRequestedRuntimeTarget(params map[string]any) string {
	return strings.ToLower(strings.TrimSpace(firstString(params, "runtime_target", "browser_target", "placement")))
}

func browserHasExplicitRuntimeTarget(params map[string]any) bool {
	return browserRequestedRuntimeTarget(params) != ""
}

func browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target browserToolTarget) error {
	label := strings.TrimSpace(target.Value)
	if label == "" && target.TabIndex > 0 {
		label = fmt.Sprintf("tab:%d", target.TabIndex)
	}
	if label == "" {
		label = "current"
	}
	return newBrowserRouteGateError(
		browserRouteGateTarget,
		fmt.Sprintf("target %q requires explicit runtime_target=host because the default browser route falls back to the legacy system host path", label),
	)
}

func browserTargetlessImplicitLegacyHostCurrentPageFallback(hiddenImplicitHostDefaultBase bool, target browserToolTarget, requestURL string) bool {
	if !hiddenImplicitHostDefaultBase {
		return false
	}
	if strings.TrimSpace(requestURL) != "" {
		return false
	}
	return strings.TrimSpace(target.TargetID) == "" && target.TabIndex <= 0
}

func browserPageActionDispatchBaseForManagedDefaultRoute(base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, target browserToolTarget, requestURL string, explicitTarget string) BrowserRuntimeInfo {
	base = normalizeBrowserRuntimeInfo(base)
	if !hiddenImplicitHostDefaultBase {
		return base
	}
	if targetRoute := strings.ToLower(strings.TrimSpace(base.Target)); targetRoute != "" && targetRoute != "host" {
		return base
	}
	if strings.TrimSpace(requestURL) != "" || strings.TrimSpace(explicitTarget) != "" {
		return base
	}
	if strings.TrimSpace(target.TargetID) != "" || target.TabIndex > 0 {
		return base
	}
	return BrowserRuntimeInfo{}
}

func browserPageActionDispatchBaseForManagedDefaultParams(base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, params map[string]any) BrowserRuntimeInfo {
	return browserPageActionDispatchBaseForManagedDefaultRoute(
		base,
		hiddenImplicitHostDefaultBase,
		browserToolTarget{TabIndex: firstInt(params, "tab_index")},
		firstString(params, "url"),
		firstString(params, "target"),
	)
}

func browserURLActionDispatchBaseForManagedDefaultRoute(base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, target browserToolTarget, explicitTarget string) BrowserRuntimeInfo {
	base = normalizeBrowserRuntimeInfo(base)
	if !hiddenImplicitHostDefaultBase {
		return base
	}
	if targetRoute := strings.ToLower(strings.TrimSpace(base.Target)); targetRoute != "" && targetRoute != "host" {
		return base
	}
	if strings.TrimSpace(explicitTarget) != "" {
		return base
	}
	if strings.TrimSpace(target.TargetID) != "" || target.TabIndex > 0 {
		return base
	}
	return BrowserRuntimeInfo{}
}

func browserURLActionDispatchBaseForManagedDefaultParams(base BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, params map[string]any) BrowserRuntimeInfo {
	return browserURLActionDispatchBaseForManagedDefaultRoute(
		base,
		hiddenImplicitHostDefaultBase,
		browserToolTarget{TabIndex: firstInt(params, "tab_index")},
		firstString(params, "target"),
	)
}

func browserImplicitLegacyHostURLRouteFallback(hiddenImplicitHostDefaultBase bool, explicitRuntimeTarget bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) bool {
	if !hiddenImplicitHostDefaultBase || explicitRuntimeTarget {
		return false
	}
	if strings.TrimSpace(requestURL) == "" || target.TabIndex > 0 {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(runtimeInfo.Target), "host")
}

func browserImplicitLegacyHostTabTargetFallback(hiddenImplicitHostDefaultBase bool, target browserToolTarget) bool {
	return browserImplicitLegacyHostTabTargetFallbackForRuntime(hiddenImplicitHostDefaultBase, BrowserRuntimeInfo{Target: "host"}, target)
}

func browserImplicitLegacyHostTabTargetFallbackForRuntime(hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget) bool {
	if !hiddenImplicitHostDefaultBase || target.TabIndex <= 0 {
		return false
	}
	targetRoute := strings.ToLower(strings.TrimSpace(runtimeInfo.Target))
	return targetRoute == "" || targetRoute == "host"
}

func browserImplicitLegacyHostDefaultBrowserFallback(hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, action string, target browserToolTarget) bool {
	if !hiddenImplicitHostDefaultBase {
		return false
	}
	if targetRoute := strings.ToLower(strings.TrimSpace(runtimeInfo.Target)); targetRoute != "" && targetRoute != "host" {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(action), "list") {
		return false
	}
	return strings.TrimSpace(target.TargetID) == "" && target.TabIndex <= 0
}

func browserImplicitLegacyHostNonListTabActionFallback(hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, action string) bool {
	if !hiddenImplicitHostDefaultBase {
		return false
	}
	if targetRoute := strings.ToLower(strings.TrimSpace(runtimeInfo.Target)); targetRoute != "" && targetRoute != "host" {
		return false
	}
	return !strings.EqualFold(strings.TrimSpace(action), "list")
}

func browserImplicitLegacyHostCurrentPageRequiresExplicitRuntimeTargetError(actor string) error {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = "browser tool"
	}
	return newBrowserRouteGateError(
		browserRouteGateCurrentPage,
		fmt.Sprintf("%s requires explicit runtime_target=host or an explicit target/url because the default browser route falls back to the legacy system host path", actor),
	)
}

func browserImplicitLegacyHostCurrentPageFallbackError(actor string, hiddenImplicitHostDefaultBase bool, target browserToolTarget, requestURL string) error {
	if !browserTargetlessImplicitLegacyHostCurrentPageFallback(hiddenImplicitHostDefaultBase, target, requestURL) {
		return nil
	}
	return browserImplicitLegacyHostCurrentPageRequiresExplicitRuntimeTargetError(actor)
}

func browserImplicitLegacyHostCurrentPageFallbackForRuntime(hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) bool {
	if !browserTargetlessImplicitLegacyHostCurrentPageFallback(hiddenImplicitHostDefaultBase, target, requestURL) {
		return false
	}
	targetRoute := strings.ToLower(strings.TrimSpace(runtimeInfo.Target))
	return targetRoute == "" || targetRoute == "host"
}

func browserImplicitLegacyHostCurrentPageFallbackErrorForRuntime(actor string, hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) error {
	if !browserImplicitLegacyHostCurrentPageFallbackForRuntime(hiddenImplicitHostDefaultBase, runtimeInfo, target, requestURL) {
		return nil
	}
	return browserImplicitLegacyHostCurrentPageRequiresExplicitRuntimeTargetError(actor)
}

func browserImplicitLegacyHostPageExecutionFallbackError(actor string, hiddenImplicitHostDefaultBase bool, explicitRuntimeTarget bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) error {
	if err := browserImplicitLegacyHostURLRouteFallbackError(actor, hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, requestURL); err != nil {
		return err
	}
	return browserImplicitLegacyHostCurrentPageFallbackErrorForRuntime(actor, hiddenImplicitHostDefaultBase, runtimeInfo, target, requestURL)
}

func browserImplicitLegacyHostDirectPageActionFallbackError(actor string, hiddenImplicitHostDefaultBase bool, target browserToolTarget, requestURL string) error {
	return browserImplicitLegacyHostDirectPageActionFallbackErrorForRuntime(
		actor,
		hiddenImplicitHostDefaultBase,
		false,
		BrowserRuntimeInfo{Target: "host"},
		target,
		requestURL,
	)
}

func browserImplicitLegacyHostDirectPageActionFallbackErrorForRuntime(actor string, hiddenImplicitHostDefaultBase bool, explicitRuntimeTarget bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) error {
	if err := browserImplicitLegacyHostTabTargetFallbackErrorForRuntime(hiddenImplicitHostDefaultBase, runtimeInfo, target); err != nil {
		return err
	}
	return browserImplicitLegacyHostPageExecutionFallbackError(
		actor,
		hiddenImplicitHostDefaultBase,
		explicitRuntimeTarget,
		runtimeInfo,
		target,
		requestURL,
	)
}

func browserImplicitLegacyHostDefaultBrowserRequiresExplicitRuntimeTargetError(actor string) error {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = "browser tool"
	}
	return newBrowserRouteGateError(
		browserRouteGateDefaultBrowser,
		fmt.Sprintf("%s requires explicit runtime_target=host because the default browser route falls back to the legacy system host path", actor),
	)
}

func browserImplicitLegacyHostDefaultBrowserFallbackError(actor string, hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, action string, target browserToolTarget) error {
	if !browserImplicitLegacyHostDefaultBrowserFallback(hiddenImplicitHostDefaultBase, runtimeInfo, action, target) {
		return nil
	}
	return browserImplicitLegacyHostDefaultBrowserRequiresExplicitRuntimeTargetError(actor)
}

func browserImplicitLegacyHostURLRouteRequiresExplicitRuntimeTargetError(actor string) error {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		actor = "browser tool"
	}
	return newBrowserRouteGateError(
		browserRouteGateURLRoute,
		fmt.Sprintf("%s requires explicit runtime_target because the default browser route falls back to the legacy system host path", actor),
	)
}

func browserImplicitLegacyHostURLRouteFallbackError(actor string, hiddenImplicitHostDefaultBase bool, explicitRuntimeTarget bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) error {
	if !browserImplicitLegacyHostURLRouteFallback(hiddenImplicitHostDefaultBase, explicitRuntimeTarget, runtimeInfo, target, requestURL) {
		return nil
	}
	return browserImplicitLegacyHostURLRouteRequiresExplicitRuntimeTargetError(actor)
}

func browserImplicitLegacyHostURLNavigationFallbackError(actor string, hiddenImplicitHostDefaultBase bool, explicitRuntimeTarget bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget, requestURL string) error {
	if !hiddenImplicitHostDefaultBase || explicitRuntimeTarget || strings.TrimSpace(requestURL) == "" {
		return nil
	}
	if targetRoute := strings.ToLower(strings.TrimSpace(runtimeInfo.Target)); targetRoute != "" && targetRoute != "host" {
		return nil
	}
	if strings.TrimSpace(target.TargetID) != "" || target.TabIndex > 0 {
		return browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target)
	}
	return browserImplicitLegacyHostURLRouteRequiresExplicitRuntimeTargetError(actor)
}

func browserImplicitLegacyHostTabsActionFallbackError(actor string, hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, action string, target browserToolTarget) error {
	if err := browserImplicitLegacyHostDefaultBrowserFallbackError(actor, hiddenImplicitHostDefaultBase, runtimeInfo, action, target); err != nil {
		return err
	}
	if !browserImplicitLegacyHostNonListTabActionFallback(hiddenImplicitHostDefaultBase, runtimeInfo, action) {
		return nil
	}
	return browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target)
}

func browserImplicitLegacyHostRuntimeActionRequiresExplicitRuntimeTargetError(action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	actor := "browser_runtime"
	if action != "" {
		actor = fmt.Sprintf("browser_runtime action %s", action)
	}
	return newBrowserRouteGateError(
		browserRouteGateRuntimeAction,
		fmt.Sprintf("%s requires explicit runtime_target because the default browser route falls back to the legacy system host path", actor),
	)
}

func browserImplicitLegacyHostManagedActKindRequiresExplicitRuntimeTargetError(kind string) error {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		kind = "unknown"
	}
	return newBrowserRouteGateError(
		browserRouteGateManagedAction,
		fmt.Sprintf(
			"browser_act kind %s requires explicit runtime_target on a managed browser route because the default browser route falls back to the legacy system host path",
			kind,
		),
	)
}

func browserImplicitLegacyHostSupportsActKind(kind string) bool {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "open", "navigate", "extract", "snapshot", "screenshot", "click", "type", "evaluate", "wait", "list_tabs", "focus_tab", "close_tab":
		return true
	default:
		return false
	}
}

func browserImplicitLegacyHostManagedActKindFallbackError(hiddenImplicitHostDefaultBase bool, explicitRuntimeTarget bool, runtimeInfo BrowserRuntimeInfo, kind string) error {
	if !hiddenImplicitHostDefaultBase || explicitRuntimeTarget || browserImplicitLegacyHostSupportsActKind(kind) {
		return nil
	}
	if targetRoute := strings.ToLower(strings.TrimSpace(runtimeInfo.Target)); targetRoute != "" && targetRoute != "host" {
		return nil
	}
	return browserImplicitLegacyHostManagedActKindRequiresExplicitRuntimeTargetError(kind)
}

func browserRuntimeCanonicalAction(action string) string {
	action = browserNormalizeToolToken(action)
	if action == "" {
		return "status"
	}
	if action == "doctor" {
		return "status"
	}
	return action
}

func browserImplicitLegacyHostRuntimeActionUsesDiagnosticsPath(action string) bool {
	switch browserRuntimeCanonicalAction(action) {
	case "status", "profiles", "sessions", "workbench":
		return true
	default:
		return false
	}
}

func browserRuntimeActionRequiresExplicitRuntimeTargetForImplicitFallback(action string) bool {
	return !browserImplicitLegacyHostRuntimeActionUsesDiagnosticsPath(action)
}

func browserRuntimeUsesImplicitLegacyHostDiagnosticsDegradePath(action string, hiddenImplicitHostDefaultBase bool, requestedProfile string, requestedTarget string, canUseManagedSessionRoute bool) bool {
	if !hiddenImplicitHostDefaultBase || strings.TrimSpace(requestedTarget) != "" {
		return false
	}
	if canUseManagedSessionRoute &&
		browserImplicitLegacyHostRuntimeActionUsesDiagnosticsPath(action) &&
		strings.TrimSpace(requestedProfile) == "" {
		return false
	}
	_, ok := browserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile(
		action,
		requestedProfile,
		strings.TrimSpace(defaultBrowserRuntimeInfo().Profile),
	)
	return ok
}

func browserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile(action string, requestedProfile string, defaultProfile string) (string, bool) {
	if !browserImplicitLegacyHostRuntimeActionUsesDiagnosticsPath(action) {
		return "", false
	}
	defaultProfile = strings.TrimSpace(defaultProfile)
	if defaultProfile == "" {
		defaultProfile = strings.TrimSpace(defaultBrowserRuntimeInfo().Profile)
	}
	requestedProfile = strings.TrimSpace(requestedProfile)
	if requestedProfile == "" {
		return "", true
	}
	if strings.EqualFold(requestedProfile, defaultProfile) {
		return "", true
	}
	return requestedProfile, false
}

func browserImplicitLegacyHostRuntimeCanUseCachedDiagnosticsSnapshot(action string, requestedProfile string, defaultProfile string) (string, bool) {
	normalizedProfile, ok := browserImplicitLegacyHostRuntimeDiagnosticsRequestedProfile(action, requestedProfile, defaultProfile)
	if !ok {
		return normalizedProfile, false
	}
	if browserRuntimeCanonicalAction(action) == "profiles" && normalizedProfile == "" {
		return normalizedProfile, false
	}
	return normalizedProfile, true
}

func browserImplicitLegacyHostDefaultRouteRequiresExplicitRuntimeTargetError() error {
	return newBrowserRouteGateError(
		browserRouteGateDefaultRequest,
		"default browser route requires explicit runtime_target because the default browser route falls back to the legacy system host path",
	)
}

func browserImplicitLegacyHostDefaultProfileRequiresExplicitRuntimeTargetError(profile string) error {
	profile = strings.TrimSpace(profile)
	return newBrowserRouteGateError(
		browserRouteGateDefaultRequest,
		fmt.Sprintf("profile %q requires an explicit runtime_target because the default browser route falls back to the legacy system host path", profile),
	)
}

func browserImplicitLegacyHostDefaultRequestError(hiddenImplicitHostDefaultBase bool, requested BrowserRuntimeInfo, defaultProfile string) error {
	if !hiddenImplicitHostDefaultBase {
		return nil
	}
	requested = normalizeBrowserRuntimeInfo(requested)
	if strings.TrimSpace(requested.Target) != "" {
		return nil
	}
	defaultProfile = strings.ToLower(strings.TrimSpace(defaultProfile))
	requestedProfile := strings.ToLower(strings.TrimSpace(requested.Profile))
	if requestedProfile == "" || (defaultProfile != "" && requestedProfile == defaultProfile) {
		return browserImplicitLegacyHostDefaultRouteRequiresExplicitRuntimeTargetError()
	}
	return browserImplicitLegacyHostDefaultProfileRequiresExplicitRuntimeTargetError(requested.Profile)
}

func browserImplicitLegacyHostRouteErrMatchesDefaultRequestError(hiddenImplicitHostDefaultBase bool, requested BrowserRuntimeInfo, defaultProfile string, err error) bool {
	if err == nil {
		return false
	}
	if browserImplicitLegacyHostDefaultRequestError(hiddenImplicitHostDefaultBase, requested, defaultProfile) == nil {
		return false
	}
	return browserRouteErrorHasGateKind(err, browserRouteGateDefaultRequest)
}

func browserImplicitLegacyHostTabTargetFallbackError(hiddenImplicitHostDefaultBase bool, target browserToolTarget) error {
	return browserImplicitLegacyHostTabTargetFallbackErrorForRuntime(hiddenImplicitHostDefaultBase, BrowserRuntimeInfo{Target: "host"}, target)
}

func browserImplicitLegacyHostTabTargetFallbackErrorForRuntime(hiddenImplicitHostDefaultBase bool, runtimeInfo BrowserRuntimeInfo, target browserToolTarget) error {
	if !browserImplicitLegacyHostTabTargetFallbackForRuntime(hiddenImplicitHostDefaultBase, runtimeInfo, target) {
		return nil
	}
	return browserImplicitLegacyHostTargetRequiresExplicitRuntimeTargetError(target)
}

func parseBrowserToolTarget(raw string) (browserToolTarget, error) {
	lower := strings.ToLower(strings.TrimSpace(raw))
	switch lower {
	case "":
		return browserToolTarget{}, nil
	case "current", "active", "current_tab", "active_tab":
		return browserToolTarget{Value: "current"}, nil
	}
	candidate := lower
	for _, prefix := range []string{"tab:", "tab/", "tab=", "tab#"} {
		if strings.HasPrefix(candidate, prefix) {
			candidate = strings.TrimSpace(strings.TrimPrefix(candidate, prefix))
			break
		}
	}
	for _, prefix := range []string{"target:", "target/", "target=", "target#", "session:", "session/", "session=", "session#"} {
		if strings.HasPrefix(candidate, prefix) {
			targetID := strings.TrimSpace(strings.TrimPrefix(candidate, prefix))
			if targetID == "" {
				return browserToolTarget{}, fmt.Errorf("target handle cannot be empty")
			}
			return browserToolTarget{
				Value:    "target:" + targetID,
				TargetID: targetID,
			}, nil
		}
	}
	tabIndex, err := strconv.Atoi(candidate)
	if err != nil || tabIndex <= 0 {
		return browserToolTarget{}, fmt.Errorf("target must be \"current\", \"tab:<n>\", or \"target:<id>\"")
	}
	return browserToolTarget{
		Value:    fmt.Sprintf("tab:%d", tabIndex),
		TabIndex: tabIndex,
	}, nil
}

func resolveBrowserElementTarget(selector string, ref string) (browserElementTarget, error) {
	target := browserElementTarget{
		Selector: browserSanitizeLooseArgumentString(selector),
		Ref:      browserSanitizeLooseArgumentString(ref),
	}
	var payload browserElementRefPayload
	if target.Ref != "" {
		var err error
		payload, err = browserDecodeElementRef(target.Ref)
		if err != nil {
			return browserElementTarget{}, err
		}
	}
	decodedSelector := strings.TrimSpace(payload.Selector)
	switch {
	case target.Selector != "" && decodedSelector != "" && target.Selector != decodedSelector:
		return browserElementTarget{}, fmt.Errorf("ref conflicts with selector")
	case target.Selector == "":
		target.Selector = decodedSelector
	}
	payload.Selector = firstNonEmpty(strings.TrimSpace(target.Selector), decodedSelector)
	target.Payload = payload
	if target.Selector != "" && target.Ref == "" {
		target.Ref = browserElementRefForSelector(target.Selector)
	}
	return target, nil
}

func browserPageBoundElementTargetForRouteSelection(params map[string]any) (browserElementTarget, bool) {
	candidates := []struct {
		selector string
		ref      string
	}{
		{
			selector: firstString(params, "selector", "element"),
			ref:      firstString(params, "ref", "element_ref", "input_ref"),
		},
		{
			selector: firstString(params, "start_selector", "startSelector"),
			ref:      firstString(params, "start_ref", "startRef"),
		},
		{
			selector: firstString(params, "end_selector", "endSelector"),
			ref:      firstString(params, "end_ref", "endRef"),
		},
	}
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.selector) == "" && strings.TrimSpace(candidate.ref) == "" {
			continue
		}
		target, err := resolveBrowserElementTarget(candidate.selector, candidate.ref)
		if err != nil {
			continue
		}
		payload := target.Payload
		if !browserElementRefHasPageBinding(payload) && strings.TrimSpace(target.Ref) != "" {
			if decoded, err := browserDecodeElementRef(target.Ref); err == nil {
				payload = decoded
				target.Payload = decoded
			}
		}
		if browserElementRefHasPageBinding(payload) {
			return target, true
		}
	}
	return browserElementTarget{}, false
}

func browserElementRefPayloadFromHint(hint *BrowserElementHint) browserElementRefPayload {
	if hint == nil {
		return browserElementRefPayload{}
	}
	return browserSanitizeElementRefPayload(browserElementRefPayload{
		Selector:      hint.Selector,
		SelectorIndex: hint.SelectorIndex,
		FramePath:     hint.FramePath,
		NativeRef:     hint.NativeRef,
		Role:          hint.Role,
		Tag:           hint.Tag,
		Label:         hint.Label,
		Type:          hint.Type,
		Href:          hint.Href,
		Placeholder:   hint.Placeholder,
		PageURL:       hint.PageURL,
		PageOrigin:    hint.PageOrigin,
		PagePath:      hint.PagePath,
		PageTitle:     hint.PageTitle,
		TabIndex:      hint.TabIndex,
	})
}

func browserElementRefPayloadHasMetadata(payload browserElementRefPayload) bool {
	return payload.SelectorIndex > 0 ||
		strings.TrimSpace(payload.FramePath) != "" ||
		strings.TrimSpace(payload.NativeRef) != "" ||
		strings.TrimSpace(payload.Role) != "" ||
		strings.TrimSpace(payload.Tag) != "" ||
		strings.TrimSpace(payload.Label) != "" ||
		strings.TrimSpace(payload.Type) != "" ||
		strings.TrimSpace(payload.Href) != "" ||
		strings.TrimSpace(payload.Placeholder) != "" ||
		strings.TrimSpace(payload.PageURL) != "" ||
		strings.TrimSpace(payload.PageOrigin) != "" ||
		strings.TrimSpace(payload.PagePath) != "" ||
		strings.TrimSpace(payload.PageTitle) != "" ||
		payload.TabIndex > 0
}

func browserMergeElementTargetHint(target browserElementTarget, hint *BrowserElementHint) (browserElementTarget, error) {
	hintPayload := browserElementRefPayloadFromHint(hint)
	hintSelector := strings.TrimSpace(hintPayload.Selector)
	switch {
	case target.Selector != "" && hintSelector != "" && target.Selector != hintSelector:
		return browserElementTarget{}, fmt.Errorf("element_hint conflicts with selector")
	case target.Selector == "":
		target.Selector = hintSelector
	}
	target.Payload.Selector = firstNonEmpty(strings.TrimSpace(target.Payload.Selector), strings.TrimSpace(target.Selector), hintSelector)
	if target.Payload.SelectorIndex <= 0 {
		target.Payload.SelectorIndex = hintPayload.SelectorIndex
	}
	if target.Payload.NativeRef == "" {
		target.Payload.NativeRef = hintPayload.NativeRef
	}
	if target.Payload.FramePath == "" {
		target.Payload.FramePath = hintPayload.FramePath
	}
	if target.Payload.Role == "" {
		target.Payload.Role = hintPayload.Role
	}
	if target.Payload.Tag == "" {
		target.Payload.Tag = hintPayload.Tag
	}
	if target.Payload.Label == "" {
		target.Payload.Label = hintPayload.Label
	}
	if target.Payload.Type == "" {
		target.Payload.Type = hintPayload.Type
	}
	if target.Payload.Href == "" {
		target.Payload.Href = hintPayload.Href
	}
	if target.Payload.Placeholder == "" {
		target.Payload.Placeholder = hintPayload.Placeholder
	}
	if target.Payload.PageURL == "" {
		target.Payload.PageURL = hintPayload.PageURL
	}
	if target.Payload.PageOrigin == "" {
		target.Payload.PageOrigin = hintPayload.PageOrigin
	}
	if target.Payload.PagePath == "" {
		target.Payload.PagePath = hintPayload.PagePath
	}
	if target.Payload.PageTitle == "" {
		target.Payload.PageTitle = hintPayload.PageTitle
	}
	if target.Payload.TabIndex <= 0 {
		target.Payload.TabIndex = hintPayload.TabIndex
	}
	target.Payload.Selector = strings.TrimSpace(target.Selector)
	if browserElementRefPayloadHasMetadata(target.Payload) && browserElementRefPayloadHasActionableLocator(target.Payload) {
		target.Ref = browserElementRefForPayload(target.Payload)
		return target, nil
	}
	if target.Selector == "" {
		return target, nil
	}
	if target.Ref == "" {
		target.Ref = browserElementRefForSelector(target.Selector)
	}
	return target, nil
}

func resolveBrowserElementTargetWithHint(selector string, ref string, hint *BrowserElementHint) (browserElementTarget, error) {
	target, err := resolveBrowserElementTarget(selector, ref)
	if err != nil {
		return browserElementTarget{}, err
	}
	return browserMergeElementTargetHint(target, hint)
}

func browserLooksLikeSelector(raw string) bool {
	raw = browserNormalizeToolToken(raw)
	if raw == "" {
		return false
	}
	if browserElementRefHasKnownPrefix(raw) {
		return false
	}
	if strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, ".") || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "[") {
		return true
	}
	if strings.ContainsAny(raw, "#.[]>:*=~+,") || strings.Contains(raw, "::") || strings.Contains(raw, "contains(") {
		return true
	}
	switch raw {
	case "a", "button", "input", "textarea", "select", "summary", "label", "form", "div", "span", "main", "header", "footer", "nav", "section", "article", "table", "tr", "td", "th", "img":
		return true
	default:
		return false
	}
}

func browserLooseElementLabelHint(raw string) *BrowserElementHint {
	label := browserSanitizeLooseArgumentString(raw)
	if label == "" || browserElementRefHasKnownPrefix(label) || browserLooksLikeSelector(label) {
		return nil
	}
	return &BrowserElementHint{Label: label}
}

func browserLooseTextMatchSelectorHint(raw string) *BrowserElementHint {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	match := browserLooseContainsSelectorPattern.FindStringSubmatch(raw)
	if len(match) != 3 {
		return nil
	}
	label := browserSanitizeLooseArgumentString(match[2])
	if label == "" {
		return nil
	}
	tag := strings.ToLower(strings.TrimSpace(match[1]))
	if tag == "*" {
		tag = ""
	}
	hint := &BrowserElementHint{
		Label: label,
		Tag:   tag,
	}
	hint.LocatorOrder = hint.EffectiveLocatorOrder()
	hint.LocatorPlan = hint.EffectiveLocatorPlan()
	hint.ResolutionMode = hint.EffectiveResolutionMode()
	return hint
}

func browserMergeLooseElementHints(primary *BrowserElementHint, secondary *BrowserElementHint) *BrowserElementHint {
	switch {
	case primary == nil:
		return secondary
	case secondary == nil:
		return primary
	}
	merged := *primary
	if strings.TrimSpace(merged.Label) == "" {
		merged.Label = strings.TrimSpace(secondary.Label)
	}
	if strings.TrimSpace(merged.Role) == "" {
		merged.Role = strings.TrimSpace(secondary.Role)
	}
	if strings.TrimSpace(merged.Tag) == "" {
		merged.Tag = strings.TrimSpace(secondary.Tag)
	}
	if strings.TrimSpace(merged.Type) == "" {
		merged.Type = strings.TrimSpace(secondary.Type)
	}
	if strings.TrimSpace(merged.Href) == "" {
		merged.Href = strings.TrimSpace(secondary.Href)
	}
	if strings.TrimSpace(merged.Placeholder) == "" {
		merged.Placeholder = strings.TrimSpace(secondary.Placeholder)
	}
	if strings.TrimSpace(merged.PageURL) == "" {
		merged.PageURL = strings.TrimSpace(secondary.PageURL)
	}
	if strings.TrimSpace(merged.PageOrigin) == "" {
		merged.PageOrigin = strings.TrimSpace(secondary.PageOrigin)
	}
	if strings.TrimSpace(merged.PagePath) == "" {
		merged.PagePath = strings.TrimSpace(secondary.PagePath)
	}
	if strings.TrimSpace(merged.PageTitle) == "" {
		merged.PageTitle = strings.TrimSpace(secondary.PageTitle)
	}
	if merged.TabIndex <= 0 {
		merged.TabIndex = secondary.TabIndex
	}
	if merged.SelectorIndex <= 0 {
		merged.SelectorIndex = secondary.SelectorIndex
	}
	if strings.TrimSpace(merged.FramePath) == "" {
		merged.FramePath = strings.TrimSpace(secondary.FramePath)
	}
	if strings.TrimSpace(merged.Selector) == "" {
		merged.Selector = strings.TrimSpace(secondary.Selector)
	}
	if strings.TrimSpace(merged.NativeRef) == "" {
		merged.NativeRef = strings.TrimSpace(secondary.NativeRef)
	}
	merged.LocatorOrder = merged.EffectiveLocatorOrder()
	merged.LocatorPlan = merged.EffectiveLocatorPlan()
	merged.ResolutionMode = merged.EffectiveResolutionMode()
	return &merged
}

func browserClickElementHintValue(params map[string]any) string {
	return firstNonEmpty(
		firstString(params, "element"),
		firstString(params, "text"),
		firstString(params, "label"),
	)
}

func browserClickElementHintForTarget(target browserElementTarget) *BrowserElementHint {
	hint := browserElementHintForTarget(target)
	if hint == nil {
		return nil
	}
	if !browserClickHintNeedsLocatorExpansion(target, hint) {
		return hint
	}
	label := strings.TrimSpace(hint.Label)
	if label == "" {
		return hint
	}
	expanded := *hint
	framePath := strings.TrimSpace(hint.FramePath)
	normalizedTag := strings.ToLower(strings.TrimSpace(hint.Tag))
	order := []string{"role_label", "tag_label", "label"}
	plan := make([]agentxbrowserruntime.BrowserLocatorCandidate, 0, 8)
	appendRoleLabel := func(role string) {
		if strings.TrimSpace(role) == "" {
			return
		}
		plan = append(plan, agentxbrowserruntime.BrowserLocatorCandidate{
			Kind:          "role_label",
			Role:          strings.TrimSpace(role),
			Label:         label,
			SelectorIndex: hint.SelectorIndex,
			FramePath:     framePath,
		})
	}
	appendTagLabel := func(tag string) {
		if strings.TrimSpace(tag) == "" {
			return
		}
		plan = append(plan, agentxbrowserruntime.BrowserLocatorCandidate{
			Kind:          "tag_label",
			Tag:           strings.TrimSpace(tag),
			Label:         label,
			SelectorIndex: hint.SelectorIndex,
			FramePath:     framePath,
		})
	}
	switch normalizedTag {
	case "a":
		appendRoleLabel("link")
		appendTagLabel("a")
	case "button":
		appendRoleLabel("button")
		appendTagLabel("button")
	default:
		appendRoleLabel("link")
		appendRoleLabel("button")
		appendTagLabel("a")
		appendTagLabel("button")
	}
	plan = append(plan, agentxbrowserruntime.BrowserLocatorCandidate{
		Kind:          "label",
		Label:         label,
		SelectorIndex: hint.SelectorIndex,
		FramePath:     framePath,
	})
	if strings.TrimSpace(hint.PageURL) != "" ||
		strings.TrimSpace(hint.PageOrigin) != "" ||
		strings.TrimSpace(hint.PagePath) != "" ||
		strings.TrimSpace(hint.PageTitle) != "" ||
		hint.TabIndex > 0 {
		order = append(order, "page_binding")
		plan = append(plan, agentxbrowserruntime.BrowserLocatorCandidate{
			Kind:       "page_binding",
			PageURL:    strings.TrimSpace(hint.PageURL),
			PageOrigin: strings.TrimSpace(hint.PageOrigin),
			PagePath:   strings.TrimSpace(hint.PagePath),
			PageTitle:  strings.TrimSpace(hint.PageTitle),
			TabIndex:   hint.TabIndex,
		})
	}
	expanded.LocatorOrder = order
	expanded.LocatorPlan = plan
	expanded.ResolutionMode = expanded.EffectiveResolutionMode()
	return &expanded
}

func browserClickHintNeedsLocatorExpansion(target browserElementTarget, hint *BrowserElementHint) bool {
	if hint == nil {
		return false
	}
	if strings.TrimSpace(hint.Label) == "" {
		return false
	}
	if strings.TrimSpace(target.Selector) != "" || strings.TrimSpace(target.Payload.Selector) != "" {
		return false
	}
	if strings.TrimSpace(hint.NativeRef) != "" ||
		strings.TrimSpace(hint.Role) != "" ||
		strings.TrimSpace(hint.Type) != "" ||
		strings.TrimSpace(hint.Href) != "" ||
		strings.TrimSpace(hint.Placeholder) != "" {
		return false
	}
	if tag := strings.ToLower(strings.TrimSpace(hint.Tag)); tag != "" {
		if tag != "a" && tag != "button" {
			return false
		}
		return true
	}
	if len(hint.LocatorPlan) == 0 {
		return true
	}
	if len(hint.LocatorPlan) != 1 {
		return false
	}
	switch strings.TrimSpace(hint.LocatorPlan[0].Kind) {
	case "label", "tag_label":
		return true
	default:
		return false
	}
}

func resolveBrowserActionElementTarget(selector string, ref string, element string) (browserElementTarget, error) {
	selector = browserSanitizeLooseArgumentString(selector)
	ref = browserSanitizeLooseArgumentString(ref)
	element = browserSanitizeLooseArgumentString(element)
	hint := browserLooseElementLabelHint(element)
	if parsed := browserLooseTextMatchSelectorHint(selector); parsed != nil {
		selector = ""
		hint = browserMergeLooseElementHints(hint, parsed)
	}
	if selector == "" && ref == "" && element != "" {
		switch {
		case browserElementRefHasKnownPrefix(element):
			ref = element
		case browserLooksLikeSelector(element):
			selector = element
		}
	}
	return resolveBrowserElementTargetWithHint(selector, ref, hint)
}

func browserElementRefForSnapshotElement(element BrowserSnapshotElement, pageURL string, pageTitle string) string {
	selector := strings.TrimSpace(element.Selector)
	payload := browserElementRefPayload{
		Selector:      selector,
		SelectorIndex: element.SelectorIndex,
		FramePath:     strings.TrimSpace(element.FramePath),
		NativeRef:     browserSnapshotElementNativeRef(element),
		Role:          strings.TrimSpace(element.Role),
		Tag:           strings.TrimSpace(element.Tag),
		Label:         strings.TrimSpace(element.Label),
		Type:          strings.TrimSpace(element.Type),
		Href:          strings.TrimSpace(element.Href),
		Placeholder:   strings.TrimSpace(element.Placeholder),
		PageTitle:     strings.TrimSpace(pageTitle),
	}
	payload.PageURL, payload.PageOrigin, payload.PagePath = browserElementPageBinding(pageURL)
	return browserElementRefForPayload(payload)
}

func browserElementRefForPayload(payload browserElementRefPayload) string {
	payload = browserSanitizeElementRefPayload(payload)
	if !browserElementRefPayloadHasActionableLocator(payload) {
		return ""
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		if payload.Selector != "" {
			return browserElementRefForSelector(payload.Selector)
		}
		return ""
	}
	return browserElementMetaRefPrefix + base64.RawURLEncoding.EncodeToString(encoded)
}

func browserSnapshotElementNativeRef(element BrowserSnapshotElement) string {
	ref := strings.TrimSpace(element.Ref)
	if ref == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(ref, browserElementMetaRefPrefix):
		payload, err := browserDecodeElementRef(ref)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(payload.NativeRef)
	case strings.HasPrefix(ref, browserElementRefPrefix):
		return ""
	default:
		return ref
	}
}

func browserNormalizeSnapshotElements(elements []BrowserSnapshotElement, pageURL string, pageTitle string) []BrowserSnapshotElement {
	if len(elements) == 0 {
		return nil
	}
	normalized := make([]BrowserSnapshotElement, 0, len(elements))
	for idx, element := range elements {
		if element.Index <= 0 {
			element.Index = idx + 1
		}
		if strings.TrimSpace(element.Selector) == "" && strings.TrimSpace(element.Ref) != "" {
			if selector, err := browserSelectorFromElementRef(element.Ref); err == nil {
				element.Selector = strings.TrimSpace(selector)
			}
		}
		if strings.TrimSpace(element.Selector) != "" || browserElementRefPayloadHasActionableLocator(browserElementRefPayload{
			Selector:    strings.TrimSpace(element.Selector),
			NativeRef:   browserSnapshotElementNativeRef(element),
			FramePath:   strings.TrimSpace(element.FramePath),
			Role:        strings.TrimSpace(element.Role),
			Tag:         strings.TrimSpace(element.Tag),
			Label:       strings.TrimSpace(element.Label),
			Type:        strings.TrimSpace(element.Type),
			Href:        strings.TrimSpace(element.Href),
			Placeholder: strings.TrimSpace(element.Placeholder),
		}) {
			// Snapshot output should always expose an agentx-canonical ref when enough
			// information exists to reconstruct one, even if the backend returned its
			// own native ref format.
			element.Ref = browserElementRefForSnapshotElement(element, pageURL, pageTitle)
		}
		normalized = append(normalized, element)
	}
	return normalized
}

func browserElementPageBinding(rawURL string) (string, string, string) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", "", ""
	}
	parsed, err := neturl.Parse(rawURL)
	if err != nil {
		return rawURL, "", ""
	}
	parsed.Fragment = ""
	path := strings.TrimSpace(parsed.EscapedPath())
	if path == "" {
		path = "/"
	}
	origin := ""
	if strings.TrimSpace(parsed.Scheme) != "" && strings.TrimSpace(parsed.Host) != "" {
		origin = strings.TrimSpace(parsed.Scheme + "://" + parsed.Host)
	}
	return parsed.String(), origin, path
}

func browserElementRefForSelector(selector string) string {
	selector = browserSanitizeLooseArgumentString(selector)
	if selector == "" {
		return ""
	}
	return browserElementRefPrefix + base64.RawURLEncoding.EncodeToString([]byte(selector))
}

func browserDecodeBase64Loose(payload string) ([]byte, error) {
	payload = browserSanitizeLooseArgumentString(payload)
	if payload == "" {
		return nil, nil
	}
	decoders := []func(string) ([]byte, error){
		base64.RawURLEncoding.DecodeString,
		base64.URLEncoding.DecodeString,
		base64.RawStdEncoding.DecodeString,
		base64.StdEncoding.DecodeString,
	}
	var lastErr error
	for _, decode := range decoders {
		decoded, err := decode(payload)
		if err == nil {
			return decoded, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func browserDecodeElementRef(ref string) (browserElementRefPayload, error) {
	ref = browserSanitizeLooseArgumentString(ref)
	if ref == "" {
		return browserElementRefPayload{}, nil
	}
	switch {
	case strings.HasPrefix(ref, browserElementMetaRefPrefix):
		payload := strings.TrimPrefix(ref, browserElementMetaRefPrefix)
		decoded, err := browserDecodeBase64Loose(payload)
		if err != nil {
			return browserElementRefPayload{}, fmt.Errorf("decode ref: %w", err)
		}
		var result browserElementRefPayload
		if err := json.Unmarshal(decoded, &result); err != nil {
			sanitized := browserStripControlCharacters(string(decoded))
			if sanitized == "" || json.Unmarshal([]byte(sanitized), &result) != nil {
				if recovered, ok := browserRecoverElementRefPayload(decoded); ok {
					return recovered, nil
				}
				return browserElementRefPayload{}, fmt.Errorf("decode ref payload: %w", err)
			}
		}
		result = browserSanitizeElementRefPayload(result)
		if !browserElementRefPayloadHasActionableLocator(result) {
			return browserElementRefPayload{}, fmt.Errorf("ref does not contain a usable locator")
		}
		return result, nil
	case strings.HasPrefix(ref, browserElementRefPrefix):
		selector, err := browserSelectorFromCSSRef(ref)
		if err != nil {
			return browserElementRefPayload{}, err
		}
		return browserElementRefPayload{Selector: selector}, nil
	default:
		return browserElementRefPayload{}, fmt.Errorf("ref must use %q or %q format", browserElementRefPrefix+"...", browserElementMetaRefPrefix+"...")
	}
}

func browserRecoverElementRefPayload(decoded []byte) (browserElementRefPayload, bool) {
	text := strings.ToValidUTF8(browserStripControlCharacters(string(decoded)), "")
	if strings.TrimSpace(text) == "" {
		return browserElementRefPayload{}, false
	}
	payload := browserElementRefPayload{
		Selector:      browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["selector"]...),
		NativeRef:     browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["native_ref"]...),
		Role:          browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["role"]...),
		Tag:           browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["tag"]...),
		Label:         browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["label"]...),
		Type:          browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["type"]...),
		Href:          browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["href"]...),
		Placeholder:   browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["placeholder"]...),
		PageURL:       browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["page_url"]...),
		PageOrigin:    browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["page_origin"]...),
		PagePath:      browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["page_path"]...),
		PageTitle:     browserRecoverElementRefStringField(text, browserElementRefStringFieldNames["page_title"]...),
		SelectorIndex: browserRecoverElementRefIntField(text, "selector_index"),
		TabIndex:      browserRecoverElementRefIntField(text, "tab_index"),
	}
	payload = browserSanitizeElementRefPayload(payload)
	return payload, browserElementRefPayloadHasActionableLocator(payload)
}

func browserRecoverElementRefStringField(raw string, keys ...string) string {
	for _, key := range keys {
		pattern := regexp.MustCompile(fmt.Sprintf(`(?i)"%s"\s*:\s*"((?:\\.|[^"\\])*)"`, regexp.QuoteMeta(key)))
		match := pattern.FindStringSubmatch(raw)
		if len(match) < 2 {
			continue
		}
		unquoted, err := strconv.Unquote(`"` + match[1] + `"`)
		if err != nil {
			return browserSanitizeLooseArgumentString(match[1])
		}
		return browserSanitizeLooseArgumentString(unquoted)
	}
	return ""
}

func browserRecoverElementRefIntField(raw string, key string) int {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i)"%s"\s*:\s*([0-9]+)`, regexp.QuoteMeta(key)))
	match := pattern.FindStringSubmatch(raw)
	if len(match) < 2 {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(match[1]))
	if err != nil {
		return 0
	}
	return value
}

func browserSelectorFromElementRef(ref string) (string, error) {
	payload, err := browserDecodeElementRef(ref)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.Selector), nil
}

func browserElementTargetHasActionableLocator(target browserElementTarget) bool {
	payload := target.Payload
	if strings.TrimSpace(target.Ref) != "" && !browserElementRefPayloadHasActionableLocator(payload) {
		if decoded, err := browserDecodeElementRef(target.Ref); err == nil {
			payload = decoded
		}
	}
	if strings.TrimSpace(target.Selector) != "" {
		payload.Selector = strings.TrimSpace(target.Selector)
	}
	return browserElementRefPayloadHasActionableLocator(payload)
}

func browserSelectorFromCSSRef(ref string) (string, error) {
	ref = browserSanitizeLooseArgumentString(ref)
	if ref == "" {
		return "", nil
	}
	if !strings.HasPrefix(ref, browserElementRefPrefix) {
		return "", fmt.Errorf("ref must use %q format", browserElementRefPrefix+"...")
	}
	payload := strings.TrimPrefix(ref, browserElementRefPrefix)
	decoded, err := browserDecodeBase64Loose(payload)
	if err != nil {
		return "", fmt.Errorf("decode ref: %w", err)
	}
	selector := browserSanitizeLooseArgumentString(string(decoded))
	if selector == "" {
		return "", fmt.Errorf("ref does not contain a selector")
	}
	return selector, nil
}

func browserElementRefHasPageBinding(payload browserElementRefPayload) bool {
	return strings.TrimSpace(payload.PageURL) != "" ||
		strings.TrimSpace(payload.PageOrigin) != "" ||
		strings.TrimSpace(payload.PagePath) != "" ||
		strings.TrimSpace(payload.PageTitle) != ""
}

func browserElementRefPageBindingSummary(payload browserElementRefPayload) string {
	if pageURL := strings.TrimSpace(payload.PageURL); pageURL != "" {
		return pageURL
	}
	parts := make([]string, 0, 3)
	if origin := strings.TrimSpace(payload.PageOrigin); origin != "" {
		parts = append(parts, origin)
	}
	if path := strings.TrimSpace(payload.PagePath); path != "" {
		parts = append(parts, path)
	}
	if title := strings.TrimSpace(payload.PageTitle); title != "" {
		parts = append(parts, title)
	}
	return strings.Join(parts, " ")
}

func browserCurrentPageBindingSummary(rawURL string, title string) string {
	pageURL, origin, path := browserElementPageBinding(rawURL)
	if pageURL != "" {
		return pageURL
	}
	parts := make([]string, 0, 3)
	if origin != "" {
		parts = append(parts, origin)
	}
	if path != "" {
		parts = append(parts, path)
	}
	if title = strings.TrimSpace(title); title != "" {
		parts = append(parts, title)
	}
	return strings.Join(parts, " ")
}

func browserElementRefMatchesCurrentPage(payload browserElementRefPayload, rawURL string, title string) bool {
	if !browserElementRefHasPageBinding(payload) {
		return true
	}
	currentURL, currentOrigin, currentPath := browserElementPageBinding(rawURL)
	targetURL := strings.TrimSpace(payload.PageURL)
	targetOrigin := strings.ToLower(strings.TrimSpace(payload.PageOrigin))
	targetPath := strings.TrimSpace(payload.PagePath)
	targetTitle := strings.ToLower(strings.TrimSpace(payload.PageTitle))
	currentOrigin = strings.ToLower(strings.TrimSpace(currentOrigin))
	currentPath = strings.TrimSpace(currentPath)
	currentTitle := strings.ToLower(strings.TrimSpace(title))

	if targetURL != "" && currentURL != "" {
		return targetURL == currentURL
	}
	pageOriginOk := targetOrigin == "" || targetOrigin == currentOrigin
	pagePathOk := targetPath == "" || targetPath == currentPath
	pageTitleOk := targetTitle == "" || targetTitle == currentTitle
	return (pageOriginOk && pagePathOk) || (pageOriginOk && pageTitleOk)
}

func browserResolveTrackedTargetForElementBinding(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, browserApp string, target browserToolTarget) (BrowserSessionTarget, bool) {
	if registry == nil {
		return BrowserSessionTarget{}, false
	}
	sessionID := ToolSessionIDFromContext(ctx)
	if sessionID == "" {
		return BrowserSessionTarget{}, false
	}
	route := browserSessionRoute(runtimeInfo, browserApp, "")
	tracked, ok := agentxbrowserruntime.ResolveSharedSessionBrowserTarget(
		registry,
		sessionID,
		route,
		target.TargetID,
		target.TabIndex,
		strings.TrimSpace(runtimeInfo.Profile) == "",
	)
	if !ok {
		return BrowserSessionTarget{}, false
	}
	if hiddenImplicitHostDefaultBase &&
		strings.TrimSpace(target.TargetID) == "" &&
		target.TabIndex <= 0 {
		trackedTarget := strings.ToLower(strings.TrimSpace(tracked.Target))
		if trackedTarget == "" || trackedTarget == "host" {
			return BrowserSessionTarget{}, false
		}
	}
	return tracked, true
}

func browserValidateElementTargetPageBinding(ctx context.Context, registry *BrowserSessionRegistry, runtimeInfo BrowserRuntimeInfo, hiddenImplicitHostDefaultBase bool, browserApp string, target browserToolTarget, requestURL string, elementTarget browserElementTarget) error {
	if strings.TrimSpace(elementTarget.Ref) == "" {
		return nil
	}
	payload := elementTarget.Payload
	if !browserElementRefHasPageBinding(payload) {
		decoded, err := browserDecodeElementRef(elementTarget.Ref)
		if err == nil {
			payload = decoded
		}
	}
	if !browserElementRefHasPageBinding(payload) {
		return nil
	}
	if requestURL = strings.TrimSpace(requestURL); requestURL != "" {
		if browserElementRefMatchesCurrentPage(payload, requestURL, "") {
			return nil
		}
		expected := firstNonEmpty(browserElementRefPageBindingSummary(payload), "the original page")
		current := firstNonEmpty(browserCurrentPageBindingSummary(requestURL, ""), requestURL)
		return fmt.Errorf("element ref page binding differs from requested page: expected %s but requested page is %s", expected, current)
	}
	tracked, ok := browserResolveTrackedTargetForElementBinding(ctx, registry, runtimeInfo, hiddenImplicitHostDefaultBase, browserApp, target)
	if !ok {
		return nil
	}
	if browserElementRefMatchesCurrentPage(payload, tracked.URL, tracked.Title) {
		return nil
	}
	expected := firstNonEmpty(browserElementRefPageBindingSummary(payload), "the original page")
	current := firstNonEmpty(browserCurrentPageBindingSummary(tracked.URL, tracked.Title), "unknown")
	return fmt.Errorf("element ref page binding differs from current target: expected %s but current page is %s", expected, current)
}

func browserElementHintFromPayload(payload browserElementRefPayload) *BrowserElementHint {
	payload.Selector = strings.TrimSpace(payload.Selector)
	payload.NativeRef = strings.TrimSpace(payload.NativeRef)
	payload.FramePath = strings.TrimSpace(payload.FramePath)
	payload.Role = strings.TrimSpace(payload.Role)
	payload.Tag = strings.TrimSpace(payload.Tag)
	payload.Label = strings.TrimSpace(payload.Label)
	payload.Type = strings.TrimSpace(payload.Type)
	payload.Href = strings.TrimSpace(payload.Href)
	payload.Placeholder = strings.TrimSpace(payload.Placeholder)
	payload.PageURL = strings.TrimSpace(payload.PageURL)
	payload.PageOrigin = strings.TrimSpace(payload.PageOrigin)
	payload.PagePath = strings.TrimSpace(payload.PagePath)
	payload.PageTitle = strings.TrimSpace(payload.PageTitle)
	if payload.Selector == "" &&
		payload.SelectorIndex <= 0 &&
		payload.NativeRef == "" &&
		payload.FramePath == "" &&
		payload.Role == "" &&
		payload.Tag == "" &&
		payload.Label == "" &&
		payload.Type == "" &&
		payload.Href == "" &&
		payload.Placeholder == "" &&
		payload.PageURL == "" &&
		payload.PageOrigin == "" &&
		payload.PagePath == "" &&
		payload.PageTitle == "" &&
		payload.TabIndex <= 0 {
		return nil
	}
	if payload.Role == "" &&
		payload.SelectorIndex <= 0 &&
		payload.NativeRef == "" &&
		payload.FramePath == "" &&
		payload.Tag == "" &&
		payload.Label == "" &&
		payload.Type == "" &&
		payload.Href == "" &&
		payload.Placeholder == "" &&
		payload.PageURL == "" &&
		payload.PageOrigin == "" &&
		payload.PagePath == "" &&
		payload.PageTitle == "" &&
		payload.TabIndex <= 0 {
		return nil
	}
	hint := &BrowserElementHint{
		Selector:      payload.Selector,
		SelectorIndex: payload.SelectorIndex,
		FramePath:     payload.FramePath,
		NativeRef:     payload.NativeRef,
		Role:          payload.Role,
		Tag:           payload.Tag,
		Label:         payload.Label,
		Type:          payload.Type,
		Href:          payload.Href,
		Placeholder:   payload.Placeholder,
		PageURL:       payload.PageURL,
		PageOrigin:    payload.PageOrigin,
		PagePath:      payload.PagePath,
		PageTitle:     payload.PageTitle,
		TabIndex:      payload.TabIndex,
	}
	hint.LocatorOrder = hint.EffectiveLocatorOrder()
	hint.LocatorPlan = hint.EffectiveLocatorPlan()
	hint.ResolutionMode = hint.EffectiveResolutionMode()
	return hint
}

func browserElementHintForTarget(target browserElementTarget) *BrowserElementHint {
	payload := target.Payload
	if strings.TrimSpace(target.Ref) != "" && payload.Selector == "" {
		if decoded, err := browserDecodeElementRef(target.Ref); err == nil {
			payload = decoded
		}
	}
	if payload.Selector == "" {
		payload.Selector = strings.TrimSpace(target.Selector)
	}
	return browserElementHintFromPayload(payload)
}

func browserRemoteLocatorProjectionForTarget(target browserElementTarget) (string, string, *BrowserElementHint, *agentxbrowserruntime.BrowserElementResolverRequest, agentxbrowserruntime.BrowserElementRemoteProjection) {
	hint := browserElementHintForTarget(target)
	if hint == nil {
		projection := agentxbrowserruntime.BrowserElementRemoteProjection{
			ResolutionMode: "selector_first",
			PrimaryKind:    "selector",
			Selector:       strings.TrimSpace(target.Selector),
		}
		resolver := &agentxbrowserruntime.BrowserElementResolverRequest{
			ResolutionMode: projection.ResolutionMode,
			PrimaryKind:    projection.PrimaryKind,
			Selector:       projection.Selector,
			LocatorOrder:   []string{"selector"},
			LocatorPlan:    []agentxbrowserruntime.BrowserLocatorCandidate{{Kind: "selector", Selector: projection.Selector}},
		}
		return "", strings.TrimSpace(target.Selector), nil, resolver, projection
	}
	projection := hint.RemoteProjection()
	return projection.ElementRef, projection.Selector, hint.RemoteHint(), hint.RemoteResolver(), projection
}

func browserRemoteElementRefForTarget(target browserElementTarget) string {
	elementRef, _, _, _, _ := browserRemoteLocatorProjectionForTarget(target)
	return elementRef
}

func browserElementRefHasKnownPrefix(ref string) bool {
	ref = strings.TrimSpace(ref)
	return strings.HasPrefix(ref, browserElementMetaRefPrefix) || strings.HasPrefix(ref, browserElementRefPrefix)
}

func browserResolveRemoteElementRef(selector string, ref string) (string, string, error) {
	if strings.TrimSpace(selector) == "" && strings.TrimSpace(ref) == "" {
		return "", "", nil
	}
	target, err := resolveBrowserElementTarget(selector, ref)
	if err != nil {
		return "", "", err
	}
	return browserRemoteElementRefForTarget(target), target.Selector, nil
}

func browserPreferredResultRef(inputRef string, resultRef string) string {
	if input := strings.TrimSpace(inputRef); input != "" {
		return input
	}
	return strings.TrimSpace(resultRef)
}

func browserElementResolverHint(target browserElementTarget) *BrowserElementHint {
	payload := target.Payload
	payload.Selector = firstNonEmpty(strings.TrimSpace(payload.Selector), strings.TrimSpace(target.Selector))
	hint := browserElementHintFromPayload(payload)
	if hint == nil {
		return &BrowserElementHint{}
	}
	return hint
}

func browserElementResolverRequestForTarget(target browserElementTarget) *agentxbrowserruntime.BrowserElementResolverRequest {
	hint := browserElementResolverHint(target)
	if hint == nil {
		return nil
	}
	return hint.RemoteResolver()
}

func browserElementResolverJS(target browserElementTarget) string {
	encoded, _ := json.Marshal(browserElementResolverRequestForTarget(target))
	return `(function(){const target=` + string(encoded) + `||{};const normalize=(value)=>String(value==null?"":value).replace(/\s+/g," ").trim();const lower=(value)=>normalize(value).toLowerCase();const attr=(el,name)=>normalize(el&&el.getAttribute?el.getAttribute(name):"");const parseURL=(value)=>{const raw=normalize(value);if(!raw){return null;}try{return new URL(raw,window.location.href);}catch(_){return null;}};const comparableURL=(parsed)=>{if(!parsed){return "";}const pathname=normalize(parsed.pathname)||"/";const search=normalize(parsed.search);return normalize(parsed.origin+pathname+search);};const currentUrl=parseURL(window.location.href);const currentTitle=normalize(document.title||"");const locatorOrder=(Array.isArray(target.locator_order)?target.locator_order:[]).map((kind)=>lower(kind)).filter(Boolean);const locatorPlan=Array.isArray(target.locator_plan)&&target.locator_plan.length>0?target.locator_plan:[];const matchPlan=Array.isArray(target.match_plan)&&target.match_plan.length>0?target.match_plan:locatorPlan.filter((candidate)=>lower(candidate&&candidate.kind)!=='page_binding');const pageBinding=(target.page_binding&&typeof target.page_binding==='object'?target.page_binding:(locatorPlan.find((candidate)=>lower(candidate&&candidate.kind)==='page_binding')||{}));const targetPageUrl=parseURL(pageBinding.page_url);const targetPageOrigin=lower(pageBinding.page_origin||(targetPageUrl?targetPageUrl.origin:""));const targetPagePath=normalize(pageBinding.page_path||(targetPageUrl?targetPageUrl.pathname:""));const targetPageTitle=lower(pageBinding.page_title);const pageHasBinding=Boolean(targetPageUrl||targetPageOrigin||targetPagePath||targetPageTitle);const currentOrigin=lower(currentUrl?currentUrl.origin:"");const currentPath=normalize(currentUrl?currentUrl.pathname:"");const currentComparable=comparableURL(currentUrl);const targetComparable=comparableURL(targetPageUrl);const pageOriginOk=!targetPageOrigin||targetPageOrigin===currentOrigin;const pagePathOk=!targetPagePath||targetPagePath===currentPath;const pageTitleOk=!targetPageTitle||targetPageTitle===lower(currentTitle);let pageOk=!pageHasBinding;if(!pageOk){if(targetPageUrl&&currentUrl){pageOk=targetComparable!==""&&currentComparable===targetComparable;}else{pageOk=(pageOriginOk&&pagePathOk)||(pageOriginOk&&pageTitleOk);}}let pageError="";if(pageHasBinding&&!pageOk){const expected=normalize(targetComparable||[targetPageOrigin,targetPagePath,targetPageTitle].filter(Boolean).join(" "));const current=normalize(currentComparable||[currentOrigin,currentPath,lower(currentTitle)].filter(Boolean).join(" "));pageError='page_changed: element ref expects '+(expected||'the original page')+' but current page is '+(current||'unknown');}const readRole=(el)=>{const explicit=lower(attr(el,"role"));if(explicit){return explicit;}const tag=lower(el&&el.tagName);switch(tag){case "a":return "link";case "button":return "button";case "textarea":return "textarea";case "select":return "select";case "summary":return "summary";case "input":{const inputType=lower(attr(el,"type"));switch(inputType){case "submit":case "button":case "reset":case "image":return "button";case "checkbox":case "radio":case "file":return inputType;default:return "input";}}}if(lower(attr(el,"contenteditable"))==="true"){return "editable";}return "";};const readLabel=(el)=>{if(!el){return "";}const labelledBy=attr(el,"aria-labelledby");if(labelledBy){const pieces=[];labelledBy.split(/\s+/).forEach((id)=>{const node=document.getElementById(id);if(node){const text=normalize(node.textContent||node.innerText||"");if(text){pieces.push(text);}}});if(pieces.length>0){return pieces.join(" ");}}const parentLabel=typeof el.closest==="function"?el.closest("label"):null;return normalize(attr(el,"aria-label")||(parentLabel&&(parentLabel.textContent||parentLabel.innerText))||attr(el,"value")||attr(el,"placeholder")||el.innerText||el.textContent||attr(el,"name")||attr(el,"title")||attr(el,"alt"));};const interactiveSelector='a[href],button,input,textarea,select,summary,[role],[contenteditable="true"],[tabindex]';const locatorWeight=(kind)=>{const idx=locatorOrder.indexOf(lower(kind));if(idx<0){return 1;}return Math.max(1,(locatorOrder.length-idx))*8;};if(pageError){return {element:null,error:pageError,note:'element ref page binding differs from current page',page_ok:false,matched_by:''};}const candidateStates=new Map();const addCandidate=(node,kind)=>{if(!node||!node.tagName){return;}let state=candidateStates.get(node);if(!state){state={element:node,kinds:[]};candidateStates.set(node,state);}const normalizedKind=lower(kind);if(normalizedKind && state.kinds.indexOf(normalizedKind)<0){state.kinds.push(normalizedKind);}};const eachQuery=(selector,visit)=>{const query=normalize(selector);if(!query){return;}try{document.querySelectorAll(query).forEach((node)=>visit(node));}catch(_){}};for(const candidate of matchPlan){const kind=lower(candidate&&candidate.kind);if(!kind){continue;}switch(kind){case 'native_ref':break;case 'selector':eachQuery(candidate&&candidate.selector,(node)=>addCandidate(node,kind));break;case 'href':{const targetHref=normalize(candidate&&candidate.href);if(!targetHref){break;}document.querySelectorAll('a[href]').forEach((el)=>{const href=normalize(attr(el,'href')||(typeof el.href==="string"?el.href:""));if(href===targetHref){addCandidate(el,kind);}});break;}case 'role_label':{const targetRole=lower(candidate&&candidate.role);const targetLabel=lower(candidate&&candidate.label);if(!targetRole||!targetLabel){break;}document.querySelectorAll(interactiveSelector).forEach((el)=>{if(readRole(el)===targetRole && lower(readLabel(el))===targetLabel){addCandidate(el,kind);}});break;}case 'tag_label':{const targetTag=lower(candidate&&candidate.tag);const targetLabel=lower(candidate&&candidate.label);if(!targetTag||!targetLabel){break;}eachQuery(targetTag,(el)=>{if(lower(readLabel(el))===targetLabel){addCandidate(el,kind);}});break;}case 'label':{const targetLabel=lower(candidate&&candidate.label);if(!targetLabel){break;}document.querySelectorAll(interactiveSelector).forEach((el)=>{if(lower(readLabel(el))===targetLabel){addCandidate(el,kind);}});break;}case 'placeholder':{const targetPlaceholder=lower(candidate&&candidate.placeholder);if(!targetPlaceholder){break;}document.querySelectorAll('input[placeholder],textarea[placeholder]').forEach((el)=>{if(lower(attr(el,'placeholder'))===targetPlaceholder){addCandidate(el,kind);}});break;}case 'tag_type':{const targetTag=lower(candidate&&candidate.tag);const targetType=lower(candidate&&candidate.type);if(!targetTag||!targetType){break;}eachQuery(targetTag,(el)=>{if(lower(attr(el,'type'))===targetType){addCandidate(el,kind);}});break;}case 'tag':{const targetTag=lower(candidate&&candidate.tag);if(!targetTag){break;}eachQuery(targetTag,(el)=>addCandidate(el,kind));break;}case 'type':{const targetType=lower(candidate&&candidate.type);if(!targetType){break;}document.querySelectorAll('[type]').forEach((el)=>{if(lower(attr(el,'type'))===targetType){addCandidate(el,kind);}});break;}}}let best=null;let bestScore=-1;let bestMatchedBy='';for(const state of candidateStates.values()){const matchedKinds=Array.isArray(state.kinds)?state.kinds:[];let score=0;for(const kind of matchedKinds){score+=locatorWeight(kind);}if(score>bestScore){best=state.element;bestScore=score;bestMatchedBy=matchedKinds.join('+');}}if(best&&bestScore>0){const notes=[];if(bestMatchedBy && bestMatchedBy!=='selector'){notes.push('resolved via '+bestMatchedBy);}return {element:best,note:notes.join('; '),page_ok:true,matched_by:bestMatchedBy};}const unresolvedNotes=[];if(locatorOrder.indexOf('native_ref')>=0 || lower(target.primary_kind)==='native_ref' || normalize(target.element_ref)!==''){unresolvedNotes.push('native_ref requires remote resolution');}if(pageHasBinding){unresolvedNotes.push('page binding matched but no local element resolved');}return {element:null,note:unresolvedNotes.join('; '),page_ok:true,matched_by:''};})()`
}
