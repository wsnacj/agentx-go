package browserruntime

import "strings"

// SharedSessionBrowserConfiguredProfilesProjectionRequest carries the small set
// of profile hints needed to derive configured profile candidates for top-level
// payloads and binding/session projections.
type SharedSessionBrowserConfiguredProfilesProjectionRequest struct {
	RequestedProfile        string
	SessionProfile          string
	SessionTargetProfile    string
	DefaultProfile          string
	DefaultRouteProfile     string
	SelectedProfile         string
	InventoryProfiles       []string
	DiscoveredProfiles      []string
	ExistingConfigured      []string
	LegacySystemHost        bool
	LegacyFallbackProfile   string
	AppendManagedGenericSet bool
}

// ProjectSharedSessionBrowserConfiguredProfiles returns the trimmed, deduped
// configured profile list for a payload/binding view while preserving the
// legacy-host and managed-route fallback contracts.
func ProjectSharedSessionBrowserConfiguredProfiles(
	req SharedSessionBrowserConfiguredProfilesProjectionRequest,
) []string {
	profiles := make([]string, 0, 6+len(req.InventoryProfiles)+len(req.DiscoveredProfiles)+len(req.ExistingConfigured))
	profiles = append(
		profiles,
		strings.TrimSpace(req.RequestedProfile),
		strings.TrimSpace(req.SessionProfile),
		strings.TrimSpace(req.SessionTargetProfile),
		strings.TrimSpace(req.DefaultProfile),
		strings.TrimSpace(req.DefaultRouteProfile),
		strings.TrimSpace(req.SelectedProfile),
	)
	for _, profile := range req.InventoryProfiles {
		profiles = append(profiles, strings.TrimSpace(profile))
	}
	for _, profile := range req.DiscoveredProfiles {
		profiles = append(profiles, strings.TrimSpace(profile))
	}
	for _, profile := range req.ExistingConfigured {
		profiles = append(profiles, strings.TrimSpace(profile))
	}
	out := sharedSessionBrowserTrimmedUniqueStrings(profiles)
	if len(out) == 0 {
		if req.LegacySystemHost {
			return []string{firstNonEmptyString(strings.TrimSpace(req.LegacyFallbackProfile), "default")}
		}
		return []string{"default", "isolated", "relay"}
	}
	if req.AppendManagedGenericSet {
		return sharedSessionBrowserTrimmedUniqueStrings(append(out, "default", "isolated", "relay"))
	}
	return out
}

func sharedSessionBrowserTrimmedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
