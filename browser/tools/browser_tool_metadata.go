package tools

import (
	"strings"

	types "github.com/wsnacj/agentx-go/components/llm"
)

func browserToolMetadata(defs []types.Tool, decorate func(string, ToolMetadata) ToolMetadata) map[string]ToolMetadata {
	return browserToolMetadataFiltered(defs, nil, decorate)
}

func browserToolMetadataFiltered(defs []types.Tool, include func(string) bool, decorate func(string, ToolMetadata) ToolMetadata) map[string]ToolMetadata {
	if len(defs) == 0 {
		return nil
	}
	out := map[string]ToolMetadata{}
	for _, def := range defs {
		name := NormalizeToolName(def.Function.Name)
		if name == "" || !isBrowserToolName(name) {
			continue
		}
		if include != nil && !include(name) {
			continue
		}
		meta := browserToolMetadataForDefinition(name, def)
		if meta.Source == "" && (isBrowserBuiltinToolName(name) || IsBrowserCompatToolName(name)) {
			meta.Source = ToolSourceBuiltin
		}
		if decorate != nil {
			meta = decorate(name, meta)
		}
		out[name] = meta
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func BrowserToolMetadata(defs []types.Tool, runtimeInfo BrowserRuntimeInfo) map[string]ToolMetadata {
	return browserToolMetadata(defs, func(_ string, meta ToolMetadata) ToolMetadata {
		return applyBrowserRuntimeInfoMetadata(meta, runtimeInfo)
	})
}

func BrowserToolMetadataForOptions(defs []types.Tool, opts BrowserToolOptions) map[string]ToolMetadata {
	preview := browserDefaultRuntimePreviewForToolOptions(opts)
	managedCompatSurfaces := browserManagedOptInMetadataSurfacesForOptions(opts, preview)
	substrateHints := browserSubstrateMetadataRouteHints(preview)
	runtimeSelectionHints := substrateHints
	runtimeSelectionHints.RuntimeInfo = browserVisibleDefaultRuntimeInfoForPreview(preview)
	runtimeInfo := runtimeSelectionHints.PreferredRuntimeInfo()
	return browserToolMetadata(defs, func(name string, meta ToolMetadata) ToolMetadata {
		if surface, ok := managedCompatSurfaces[name]; ok {
			if name == "browser" && len(surface.Kinds) != 0 {
				return ApplyBrowserToolMetadataRouteHints(browserToolMetadataApplyManagedUnifiedSurface(meta, surface), substrateHints)
			}
			if name == "browser_act" && len(surface.Kinds) != 0 {
				return ApplyBrowserToolMetadataRouteHints(browserToolMetadataApplyManagedActSurface(meta, surface), substrateHints)
			}
			return ApplyBrowserToolMetadataRouteHints(browserToolMetadataApplyManagedCompatSurface(meta, surface), substrateHints)
		}
		return ApplyBrowserToolMetadataRouteHints(applyBrowserRuntimeInfoMetadata(meta, runtimeInfo), substrateHints)
	})
}

func inferBrowserToolMetadata(defs []types.Tool) map[string]ToolMetadata {
	return browserToolMetadata(defs, nil)
}

func inferBrowserToolMetadataMissing(defs []types.Tool, provided map[string]ToolMetadata) map[string]ToolMetadata {
	covered := browserCompleteProvidedMetadataNames(provided)
	if len(covered) == 0 {
		return inferBrowserToolMetadata(defs)
	}
	return browserToolMetadataFiltered(defs, func(name string) bool {
		return !covered[name]
	}, nil)
}

func browserCompleteProvidedMetadataNames(items map[string]ToolMetadata) map[string]bool {
	if len(items) == 0 {
		return nil
	}
	out := map[string]bool{}
	for rawName, meta := range items {
		name := NormalizeToolName(rawName)
		if name == "" || !isBrowserToolName(name) {
			continue
		}
		if !browserMetadataLooksComplete(meta) {
			continue
		}
		out[name] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func browserMetadataLooksComplete(meta ToolMetadata) bool {
	return strings.EqualFold(strings.TrimSpace(meta.Type), "browser") &&
		len(meta.Groups) != 0 &&
		len(meta.Capabilities) != 0
}

func browserActKindsFromToolDefinition(def types.Tool) []string {
	properties, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok || len(properties) == 0 {
		return nil
	}
	kindDef, ok := properties["kind"].(map[string]any)
	if !ok || len(kindDef) == 0 {
		return nil
	}
	if rawKinds, ok := kindDef["enum"].([]string); ok {
		return mergeToolMetadataStrings(nil, rawKinds)
	}
	rawKinds, ok := kindDef["enum"].([]any)
	if !ok || len(rawKinds) == 0 {
		return nil
	}
	kinds := make([]string, 0, len(rawKinds))
	for _, item := range rawKinds {
		text, ok := item.(string)
		if !ok {
			continue
		}
		kinds = append(kinds, text)
	}
	return mergeToolMetadataStrings(nil, kinds)
}

func browserRuntimeActionsFromToolDefinition(def types.Tool) []string {
	properties, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok || len(properties) == 0 {
		return []string{"status"}
	}
	actionDef, ok := properties["action"].(map[string]any)
	if !ok || len(actionDef) == 0 {
		return []string{"status"}
	}
	if rawActions, ok := actionDef["enum"].([]string); ok && len(rawActions) > 0 {
		return mergeToolMetadataStrings(nil, rawActions)
	}
	rawActions, ok := actionDef["enum"].([]any)
	if !ok || len(rawActions) == 0 {
		return []string{"status"}
	}
	actions := make([]string, 0, len(rawActions))
	for _, item := range rawActions {
		text, ok := item.(string)
		if !ok {
			continue
		}
		actions = append(actions, text)
	}
	if len(actions) == 0 {
		return []string{"status"}
	}
	return mergeToolMetadataStrings(nil, actions)
}

func isBrowserToolName(name string) bool {
	return IsBrowserUnifiedToolName(name) || IsBrowserSpecialistToolName(name) || IsBrowserCompatToolName(name)
}

func browserUnifiedActionsFromToolDefinition(def types.Tool) ([]string, []string) {
	properties, ok := def.Function.Parameters["properties"].(map[string]any)
	if !ok || len(properties) == 0 {
		return nil, nil
	}
	actionDef, ok := properties["action"].(map[string]any)
	if !ok || len(actionDef) == 0 {
		return nil, browserActKindsFromToolDefinition(def)
	}
	actKinds := browserActKindsFromToolDefinition(def)
	rawActions, _ := actionDef["enum"].([]any)
	if len(rawActions) == 0 {
		if stringsActions, ok := actionDef["enum"].([]string); ok {
			rawActions = make([]any, 0, len(stringsActions))
			for _, action := range stringsActions {
				rawActions = append(rawActions, action)
			}
		}
	}
	actSet := map[string]bool{"act": true}
	for _, kind := range actKinds {
		actSet[strings.ToLower(strings.TrimSpace(kind))] = true
	}
	for alias, kind := range browserUnifiedActionAliases {
		if actSet[strings.ToLower(strings.TrimSpace(kind))] {
			actSet[strings.ToLower(strings.TrimSpace(alias))] = true
		}
	}
	runtimeActions := make([]string, 0, len(rawActions))
	for _, item := range rawActions {
		action, ok := item.(string)
		if !ok {
			continue
		}
		normalized := strings.ToLower(strings.TrimSpace(action))
		if normalized == "" || actSet[normalized] {
			continue
		}
		runtimeActions = append(runtimeActions, normalized)
	}
	return mergeToolMetadataStrings(nil, runtimeActions), actKinds
}

func browserToolMetadataForDefinition(name string, def types.Tool) ToolMetadata {
	meta := ToolMetadata{
		Type: "browser",
	}
	switch name {
	case "browser":
		meta.Groups = []string{"browser_unified"}
		meta.ReadOnly = builtinToolMetadataBoolPtr(false)
		meta.ConcurrencySafe = builtinToolMetadataBoolPtr(false)
		runtimeActions, actKinds := browserUnifiedActionsFromToolDefinition(def)
		meta.Capabilities = append(
			[]string{"browser", "browser_workbench", "browser_runtime", "browser_act"},
			browserActMetadataCapabilities(actKinds)...,
		)
		meta.AuditTags = []string{"browser", "interactive_browser", "external_content"}
		for _, action := range runtimeActions {
			switch action {
			case "inspect":
				meta.Capabilities = append(meta.Capabilities, "read")
			case "status", "profiles", "sessions":
				meta.Capabilities = append(meta.Capabilities, "read")
			case "workbench":
				meta.Capabilities = append(meta.Capabilities, "read", "browser_workbench", "browser_session_inspect")
			case "ready":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_prepare")
			case "inventory":
				meta.Capabilities = append(meta.Capabilities, "read")
			case "handles":
				meta.Capabilities = append(meta.Capabilities, "read", "browser_session_inspect")
			case "prepare":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_prepare")
			case "repair":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_bootstrap_repair")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "coordinate", "ensure", "refresh", "sync", "teardown":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_session_coordination")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "reset":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_clear")
			case "adopt":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_sync")
			case "start", "restart", "stop":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "launch", "halt":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "new_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_create")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "remove_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_delete")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "create_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_create")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "delete_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_delete")
				meta.AuditTags = append(meta.AuditTags, "side_effect")
			case "select_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_select")
			case "pin_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_select")
			case "clear_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_clear")
			case "unpin_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_clear")
			case "clear_session":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_clear")
			case "sync_session":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_sync")
			case "select_target":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_target_control", "browser_target_select")
			case "pin_target":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_target_control", "browser_target_select")
			case "clear_target":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_target_control", "browser_target_clear")
			case "unpin_target":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_target_control", "browser_target_clear")
			}
			meta.Capabilities = append(meta.Capabilities, "browser_kind:runtime_"+strings.ToLower(strings.TrimSpace(action)))
		}
		meta.Capabilities = mergeToolMetadataStrings(nil, meta.Capabilities)
		meta.AuditTags = mergeToolMetadataStrings(nil, meta.AuditTags)
	case "browser_runtime":
		actions := browserRuntimeActionsFromToolDefinition(def)
		meta.Groups = []string{"browser_specialist"}
		meta.ReadOnly = builtinToolMetadataBoolPtr(false)
		meta.ConcurrencySafe = builtinToolMetadataBoolPtr(false)
		meta.Capabilities = []string{"browser", "read", "browser_runtime", "browser_kind:runtime"}
		meta.AuditTags = []string{"browser"}
		for _, action := range actions {
			switch action {
			case "status", "profiles":
				meta.Capabilities = append(meta.Capabilities, "read")
			case "workbench":
				meta.Capabilities = append(meta.Capabilities, "read", "browser_workbench", "browser_session_inspect")
			case "prepare":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_prepare")
			case "repair":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_bootstrap_repair")
				meta.AuditTags = append(meta.AuditTags, "interactive_browser", "side_effect")
			case "coordinate":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_session_coordination")
				meta.AuditTags = append(meta.AuditTags, "interactive_browser", "side_effect")
			case "start", "restart", "stop":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control")
				meta.AuditTags = append(meta.AuditTags, "interactive_browser", "side_effect")
			case "create_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_create")
				meta.AuditTags = append(meta.AuditTags, "interactive_browser", "side_effect")
			case "delete_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_delete")
				meta.AuditTags = append(meta.AuditTags, "interactive_browser", "side_effect")
			case "select_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_select")
			case "clear_profile":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_profile_clear")
			case "clear_session":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_clear")
			case "sync_session":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_sync")
			case "select_target":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_target_control", "browser_target_select")
			case "clear_target":
				meta.Capabilities = append(meta.Capabilities, "write", "browser_target_control", "browser_target_clear")
			}
			meta.Capabilities = append(meta.Capabilities, "browser_kind:runtime_"+strings.ToLower(strings.TrimSpace(action)))
		}
		meta.Capabilities = mergeToolMetadataStrings(nil, meta.Capabilities)
		meta.AuditTags = mergeToolMetadataStrings(nil, meta.AuditTags)
	case "browser_act":
		meta.Groups = []string{"browser_specialist"}
		meta.ReadOnly = builtinToolMetadataBoolPtr(false)
		meta.ConcurrencySafe = builtinToolMetadataBoolPtr(false)
		meta.Capabilities = browserActMetadataCapabilities(browserActKindsFromToolDefinition(def))
		meta.AuditTags = []string{"browser", "interactive_browser", "external_content"}
	}
	if compatMeta, ok := browserCompatMetadataForTool(name); ok {
		meta = mergeToolMetadataItem(meta, compatMeta)
	}
	if len(meta.Groups) > 0 {
		meta.Groups = mergeToolMetadataStrings(nil, meta.Groups)
	}
	if strings.TrimSpace(meta.RiskProfile) == "" {
		if level, ok := BrowserBuiltinRiskLevel(name); ok && level != RiskUnknown {
			meta.RiskProfile = level.String()
		}
	}
	meta.Capabilities = mergeToolMetadataStrings(nil, meta.Capabilities)
	meta.AuditTags = mergeToolMetadataStrings(nil, meta.AuditTags)
	return meta
}

