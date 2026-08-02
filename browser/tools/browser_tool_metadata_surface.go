package tools

import "strings"

type browserManagedCompatMetadataSurface struct {
	Targets []string
	Kinds   []string
}

func browserManagedOptInMetadataSurfacesForOptions(opts BrowserToolOptions, preview browserDefaultRuntimePreview) map[string]browserManagedCompatMetadataSurface {
	enabled := buildEnabledToolSet(opts.EnabledTools)
	if len(enabled) == 0 {
		return nil
	}
	visible := preview.RegistrationCapabilities
	return browserManagedOptInProjectionForCapabilities(enabled, visible, opts.NodeBackend, opts.SandboxBackend).MetadataSurfaces()
}

func browserToolMetadataApplyManagedCompatSurface(meta ToolMetadata, surface browserManagedCompatMetadataSurface) ToolMetadata {
	return ApplyBrowserToolMetadataRouteHints(meta, BrowserToolMetadataRouteHints{
		Surface:      "explicit_managed_opt_in",
		OptInTargets: surface.Targets,
	})
}

func browserToolMetadataApplyManagedActSurface(meta ToolMetadata, surface browserManagedCompatMetadataSurface) ToolMetadata {
	if len(surface.Kinds) != 0 {
		meta.Capabilities = browserActMetadataCapabilities(surface.Kinds)
	}
	return browserToolMetadataApplyManagedCompatSurface(meta, surface)
}

func browserToolMetadataApplyManagedUnifiedSurface(meta ToolMetadata, surface browserManagedCompatMetadataSurface) ToolMetadata {
	if len(surface.Kinds) != 0 {
		meta.Capabilities = browserUnifiedMetadataCapabilitiesForManagedSurface(meta, surface.Kinds)
	}
	return browserToolMetadataApplyManagedCompatSurface(meta, surface)
}

func browserUnifiedMetadataCapabilitiesForManagedSurface(meta ToolMetadata, actKinds []string) []string {
	capabilities := append(
		[]string{"browser", "browser_workbench", "browser_runtime", "browser_act"},
		browserActMetadataCapabilities(actKinds)...,
	)
	for _, action := range browserUnifiedRuntimeActionsFromMetadata(meta) {
		switch action {
		case "inspect":
			capabilities = append(capabilities, "read")
		case "status", "profiles", "sessions":
			capabilities = append(capabilities, "read")
		case "workbench":
			capabilities = append(capabilities, "read", "browser_workbench", "browser_session_inspect")
		case "ready":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_prepare")
		case "inventory":
			capabilities = append(capabilities, "read")
		case "handles":
			capabilities = append(capabilities, "read", "browser_session_inspect")
		case "prepare":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_prepare")
		case "repair":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_bootstrap_repair")
		case "coordinate", "ensure", "refresh", "sync", "teardown":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_session_coordination")
		case "reset":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_clear")
		case "adopt":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_sync")
		case "start", "restart", "stop":
			capabilities = append(capabilities, "write", "browser_profile_control")
		case "launch", "halt":
			capabilities = append(capabilities, "write", "browser_profile_control")
		case "new_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_create")
		case "remove_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_delete")
		case "create_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_create")
		case "delete_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_delete")
		case "select_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_select")
		case "pin_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_select")
		case "clear_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_clear")
		case "unpin_profile":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_profile_clear")
		case "clear_session":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_clear")
		case "sync_session":
			capabilities = append(capabilities, "write", "browser_profile_control", "browser_target_control", "browser_session_control", "browser_session_sync")
		case "select_target":
			capabilities = append(capabilities, "write", "browser_target_control", "browser_target_select")
		case "pin_target":
			capabilities = append(capabilities, "write", "browser_target_control", "browser_target_select")
		case "clear_target":
			capabilities = append(capabilities, "write", "browser_target_control", "browser_target_clear")
		case "unpin_target":
			capabilities = append(capabilities, "write", "browser_target_control", "browser_target_clear")
		}
		capabilities = append(capabilities, "browser_kind:runtime_"+strings.ToLower(strings.TrimSpace(action)))
	}
	return mergeToolMetadataStrings(nil, capabilities)
}

func browserUnifiedRuntimeActionsFromMetadata(meta ToolMetadata) []string {
	actions := make([]string, 0, len(meta.Capabilities))
	for _, capability := range meta.Capabilities {
		value := strings.ToLower(strings.TrimSpace(capability))
		if !strings.HasPrefix(value, "browser_kind:runtime_") {
			continue
		}
		if action := strings.TrimPrefix(value, "browser_kind:runtime_"); action != "" {
			actions = append(actions, action)
		}
	}
	return mergeToolMetadataStrings(nil, actions)
}
