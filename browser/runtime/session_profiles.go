package browserruntime

import (
	"context"
	"strings"
	"time"
)

// SharedSessionBrowserProfilesObservation captures a raw RuntimeProfiles
// observation together with the scoped lifecycle snapshot and projected profile
// states derived from it.
type SharedSessionBrowserProfilesObservation struct {
	Profiles    *BrowserProfilesResult
	ProfilesErr error
	ObservedAt  time.Time
	Snapshot    []SharedSessionBrowserProfileState
	Projected   []SharedSessionBrowserProjectedProfileState
}

// ObserveSharedSessionBrowserProfiles loads a RuntimeProfiles observation and,
// when a session registry is available, resolves it through the shared scoped
// lifecycle snapshot before projecting selected-profile metadata.
func ObserveSharedSessionBrowserProfiles(
	ctx context.Context,
	control BrowserRuntimeControlBackend,
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	requestedProfile string,
	reconnectWindow time.Duration,
) SharedSessionBrowserProfilesObservation {
	return sharedSessionBrowserObserverManager(
		nil,
		nil,
		registry,
		reconnectWindow,
	).ObserveProfiles(ctx, control, sessionID, selectedInfo, requestedProfile)
}

func sharedSessionBrowserProfilesObservationFromCycle(
	registry SharedSessionBrowserStateRegistry,
	sessionID string,
	selectedInfo BrowserRuntimeInfo,
	observation SharedSessionBrowserStatusAndProfilesObservation,
) SharedSessionBrowserProfilesObservation {
	profiles := SharedSessionBrowserProfilesObservation{
		Profiles:    observation.Profiles,
		ProfilesErr: observation.ProfilesErr,
		ObservedAt:  observation.ProfilesObservedAt,
		Snapshot:    observation.Snapshot,
	}
	if observation.Profiles == nil {
		return profiles
	}
	selection := sharedSessionBrowserSelectedProfileForTarget(registry, sessionID, selectedInfo.Target)
	if len(observation.Snapshot) > 0 {
		profiles.Projected = ProjectSharedSessionBrowserProfileSnapshot(observation.Snapshot, selection)
		return profiles
	}
	profiles.Projected = ProjectSharedSessionBrowserObservedProfiles(
		observation.Profiles.Backend,
		selectedInfo.Target,
		observation.Profiles.Profiles,
		selection,
	)
	return profiles
}

// ValidateSharedSessionBrowserSelectedProfile reports whether a RuntimeProfiles
// observation contained enough inventory to validate a selected profile and, if
// so, whether the requested profile exists together with any browser-app hint
// discovered for it.
func ValidateSharedSessionBrowserSelectedProfile(profile string, browserApp string, result BrowserProfilesResult) (string, bool, bool) {
	profile = strings.TrimSpace(profile)
	browserApp = strings.TrimSpace(browserApp)
	if profile == "" {
		return browserApp, false, false
	}
	if len(result.Profiles) == 0 && strings.TrimSpace(result.DefaultProfile) == "" && strings.TrimSpace(result.Note) == "" {
		return browserApp, false, false
	}
	for _, item := range result.Profiles {
		if !strings.EqualFold(strings.TrimSpace(item.Profile), profile) {
			continue
		}
		if browserApp == "" {
			browserApp = strings.TrimSpace(item.BrowserApp)
		}
		return browserApp, true, true
	}
	return browserApp, true, false
}

// SharedSessionBrowserProjectedProfileState is a shared scoped profile snapshot
// annotated with whether it matches the remembered session profile selection.
type SharedSessionBrowserProjectedProfileState struct {
	State    SharedSessionBrowserProfileState
	Selected bool
}

// SharedSessionBrowserProfileSelectionMatches reports whether a remembered
// session profile selection applies to the given backend/target/profile tuple.
func SharedSessionBrowserProfileSelectionMatches(selection *SharedSessionBrowserProfileSelection, backend string, runtimeTarget string, profile string) bool {
	if selection == nil {
		return false
	}
	if selected := strings.TrimSpace(selection.Profile); selected != "" && !strings.EqualFold(selected, strings.TrimSpace(profile)) {
		return false
	}
	if selected := strings.TrimSpace(selection.RuntimeTarget); selected != "" && !strings.EqualFold(selected, strings.TrimSpace(runtimeTarget)) {
		return false
	}
	if selected := browserSessionCanonicalBackend(strings.TrimSpace(selection.Backend)); selected != "" {
		current := browserSessionCanonicalBackend(strings.TrimSpace(backend))
		if current != "" && current != selected {
			return false
		}
	}
	return true
}