func isBrowserLegacyCompatTool(name string) bool {
	return IsBrowserCompatToolName(name)
}

func browserActMetadataCapabilities(kinds []string) []string {
	capabilities := []string{"browser", "browser_act"}
	for _, kind := range kinds {
		switch strings.ToLower(strings.TrimSpace(kind)) {
		case "open", "navigate":
			capabilities = append(capabilities, "network")
		case "extract", "snapshot":
			capabilities = append(capabilities, "network", "read")
		case "screenshot":
			capabilities = append(capabilities, "read", "screenshot", "artifact_output", "artifact_contract:"+strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:screenshot")
		case "console":
			capabilities = append(capabilities, "read", "console")
		case "requests":
			capabilities = append(capabilities, "read", "network", "requests")
		case "response_body":
			capabilities = append(capabilities, "read", "network", "response_body")
		case "errors":
			capabilities = append(capabilities, "read", "errors")
		case "cookies":
			capabilities = append(capabilities, "read", "state", "cookies")
		case "cookies_set", "cookies_clear":
			capabilities = append(capabilities, "write", "state", "cookies")
		case "storage":
			capabilities = append(capabilities, "read", "state", "storage")
		case "storage_set", "storage_clear":
			capabilities = append(capabilities, "write", "state", "storage")
		case "offline":
			capabilities = append(capabilities, "write", "settings", "network")
		case "headers":
			capabilities = append(capabilities, "write", "settings", "network", "headers")
		case "credentials":
			capabilities = append(capabilities, "write", "settings", "auth")
		case "geolocation":
			capabilities = append(capabilities, "write", "settings", "geolocation")
		case "media":
			capabilities = append(capabilities, "write", "settings", "media")
		case "timezone", "locale":
			capabilities = append(capabilities, "write", "settings", "emulation")
		case "device":
			capabilities = append(capabilities, "write", "settings", "emulation", "viewport")
		case "highlight":
			capabilities = append(capabilities, "write", "dom", "highlight")
		case "trace_start":
			capabilities = append(capabilities, "read", "trace")
		case "trace_stop":
			capabilities = append(capabilities, "read", "trace", "artifact_output", "artifact_contract:"+strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:trace")
		case "download":
			capabilities = append(capabilities, "read", "artifact_output", "artifact_contract:"+strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:download")
		case "wait_download":
			capabilities = append(capabilities, "read", "artifact_output", "artifact_contract:"+strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:download")
		case "save_pdf":
			capabilities = append(capabilities, "read", "artifact_output", "artifact_contract:"+strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:pdf")
		case "save_html":
			capabilities = append(capabilities, "read", "artifact_output", "artifact_contract:"+strings.ReplaceAll(browserArtifactContract, "+", "_"), "artifact_kind:html")
		case "dialog":
			capabilities = append(capabilities, "write", "dom", "modal")
		case "upload":
			capabilities = append(capabilities, "write", "dom", "file_input")
		case "press":
			capabilities = append(capabilities, "write", "keyboard")
		case "hover", "drag":
			capabilities = append(capabilities, "write", "mouse")
		case "select", "fill":
			capabilities = append(capabilities, "write", "dom", "form")
		case "resize":
			capabilities = append(capabilities, "write", "viewport")
		case "click", "type":
			capabilities = append(capabilities, "write", "dom")
		case "evaluate":
			capabilities = append(capabilities, "exec", "dom")
		case "list_tabs", "focus_tab", "close_tab":
			capabilities = append(capabilities, "tabs")
		}
		capabilities = append(capabilities, "browser_kind:"+strings.ToLower(strings.TrimSpace(kind)))
	}
	return mergeToolMetadataStrings(nil, capabilities)
}

func applyBrowserRuntimeInfoMetadata(meta ToolMetadata, runtimeInfo BrowserRuntimeInfo) ToolMetadata {
	return ApplyBrowserToolMetadataRouteHints(meta, BrowserToolMetadataRouteHints{RuntimeInfo: runtimeInfo})
}

func browserSubstrateMetadataRouteHints(preview browserDefaultRuntimePreview) BrowserToolMetadataRouteHints {
	summary := preview.SubstrateSummary
	defaultRoute := normalizeBrowserRuntimeInfo(summary.DefaultRoute)
	defaultCandidateRoute := normalizeBrowserRuntimeInfo(summary.DefaultCandidateRoute)
	hiddenDefaultCandidate := defaultCandidateRoute != (BrowserRuntimeInfo{}) &&
		(defaultRoute == (BrowserRuntimeInfo{}) ||
			browserTopLevelShouldPreferDefaultCandidateRoute(
				defaultRoute.Backend,
				defaultRoute.Target,
				defaultCandidateRoute.Backend,
				defaultCandidateRoute.Target,
				summary.SelectionStrategy,
			))
	if hiddenDefaultCandidate {
		defaultCandidateRoute = normalizeBrowserRuntimeInfo(summary.DefaultCandidateRoute)
	} else {
		defaultCandidateRoute = BrowserRuntimeInfo{}
	}
	diagnosticsPreview := browserRuntimeDiagnosticsPreviewBaseForRegistrationPreview(preview)
	hints := BrowserToolMetadataRouteHints{
		Source:                   strings.TrimSpace(summary.SubstrateSource),
		Endpoint:                 strings.TrimSpace(summary.SubstrateEndpoint),
		DefaultCandidateRoute:    defaultCandidateRoute,
		DefaultCandidateSource:   strings.TrimSpace(summary.DefaultCandidateSource),
		DefaultCandidateEndpoint: strings.TrimSpace(summary.DefaultCandidateEndpoint),
		HiddenDefaultCandidate:   hiddenDefaultCandidate,
		SubstratePosture:         strings.TrimSpace(summary.SubstratePosture),
		SubstrateStatus:          strings.TrimSpace(summary.SubstrateStatus),
		SelectionStrategy:        strings.TrimSpace(summary.SelectionStrategy),
		SubstrateReason:          strings.TrimSpace(summary.SubstrateReason),
		RepairAvailable:          strings.TrimSpace(summary.RepairCommand) != "",
	}
	if hints.Source == "" {
		if visible := browserVisibleDefaultRuntimeInfoForPreview(preview); visible != (BrowserRuntimeInfo{}) {
			projection := browserRuntimeDoctorVisibleDefaultRouteProjection(diagnosticsPreview, visible)
			hints.Source = strings.TrimSpace(projection.Metadata.Source)
			hints.Endpoint = strings.TrimSpace(projection.Metadata.Endpoint)
		}
	}
	if hints.DefaultCandidateRoute != (BrowserRuntimeInfo{}) &&
		(hints.DefaultCandidateSource == "" || hints.DefaultCandidateEndpoint == "") {
		if candidateProjection, ok := browserRuntimeDoctorManagedCandidateRouteProjection(diagnosticsPreview); ok {
			if hints.DefaultCandidateSource == "" {
				hints.DefaultCandidateSource = strings.TrimSpace(candidateProjection.Metadata.Source)
			}
			if hints.DefaultCandidateEndpoint == "" {
				hints.DefaultCandidateEndpoint = strings.TrimSpace(candidateProjection.Metadata.Endpoint)
			}
		}
	}
	return hints
}
