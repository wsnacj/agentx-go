package tools

import agentxbrowserruntime "github.com/wsnacj/agentx-go/browser/runtime"

type browserRuntimeConfiguredProfilesProjection struct {
	Profiles             []string
	PreserveExisting     bool
	RequireSelectedRoute bool
}

func browserRuntimeSharedConfiguredProfilesProjectionRequest(
	payload browserRuntimePayload,
	selectedInfo BrowserRuntimeInfo,
	includeTargetSelection bool,
) agentxbrowserruntime.SharedSessionBrowserConfiguredProfilesProjectionRequest {
	selectedInfo = normalizeBrowserRuntimeInfo(selectedInfo)
	req := agentxbrowserruntime.SharedSessionBrowserConfiguredProfilesProjectionRequest{
		RequestedProfile:      payload.RequestedProfile,
		DefaultProfile:        payload.DefaultProfile,
		DefaultRouteProfile:   payload.DefaultRoute.Profile,
		SelectedProfile:       selectedInfo.Profile,
		LegacySystemHost:      BrowserSubstratePosture(selectedInfo.Backend, selectedInfo.Target) == BrowserSubstrateLegacySystemHost,
		LegacyFallbackProfile: firstNonEmpty(selectedInfo.Profile, defaultBrowserRuntimeInfo().Profile),
		AppendManagedGenericSet: BrowserSubstratePosture(selectedInfo.Backend, selectedInfo.Target) != BrowserSubstrateLegacySystemHost &&
			len(payload.Profiles) == 0 &&
			len(payload.discoveredProfiles) == 0 &&
			len(payload.ConfiguredProfiles) == 0,
	}
	if selection := payload.SessionProfileSelection; selection != nil {
		req.SessionProfile = selection.Profile
	}
	if includeTargetSelection {
		if selection := payload.SessionTargetSelection; selection != nil {
			req.SessionTargetProfile = selection.Profile
		}
	}
	if len(payload.Profiles) > 0 {
		req.InventoryProfiles = make([]string, 0, len(payload.Profiles))
		for _, item := range payload.Profiles {
			req.InventoryProfiles = append(req.InventoryProfiles, item.Profile)
		}
	}
	if len(payload.discoveredProfiles) > 0 {
		req.DiscoveredProfiles = append([]string(nil), payload.discoveredProfiles...)
	}
	if len(payload.ConfiguredProfiles) > 0 {
		req.ExistingConfigured = append([]string(nil), payload.ConfiguredProfiles...)
	}
	return req
}

func browserRuntimeConfiguredProfilesProjectionForPayload(
	payload browserRuntimePayload,
	selectedInfo BrowserRuntimeInfo,
) browserRuntimeConfiguredProfilesProjection {
	return browserRuntimeConfiguredProfilesProjection{
		Profiles: agentxbrowserruntime.ProjectSharedSessionBrowserConfiguredProfiles(
			browserRuntimeSharedConfiguredProfilesProjectionRequest(payload, selectedInfo, false),
		),
	}
}

func browserRuntimeApplyConfiguredProfilesProjection(
	payload *browserRuntimePayload,
	projection browserRuntimeConfiguredProfilesProjection,
) {
	if payload == nil {
		return
	}
	if projection.RequireSelectedRoute && payload.SelectedRoute == nil {
		return
	}
	if projection.PreserveExisting && len(payload.ConfiguredProfiles) > 0 {
		return
	}
	payload.ConfiguredProfiles = append([]string(nil), projection.Profiles...)
}