// ProjectSharedSessionBrowserProfileSnapshot annotates a scoped shared profile
// snapshot with selected-state metadata and fills missing browser-app identity
// from the remembered selection when it uniquely matches.
func ProjectSharedSessionBrowserProfileSnapshot(snapshot []SharedSessionBrowserProfileState, selection *SharedSessionBrowserProfileSelection) []SharedSessionBrowserProjectedProfileState {
	if len(snapshot) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserProjectedProfileState, 0, len(snapshot))
	for _, item := range snapshot {
		item.Backend = strings.TrimSpace(item.Backend)
		item.Profile = strings.TrimSpace(item.Profile)
		item.RuntimeTarget = strings.TrimSpace(item.RuntimeTarget)
		item.BrowserApp = strings.TrimSpace(item.BrowserApp)
		item.Status = strings.TrimSpace(item.Status)
		item.Note = strings.TrimSpace(item.Note)
		selected := SharedSessionBrowserProfileSelectionMatches(selection, item.Backend, item.RuntimeTarget, item.Profile)
		if selected && item.BrowserApp == "" && selection != nil {
			item.BrowserApp = strings.TrimSpace(selection.BrowserApp)
		}
		out = append(out, SharedSessionBrowserProjectedProfileState{
			State:    item,
			Selected: selected,
		})
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SnapshotSharedSessionBrowserProjectedProfilesForScope loads the scoped
// session lifecycle snapshot and projects selection metadata from the registry.
func SnapshotSharedSessionBrowserProjectedProfilesForScope(registry SharedSessionBrowserStateRegistry, sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string) []SharedSessionBrowserProjectedProfileState {
	if registry == nil {
		return nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	snapshot := registry.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
	if len(snapshot) == 0 {
		return nil
	}
	return ProjectSharedSessionBrowserProfileSnapshot(snapshot, sharedSessionBrowserSelectedProfileForTarget(registry, sessionID, selectedInfo.Target))
}

// ProjectSharedSessionBrowserObservedProfiles maps a raw RuntimeProfiles
// observation into the shared scoped profile contract and marks selected items
// using the remembered session profile selection.
func ProjectSharedSessionBrowserObservedProfiles(backend string, runtimeTarget string, items []BrowserProfileInfo, selection *SharedSessionBrowserProfileSelection) []SharedSessionBrowserProjectedProfileState {
	if len(items) == 0 {
		return nil
	}
	snapshot := make([]SharedSessionBrowserProfileState, 0, len(items))
	backend = strings.TrimSpace(backend)
	runtimeTarget = strings.TrimSpace(runtimeTarget)
	for _, item := range items {
		snapshot = append(snapshot, SharedSessionBrowserProfileState{
			Backend:       backend,
			Profile:       strings.TrimSpace(item.Profile),
			RuntimeTarget: runtimeTarget,
			BrowserApp:    strings.TrimSpace(item.BrowserApp),
			Status:        strings.TrimSpace(item.Status),
			Running:       item.Running,
			Connected:     item.Connected,
			Note:          strings.TrimSpace(item.Note),
		})
	}
	return ProjectSharedSessionBrowserProfileSnapshot(snapshot, selection)
}

// SharedSessionBrowserProfileStateSelected reports whether a runtime/browser
// profile state matches the remembered session selection, using the provided
// runtime-target fallback when the state itself omits it.
func SharedSessionBrowserProfileStateSelected(selection *SharedSessionBrowserProfileSelection, fallbackRuntimeTarget string, state SharedSessionBrowserProfileState) bool {
	return SharedSessionBrowserProfileSelectionMatches(
		selection,
		strings.TrimSpace(state.Backend),
		firstNonEmptyString(strings.TrimSpace(state.RuntimeTarget), strings.TrimSpace(fallbackRuntimeTarget)),
		strings.TrimSpace(state.Profile),
	)
}

// SharedSessionBrowserDiscoveredProfiles returns the trimmed unique profile
// names advertised by a RuntimeProfiles observation.
func SharedSessionBrowserDiscoveredProfiles(items []BrowserProfileInfo) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		profile := strings.TrimSpace(item.Profile)
		if profile == "" {
			continue
		}
		if _, ok := seen[profile]; ok {
			continue
		}
		seen[profile] = struct{}{}
		out = append(out, profile)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func sharedSessionBrowserSelectedProfileForTarget(registry SharedSessionBrowserStateRegistry, sessionID string, runtimeTarget string) *SharedSessionBrowserProfileSelection {
	if registry == nil {
		return nil
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}
	current, ok := registry.SelectedBrowserProfile(sessionID, strings.TrimSpace(runtimeTarget))
	if !ok {
		return nil
	}
	current.Backend = strings.TrimSpace(current.Backend)
	current.Profile = strings.TrimSpace(current.Profile)
	current.RuntimeTarget = strings.TrimSpace(current.RuntimeTarget)
	current.BrowserApp = strings.TrimSpace(current.BrowserApp)
	current.Source = strings.TrimSpace(current.Source)
	return &current
}
