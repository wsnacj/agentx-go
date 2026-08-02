package browserruntime

import "strings"

// SharedSessionBrowserRouteResolution describes which owner selected the final
// profile/runtime-target/current-target tuple for a browser runtime payload.
type SharedSessionBrowserRouteResolution struct {
	ProfileSource       string
	RuntimeTargetSource string
	TargetSource        string
}

// ResolveSharedSessionBrowserRouteResolution projects the profile/runtime-target
// selection sources for a resolved browser runtime route.
func ResolveSharedSessionBrowserRouteResolution(
	requestedProfile string,
	requestedRuntimeTarget string,
	defaultInfo BrowserRuntimeInfo,
	selectedInfo BrowserRuntimeInfo,
	profileSelection *SharedSessionBrowserProfileSelection,
	targetSelection *BrowserSessionTargetSelection,
) (SharedSessionBrowserRouteResolution, bool) {
	requestedProfile = strings.TrimSpace(requestedProfile)
	requestedRuntimeTarget = strings.TrimSpace(requestedRuntimeTarget)
	defaultInfo.Backend = strings.TrimSpace(defaultInfo.Backend)
	defaultInfo.Profile = strings.TrimSpace(defaultInfo.Profile)
	defaultInfo.Target = strings.TrimSpace(defaultInfo.Target)
	selectedInfo.Backend = strings.TrimSpace(selectedInfo.Backend)
	selectedInfo.Profile = strings.TrimSpace(selectedInfo.Profile)
	selectedInfo.Target = strings.TrimSpace(selectedInfo.Target)

	resolution := SharedSessionBrowserRouteResolution{}
	switch {
	case requestedProfile != "":
		resolution.ProfileSource = "explicit_request"
	case profileSelection != nil &&
		strings.TrimSpace(profileSelection.Profile) != "" &&
		strings.EqualFold(strings.TrimSpace(profileSelection.Profile), selectedInfo.Profile):
		resolution.ProfileSource = firstNonEmptyString(strings.TrimSpace(profileSelection.Source), "session_profile_selection")
	case defaultInfo.Profile != "" && strings.EqualFold(defaultInfo.Profile, selectedInfo.Profile):
		resolution.ProfileSource = "default_route"
	case selectedInfo.Profile != "":
		resolution.ProfileSource = "backend_resolved"
	}

	switch {
	case requestedRuntimeTarget != "":
		resolution.RuntimeTargetSource = "explicit_request"
	case profileSelection != nil &&
		strings.TrimSpace(profileSelection.RuntimeTarget) != "" &&
		strings.EqualFold(strings.TrimSpace(profileSelection.RuntimeTarget), selectedInfo.Target):
		resolution.RuntimeTargetSource = firstNonEmptyString(strings.TrimSpace(profileSelection.Source), "session_profile_selection")
	case defaultInfo.Target != "" && strings.EqualFold(defaultInfo.Target, selectedInfo.Target):
		resolution.RuntimeTargetSource = "default_route"
	case selectedInfo.Target != "":
		resolution.RuntimeTargetSource = "backend_resolved"
	}

	if targetSelection != nil {
		resolution.TargetSource = firstNonEmptyString(strings.TrimSpace(targetSelection.Source), "session_target_selection")
	}

	if resolution.ProfileSource == "" && resolution.RuntimeTargetSource == "" && resolution.TargetSource == "" {
		return SharedSessionBrowserRouteResolution{}, false
	}
	return resolution, true
}
