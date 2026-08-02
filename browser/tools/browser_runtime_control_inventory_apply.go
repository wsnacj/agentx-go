package tools

type browserRuntimeTopLevelProfileInventoryProjection struct {
	ProfileStatus         *browserRuntimeProfileState
	ApplyProfileStatus    bool
	Profiles              []browserRuntimeProfileState
	DiscoveredProfiles    []string
	DefaultProfile        string
	ConfiguredInfo        BrowserRuntimeInfo
	ApplyProfileInventory bool
}

func browserRuntimeApplyTopLevelProfileInventory(
	payload *browserRuntimePayload,
	projection browserRuntimeTopLevelProfileInventoryProjection,
) {
	if payload == nil {
		return
	}
	if projection.ApplyProfileStatus {
		payload.ProfileStatus = projection.ProfileStatus
	}
	if !projection.ApplyProfileInventory {
		return
	}
	payload.Profiles = projection.Profiles
	payload.discoveredProfiles = projection.DiscoveredProfiles
	payload.DefaultProfile = projection.DefaultProfile
	browserRuntimeApplyConfiguredProfilesProjection(
		payload,
		browserRuntimeConfiguredProfilesProjectionForPayload(*payload, projection.ConfiguredInfo),
	)
}
