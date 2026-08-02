package browserruntime

import (
	"sort"
	"strings"
	"sync"
	"time"
)

type BrowserSessionStateRegistry struct {
	mu       sync.RWMutex
	sessions map[string]map[string]SharedSessionBrowserProfileState
	selected map[string]map[string]SharedSessionBrowserProfileSelection
}

func NewBrowserSessionStateRegistry() *BrowserSessionStateRegistry {
	return &BrowserSessionStateRegistry{
		sessions: map[string]map[string]SharedSessionBrowserProfileState{},
		selected: map[string]map[string]SharedSessionBrowserProfileSelection{},
	}
}

func (r *BrowserSessionStateRegistry) invalidateWatchManagers() {
	invalidateSharedSessionBrowserObserverManagersForStateRegistry(r)
}

func (r *BrowserSessionStateRegistry) RecordBrowserProfileState(sessionID string, state SharedSessionBrowserProfileState) {
	sessionID = strings.TrimSpace(sessionID)
	state = normalizeBrowserSessionProfileState(state)
	if r == nil || sessionID == "" {
		return
	}
	key := browserSessionStateKey(state)
	if key == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.sessions[sessionID] == nil {
		r.sessions[sessionID] = map[string]SharedSessionBrowserProfileState{}
	}
	r.sessions[sessionID][key] = browserPrepareSessionProfileState(r.sessions[sessionID][key], state)
}

func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfiles(sessionID string, filter SharedSessionBrowserProfileState, states []SharedSessionBrowserProfileState) {
	sessionID = strings.TrimSpace(sessionID)
	filter = normalizeBrowserSessionProfileState(filter)
	if r == nil || sessionID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	currentSession := r.sessions[sessionID]
	if len(currentSession) == 0 && len(states) == 0 {
		return
	}
	nextSession := map[string]SharedSessionBrowserProfileState{}
	for key, item := range currentSession {
		if browserSessionStateMatchesFilter(item, filter) {
			continue
		}
		nextSession[key] = item
	}
	for _, state := range states {
		state = normalizeBrowserSessionProfileState(state)
		key := browserSessionStateKey(state)
		if key == "" {
			continue
		}
		nextSession[key] = browserPrepareSessionProfileState(currentSession[key], state)
	}
	if len(nextSession) == 0 {
		delete(r.sessions, sessionID)
		return
	}
	r.sessions[sessionID] = nextSession
}

func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileState(sessionID string, state SharedSessionBrowserProfileState, reconnectWindow time.Duration) {
	_, _ = r.SyncSessionBrowserProfileObservation(sessionID, state, reconnectWindow)
}

func (r *BrowserSessionStateRegistry) ResolveSessionBrowserProfileStatus(sessionID string, selectedInfo BrowserRuntimeInfo, profile string, fallback BrowserProfileStatusResult) (BrowserProfileStatusResult, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return BrowserProfileStatusResult{}, false
	}
	desired := SharedSessionBrowserProfileStateFromStatus(selectedInfo, fallback)
	desired.Profile = firstNonEmptyString(
		strings.TrimSpace(desired.Profile),
		strings.TrimSpace(profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	if strings.TrimSpace(desired.Profile) == "" {
		return BrowserProfileStatusResult{}, false
	}
	resolved, ok := r.ResolveSessionBrowserProfileState(sessionID, desired)
	if !ok {
		return BrowserProfileStatusResult{}, false
	}
	return SharedSessionBrowserProfileStatusResultFromState(resolved, selectedInfo, desired.Profile), true
}

// SyncSessionBrowserProfileObservation applies managed transition preservation,
// replaces the route/profile-scoped snapshot, and returns the final synced state.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileObservation(sessionID string, state SharedSessionBrowserProfileState, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool) {
	sessionID = strings.TrimSpace(sessionID)
	state = normalizeBrowserSessionProfileState(state)
	if r == nil || sessionID == "" {
		return SharedSessionBrowserProfileState{}, false
	}
	if state.Profile == "" {
		r.RecordBrowserProfileState(sessionID, state)
		return state, true
	}
	if existing, ok := r.ResolveSessionBrowserProfileState(sessionID, state); ok {
		state = PrepareManagedSessionBrowserProfileTransition(existing, state, reconnectWindow)
	}
	r.SyncSessionBrowserProfiles(sessionID, SharedSessionBrowserProfileState{
		Backend:       strings.TrimSpace(state.Backend),
		RuntimeTarget: strings.TrimSpace(state.RuntimeTarget),
		Profile:       strings.TrimSpace(state.Profile),
	}, []SharedSessionBrowserProfileState{state})
	if synced, ok := r.ResolveSessionBrowserProfileState(sessionID, state); ok {
		return synced, true
	}
	return state, true
}

// SyncSessionBrowserProfileStatusObservation maps a runtime status observation
// onto the shared lifecycle contract, applies managed transition preservation,
// and returns the final synced state.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileStatusObservation(sessionID string, selectedInfo BrowserRuntimeInfo, result BrowserProfileStatusResult, observedAt time.Time, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool) {
	return r.SyncSessionBrowserProfileObservation(sessionID, SharedSessionBrowserProfileStateFromObservedStatus(selectedInfo, result, observedAt), reconnectWindow)
}

// SyncSessionBrowserProfileStatusResolution applies a single runtime status
// observation and returns the lifecycle-owned resolved status together with the
// final synced state.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileStatusResolution(sessionID string, selectedInfo BrowserRuntimeInfo, result BrowserProfileStatusResult, observedAt time.Time, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool) {
	synced, ok := r.SyncSessionBrowserProfileStatusObservation(sessionID, selectedInfo, result, observedAt, reconnectWindow)
	profile := firstNonEmptyString(strings.TrimSpace(result.Profile), strings.TrimSpace(selectedInfo.Profile))
	if ok {
		return SharedSessionBrowserProfileStatusResultFromState(synced, selectedInfo, profile), synced, true
	}
	if !sharedSessionBrowserProfileStatusResultEmpty(result) {
		return result, SharedSessionBrowserProfileState{}, false
	}
	return BrowserProfileStatusResult{}, SharedSessionBrowserProfileState{}, false
}

// SyncSessionBrowserProfileLifecycleObservation maps a lifecycle decision onto
// the shared lifecycle contract, applies managed transition preservation, and
// returns the final synced state.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileLifecycleObservation(sessionID string, selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string, observedAt time.Time, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool) {
	state, ok := SharedSessionBrowserProfileStateFromObservedLifecycle(selectedInfo, profile, result, decision, observedAt)
	if !ok {
		return SharedSessionBrowserProfileState{}, false
	}
	return r.SyncSessionBrowserProfileObservation(sessionID, state, reconnectWindow)
}

// SyncSessionBrowserProfileLifecycleResolution applies a single lifecycle
// decision and returns the lifecycle-owned resolved status together with the
// final synced state.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileLifecycleResolution(sessionID string, selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string, observedAt time.Time, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool) {
	profile = firstNonEmptyString(
		strings.TrimSpace(profile),
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	synced, ok := r.SyncSessionBrowserProfileLifecycleObservation(sessionID, selectedInfo, profile, result, decision, observedAt, reconnectWindow)
	if ok {
		return SharedSessionBrowserProfileStatusResultFromState(synced, selectedInfo, profile), synced, true
	}
	if !sharedSessionBrowserProfileStatusResultEmpty(result) {
		return result, SharedSessionBrowserProfileState{}, false
	}
	return BrowserProfileStatusResult{}, SharedSessionBrowserProfileState{}, false
}

// SyncSessionBrowserProfilesObservation maps runtime profile observations onto
// the shared lifecycle contract, applies managed transition preservation across
// the scoped profile set, and returns the final synced scoped snapshot.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfilesObservation(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result BrowserProfilesResult, observedAt time.Time, reconnectWindow time.Duration) []SharedSessionBrowserProfileState {
	filter := SharedSessionBrowserProfileState{
		Backend:       strings.TrimSpace(result.Backend),
		RuntimeTarget: strings.TrimSpace(selectedInfo.Target),
		Profile:       strings.TrimSpace(requestedProfile),
	}
	if filter.Backend == "" {
		filter.Backend = strings.TrimSpace(selectedInfo.Backend)
	}
	r.SyncSessionBrowserProfileObservations(sessionID, filter, SharedSessionBrowserProfileStatesFromObservedProfiles(selectedInfo, result, observedAt), reconnectWindow)
	return r.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
}

// SyncSessionBrowserProfilesResolution applies runtime profile observations and
// returns the final synced scoped snapshot.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfilesResolution(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, result BrowserProfilesResult, observedAt time.Time, reconnectWindow time.Duration) []SharedSessionBrowserProfileState {
	return r.SyncSessionBrowserProfilesObservation(sessionID, selectedInfo, requestedProfile, result, observedAt, reconnectWindow)
}

// SyncSessionBrowserStatusAndProfilesObservations applies route-scoped
// profile inventory sync plus current-profile status sync, then returns the
// final synced state and scoped snapshot.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserStatusAndProfilesObservations(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, status *BrowserProfileStatusResult, statusObservedAt time.Time, profiles *BrowserProfilesResult, profilesObservedAt time.Time, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return SharedSessionBrowserProfileState{}, false, nil
	}
	if profiles != nil {
		r.SyncSessionBrowserProfilesObservation(sessionID, selectedInfo, requestedProfile, *profiles, profilesObservedAt, reconnectWindow)
	}
	if status == nil {
		return SharedSessionBrowserProfileState{}, false, r.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
	}
	synced, ok := r.SyncSessionBrowserProfileStatusObservation(sessionID, selectedInfo, *status, statusObservedAt, reconnectWindow)
	return synced, ok, r.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
}

// SyncSessionBrowserStatusAndProfilesResolution applies route-scoped status and
// profile observations, then returns the resolved status result together with
// the final synced state and scoped snapshot.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserStatusAndProfilesResolution(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, status *BrowserProfileStatusResult, statusObservedAt time.Time, profiles *BrowserProfilesResult, profilesObservedAt time.Time, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	synced, ok, snapshot := r.SyncSessionBrowserStatusAndProfilesObservations(sessionID, selectedInfo, requestedProfile, status, statusObservedAt, profiles, profilesObservedAt, reconnectWindow)
	profile := firstNonEmptyString(strings.TrimSpace(requestedProfile), strings.TrimSpace(selectedInfo.Profile))
	if status != nil {
		profile = firstNonEmptyString(strings.TrimSpace(status.Profile), profile)
	}
	if ok {
		return SharedSessionBrowserProfileStatusResultFromState(synced, selectedInfo, profile), synced, true, snapshot
	}
	if status != nil && !sharedSessionBrowserProfileStatusResultEmpty(*status) {
		return *status, SharedSessionBrowserProfileState{}, false, snapshot
	}
	return BrowserProfileStatusResult{}, SharedSessionBrowserProfileState{}, false, snapshot
}

// SyncSessionBrowserExecutionObservations applies scoped profile inventory
// sync plus either lifecycle or status sync for the active profile, then
// returns the final synced state and scoped snapshot.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserExecutionObservations(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, profile string, profiles *BrowserProfilesResult, profilesObservedAt time.Time, result BrowserProfileStatusResult, resultObservedAt time.Time, decision string, reconnectWindow time.Duration) (SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return SharedSessionBrowserProfileState{}, false, nil
	}
	if profiles != nil {
		r.SyncSessionBrowserProfilesObservation(sessionID, selectedInfo, requestedProfile, *profiles, profilesObservedAt, reconnectWindow)
	}

	profile = firstNonEmptyString(
		strings.TrimSpace(profile),
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
	)

	var synced SharedSessionBrowserProfileState
	var syncedOK bool
	if strings.TrimSpace(decision) != "" {
		synced, syncedOK = r.SyncSessionBrowserProfileLifecycleObservation(sessionID, selectedInfo, profile, result, decision, resultObservedAt, reconnectWindow)
	}
	if !syncedOK && !sharedSessionBrowserProfileStatusResultEmpty(result) {
		synced, syncedOK = r.SyncSessionBrowserProfileStatusObservation(sessionID, selectedInfo, result, resultObservedAt, reconnectWindow)
	}
	return synced, syncedOK, r.SnapshotSessionBrowserProfilesForScope(sessionID, selectedInfo, requestedProfile)
}

// SyncSessionBrowserExecutionResolution applies scoped execution observations,
// then returns the resolved status result together with the final synced state
// and scoped snapshot.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserExecutionResolution(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string, profile string, profiles *BrowserProfilesResult, profilesObservedAt time.Time, result BrowserProfileStatusResult, resultObservedAt time.Time, decision string, reconnectWindow time.Duration) (BrowserProfileStatusResult, SharedSessionBrowserProfileState, bool, []SharedSessionBrowserProfileState) {
	synced, ok, snapshot := r.SyncSessionBrowserExecutionObservations(sessionID, selectedInfo, requestedProfile, profile, profiles, profilesObservedAt, result, resultObservedAt, decision, reconnectWindow)
	profile = firstNonEmptyString(
		strings.TrimSpace(profile),
		strings.TrimSpace(result.Profile),
		strings.TrimSpace(requestedProfile),
		strings.TrimSpace(selectedInfo.Profile),
	)
	if ok {
		return SharedSessionBrowserProfileStatusResultFromState(synced, selectedInfo, profile), synced, true, snapshot
	}
	if !sharedSessionBrowserProfileStatusResultEmpty(result) {
		return result, SharedSessionBrowserProfileState{}, false, snapshot
	}
	return BrowserProfileStatusResult{}, SharedSessionBrowserProfileState{}, false, snapshot
}

// SnapshotSessionBrowserProfilesForScope returns the session-scoped lifecycle
// snapshot filtered to the selected backend/runtime target and optional profile.
func (r *BrowserSessionStateRegistry) SnapshotSessionBrowserProfilesForScope(sessionID string, selectedInfo BrowserRuntimeInfo, requestedProfile string) []SharedSessionBrowserProfileState {
	snapshot := r.SnapshotSessionBrowserProfiles(sessionID)
	if len(snapshot) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserProfileState, 0, len(snapshot))
	for _, item := range snapshot {
		if !sharedSessionBrowserProfileMatchesSelectedInfo(item, requestedProfile, selectedInfo) {
			continue
		}
		out = append(out, item)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// SyncSessionBrowserProfileObservations applies managed transition preservation
// across a route/profile-scoped observation set before replacing that scope.
func (r *BrowserSessionStateRegistry) SyncSessionBrowserProfileObservations(sessionID string, filter SharedSessionBrowserProfileState, states []SharedSessionBrowserProfileState, reconnectWindow time.Duration) {
	sessionID = strings.TrimSpace(sessionID)
	filter = normalizeBrowserSessionProfileState(filter)
	if r == nil || sessionID == "" {
		return
	}
	prepared := make([]SharedSessionBrowserProfileState, 0, len(states))
	for _, state := range states {
		state = normalizeBrowserSessionProfileState(state)
		if state.Profile != "" {
			if existing, ok := r.ResolveSessionBrowserProfileState(sessionID, state); ok {
				state = PrepareManagedSessionBrowserProfileTransition(existing, state, reconnectWindow)
			}
		}
		prepared = append(prepared, state)
	}
	r.SyncSessionBrowserProfiles(sessionID, filter, prepared)
}

// SharedSessionBrowserProfileStateFromStatus maps a runtime status observation
// onto the shared session profile state contract using the selected route as
// identity fallback when the backend omits backend/profile fields.
func SharedSessionBrowserProfileStateFromObservedStatus(selectedInfo BrowserRuntimeInfo, result BrowserProfileStatusResult, observedAt time.Time) SharedSessionBrowserProfileState {
	backend := strings.TrimSpace(result.Backend)
	if backend == "" {
		backend = strings.TrimSpace(selectedInfo.Backend)
	}
	profile := strings.TrimSpace(result.Profile)
	if profile == "" {
		profile = strings.TrimSpace(selectedInfo.Profile)
	}
	return SharedSessionBrowserProfileState{
		Backend:       backend,
		Profile:       profile,
		RuntimeTarget: strings.TrimSpace(selectedInfo.Target),
		BrowserApp:    strings.TrimSpace(result.BrowserApp),
		Status:        strings.TrimSpace(result.Status),
		Running:       result.Running,
		Connected:     result.Connected,
		Note:          strings.TrimSpace(result.Note),
		ObservedAt:    observedAt,
	}
}

func SharedSessionBrowserProfileStateFromStatus(selectedInfo BrowserRuntimeInfo, result BrowserProfileStatusResult) SharedSessionBrowserProfileState {
	return SharedSessionBrowserProfileStateFromObservedStatus(selectedInfo, result, time.Time{})
}

func SharedSessionBrowserProfileStatesFromObservedProfiles(selectedInfo BrowserRuntimeInfo, result BrowserProfilesResult, observedAt time.Time) []SharedSessionBrowserProfileState {
	backend := strings.TrimSpace(result.Backend)
	if backend == "" {
		backend = strings.TrimSpace(selectedInfo.Backend)
	}
	states := make([]SharedSessionBrowserProfileState, 0, len(result.Profiles))
	for _, item := range result.Profiles {
		states = append(states, SharedSessionBrowserProfileState{
			Backend:       backend,
			Profile:       strings.TrimSpace(item.Profile),
			RuntimeTarget: strings.TrimSpace(selectedInfo.Target),
			BrowserApp:    strings.TrimSpace(item.BrowserApp),
			Status:        strings.TrimSpace(item.Status),
			Running:       item.Running,
			Connected:     item.Connected,
			Note:          strings.TrimSpace(item.Note),
			ObservedAt:    observedAt,
		})
	}
	return states
}

func SharedSessionBrowserProfileStatesFromProfiles(selectedInfo BrowserRuntimeInfo, result BrowserProfilesResult) []SharedSessionBrowserProfileState {
	return SharedSessionBrowserProfileStatesFromObservedProfiles(selectedInfo, result, time.Time{})
}

func SharedSessionBrowserProfileStateFromObservedLifecycle(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string, observedAt time.Time) (SharedSessionBrowserProfileState, bool) {
	state := SharedSessionBrowserProfileStateFromObservedStatus(selectedInfo, result, observedAt)
	state.Profile = firstNonEmptyString(strings.TrimSpace(state.Profile), strings.TrimSpace(profile))
	switch strings.TrimSpace(decision) {
	case "restart_started", "restarted":
		if !state.Connected {
			state.Status = "reconnecting"
			if state.Note == "" {
				state.Note = "restart requested"
			}
			return state, true
		}
	case "restart_reconnect_in_progress", "already_ready":
		if strings.TrimSpace(state.Status) != "" || state.Running || state.Connected {
			return state, true
		}
	case "started":
		if !state.Running || !state.Connected {
			state.Status = "starting"
			if state.Note == "" {
				state.Note = "start requested"
			}
			return state, true
		}
	case "stop_already_stopped", "stopped", "teardown_stopped", "teardown_already_stopped":
		state.Status = "stopped"
		state.Running = false
		state.Connected = false
		return state, true
	case "deleted", "delete_requested":
		state.Status = strings.TrimSpace(decision)
		state.Running = false
		state.Connected = false
		return state, true
	}
	return SharedSessionBrowserProfileState{}, false
}

func SharedSessionBrowserProfileStateFromLifecycle(selectedInfo BrowserRuntimeInfo, profile string, result BrowserProfileStatusResult, decision string) (SharedSessionBrowserProfileState, bool) {
	return SharedSessionBrowserProfileStateFromObservedLifecycle(selectedInfo, profile, result, decision, time.Time{})
}

func sharedSessionBrowserProfileStatusResultEmpty(result BrowserProfileStatusResult) bool {
	return strings.TrimSpace(result.Backend) == "" &&
		strings.TrimSpace(result.BrowserApp) == "" &&
		strings.TrimSpace(result.Profile) == "" &&
		strings.TrimSpace(result.Status) == "" &&
		!result.Running &&
		!result.Connected &&
		strings.TrimSpace(result.Note) == ""
}

func (r *BrowserSessionStateRegistry) SnapshotSessionBrowserProfiles(sessionID string) []SharedSessionBrowserProfileState {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := r.sessions[sessionID]
	if len(state) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserProfileState, 0, len(state))
	for _, item := range state {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Backend != out[j].Backend {
			return out[i].Backend < out[j].Backend
		}
		if out[i].RuntimeTarget != out[j].RuntimeTarget {
			return out[i].RuntimeTarget < out[j].RuntimeTarget
		}
		if out[i].Profile != out[j].Profile {
			return out[i].Profile < out[j].Profile
		}
		return out[i].BrowserApp < out[j].BrowserApp
	})
	return out
}

func (r *BrowserSessionStateRegistry) ResolveSessionBrowserProfileState(sessionID string, desired SharedSessionBrowserProfileState) (SharedSessionBrowserProfileState, bool) {
	sessionID = strings.TrimSpace(sessionID)
	desired = normalizeBrowserSessionProfileState(desired)
	if r == nil || sessionID == "" || strings.TrimSpace(desired.Profile) == "" {
		return SharedSessionBrowserProfileState{}, false
	}
	r.mu.RLock()
	session := r.sessions[sessionID]
	if len(session) == 0 {
		r.mu.RUnlock()
		return SharedSessionBrowserProfileState{}, false
	}
	items := make([]SharedSessionBrowserProfileState, 0, len(session))
	for _, item := range session {
		items = append(items, item)
	}
	r.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool {
		if items[i].Backend != items[j].Backend {
			return items[i].Backend < items[j].Backend
		}
		if items[i].RuntimeTarget != items[j].RuntimeTarget {
			return items[i].RuntimeTarget < items[j].RuntimeTarget
		}
		if items[i].Profile != items[j].Profile {
			return items[i].Profile < items[j].Profile
		}
		return items[i].BrowserApp < items[j].BrowserApp
	})
	var fallback SharedSessionBrowserProfileState
	var fallbackSet bool
	for _, item := range items {
		if !browserSessionStateMatchesFilter(item, SharedSessionBrowserProfileState{
			Backend:       desired.Backend,
			RuntimeTarget: desired.RuntimeTarget,
			Profile:       desired.Profile,
		}) {
			continue
		}
		if desired.BrowserApp != "" && strings.EqualFold(strings.TrimSpace(item.BrowserApp), strings.TrimSpace(desired.BrowserApp)) {
			return item, true
		}
		if strings.TrimSpace(desired.BrowserApp) == "" {
			if browserSessionStateHealthKey(item) == browserSessionStateHealthKey(desired) {
				return item, true
			}
			if !fallbackSet && strings.TrimSpace(item.BrowserApp) != "" {
				fallback = item
				fallbackSet = true
			}
		}
	}
	if fallbackSet {
		return fallback, true
	}
	return SharedSessionBrowserProfileState{}, false
}

func (r *BrowserSessionStateRegistry) SelectBrowserProfile(sessionID string, selection SharedSessionBrowserProfileSelection) {
	sessionID = strings.TrimSpace(sessionID)
	selection.Backend = strings.TrimSpace(selection.Backend)
	selection.Profile = strings.TrimSpace(selection.Profile)
	selection.RuntimeTarget = strings.TrimSpace(selection.RuntimeTarget)
	selection.BrowserApp = strings.TrimSpace(selection.BrowserApp)
	selection.Source = strings.TrimSpace(selection.Source)
	if r == nil || sessionID == "" || selection.Profile == "" {
		return
	}
	key := browserSessionSelectionKey(selection.RuntimeTarget)
	if key == "" {
		return
	}
	r.mu.Lock()
	if r.selected[sessionID] == nil {
		r.selected[sessionID] = map[string]SharedSessionBrowserProfileSelection{}
	}
	current := r.selected[sessionID][key]
	if selection.Backend == "" {
		selection.Backend = current.Backend
	}
	if selection.RuntimeTarget == "" {
		selection.RuntimeTarget = current.RuntimeTarget
	}
	if selection.BrowserApp == "" {
		selection.BrowserApp = current.BrowserApp
	}
	if selection.Source == "" {
		selection.Source = current.Source
	}
	if current == selection {
		r.mu.Unlock()
		return
	}
	r.selected[sessionID][key] = selection
	r.mu.Unlock()
	r.invalidateWatchManagers()
}

func (r *BrowserSessionStateRegistry) ClearSelectedBrowserProfile(sessionID string, runtimeTarget string) {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return
	}
	key := browserSessionSelectionKey(runtimeTarget)
	if key == "" {
		return
	}
	r.mu.Lock()
	selected := r.selected[sessionID]
	if len(selected) == 0 {
		r.mu.Unlock()
		return
	}
	if _, ok := selected[key]; !ok {
		r.mu.Unlock()
		return
	}
	delete(selected, key)
	if len(selected) == 0 {
		delete(r.selected, sessionID)
	}
	r.mu.Unlock()
	r.invalidateWatchManagers()
}

func (r *BrowserSessionStateRegistry) SelectedBrowserProfile(sessionID string, runtimeTarget string) (SharedSessionBrowserProfileSelection, bool) {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return SharedSessionBrowserProfileSelection{}, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	selected := r.selected[sessionID]
	if len(selected) == 0 {
		return SharedSessionBrowserProfileSelection{}, false
	}
	if strings.TrimSpace(runtimeTarget) == "" {
		if len(selected) == 1 {
			for _, selection := range selected {
				return selection, true
			}
		}
		return SharedSessionBrowserProfileSelection{}, false
	}
	key := browserSessionSelectionKey(runtimeTarget)
	selection, ok := selected[key]
	return selection, ok
}

func (r *BrowserSessionStateRegistry) SnapshotSelectedBrowserProfiles(sessionID string) []SharedSessionBrowserProfileSelection {
	sessionID = strings.TrimSpace(sessionID)
	if r == nil || sessionID == "" {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	selected := r.selected[sessionID]
	if len(selected) == 0 {
		return nil
	}
	out := make([]SharedSessionBrowserProfileSelection, 0, len(selected))
	for _, item := range selected {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].RuntimeTarget != out[j].RuntimeTarget {
			return out[i].RuntimeTarget < out[j].RuntimeTarget
		}
		if out[i].Backend != out[j].Backend {
			return out[i].Backend < out[j].Backend
		}
		if out[i].Profile != out[j].Profile {
			return out[i].Profile < out[j].Profile
		}
		return out[i].BrowserApp < out[j].BrowserApp
	})
	return out
}

func (r *BrowserSessionStateRegistry) ClearSessionBrowserProfiles(sessionID string, filter SharedSessionBrowserProfileState) int {
	sessionID = strings.TrimSpace(sessionID)
	filter.Backend = strings.TrimSpace(filter.Backend)
	filter.Profile = strings.TrimSpace(filter.Profile)
	filter.RuntimeTarget = strings.TrimSpace(filter.RuntimeTarget)
	filter.BrowserApp = strings.TrimSpace(filter.BrowserApp)
	if r == nil || sessionID == "" {
		return 0
	}
	r.mu.Lock()
	state := r.sessions[sessionID]
	if len(state) == 0 {
		r.mu.Unlock()
		return 0
	}
	cleared := 0
	for key, item := range state {
		if !browserSessionStateMatchesFilter(item, filter) {
			continue
		}
		delete(state, key)
		cleared++
	}
	if len(state) == 0 {
		delete(r.sessions, sessionID)
	}
	r.mu.Unlock()
	if cleared > 0 {
		r.invalidateWatchManagers()
	}
	return cleared
}

func browserSessionStateKey(state SharedSessionBrowserProfileState) string {
	parts := []string{
		strings.ToLower(strings.TrimSpace(state.Backend)),
		strings.ToLower(strings.TrimSpace(state.RuntimeTarget)),
		strings.ToLower(strings.TrimSpace(state.Profile)),
		strings.ToLower(strings.TrimSpace(state.BrowserApp)),
	}
	if strings.Join(parts, "") == "" {
		return ""
	}
	return strings.Join(parts, "|")
}

func browserSessionSelectionKey(runtimeTarget string) string {
	key := strings.ToLower(strings.TrimSpace(runtimeTarget))
	if key == "" {
		key = "host"
	}
	return key
}

func browserSessionStateMatchesFilter(state SharedSessionBrowserProfileState, filter SharedSessionBrowserProfileState) bool {
	if selected := strings.TrimSpace(filter.Backend); selected != "" && !strings.EqualFold(strings.TrimSpace(state.Backend), selected) {
		return false
	}
	if selected := strings.TrimSpace(filter.RuntimeTarget); selected != "" && !strings.EqualFold(strings.TrimSpace(state.RuntimeTarget), selected) {
		return false
	}
	if selected := strings.TrimSpace(filter.Profile); selected != "" && !strings.EqualFold(strings.TrimSpace(state.Profile), selected) {
		return false
	}
	if selected := strings.TrimSpace(filter.BrowserApp); selected != "" && !strings.EqualFold(strings.TrimSpace(state.BrowserApp), selected) {
		return false
	}
	return true
}

func browserSessionStateHealthKey(state SharedSessionBrowserProfileState) string {
	status := strings.ToLower(strings.TrimSpace(state.Status))
	switch {
	case status != "":
		return status
	case state.Running && state.Connected:
		return "running_connected"
	case state.Running:
		return "running"
	default:
		return "stopped"
	}
}

func sharedSessionBrowserProfileMatchesSelectedInfo(state SharedSessionBrowserProfileState, requestedProfile string, selectedInfo BrowserRuntimeInfo) bool {
	if requestedProfile = strings.TrimSpace(requestedProfile); requestedProfile != "" && !strings.EqualFold(strings.TrimSpace(state.Profile), requestedProfile) {
		return false
	}
	if target := strings.TrimSpace(selectedInfo.Target); target != "" && !strings.EqualFold(strings.TrimSpace(state.RuntimeTarget), target) {
		return false
	}
	if backend := strings.TrimSpace(selectedInfo.Backend); backend != "" && strings.TrimSpace(state.Backend) != "" && !sharedSessionBrowserRuntimeBackendMatches(state.Backend, backend) {
		return false
	}
	return true
}

func sharedSessionBrowserRuntimeBackendMatches(left string, right string) bool {
	left = strings.ToLower(strings.TrimSpace(left))
	right = strings.ToLower(strings.TrimSpace(right))
	if left == "" || right == "" {
		return left == right
	}
	return left == right
}

func normalizeBrowserSessionProfileState(state SharedSessionBrowserProfileState) SharedSessionBrowserProfileState {
	state.Backend = strings.TrimSpace(state.Backend)
	state.Profile = strings.TrimSpace(state.Profile)
	state.RuntimeTarget = strings.TrimSpace(state.RuntimeTarget)
	state.BrowserApp = strings.TrimSpace(state.BrowserApp)
	state.Status = strings.TrimSpace(state.Status)
	state.Note = strings.TrimSpace(state.Note)
	return state
}

func browserPrepareSessionProfileState(current SharedSessionBrowserProfileState, state SharedSessionBrowserProfileState) SharedSessionBrowserProfileState {
	state = normalizeBrowserSessionProfileState(state)
	current = normalizeBrowserSessionProfileState(current)
	if state.ObservedAt.IsZero() {
		state.ObservedAt = time.Now()
	}
	if state.Backend == "" {
		state.Backend = current.Backend
	}
	if state.Profile == "" {
		state.Profile = current.Profile
	}
	if state.RuntimeTarget == "" {
		state.RuntimeTarget = current.RuntimeTarget
	}
	if state.BrowserApp == "" {
		state.BrowserApp = current.BrowserApp
	}
	if state.Status == "" {
		state.Status = current.Status
	}
	if state.Note == "" {
		state.Note = current.Note
	}
	if state.Status == "" && !state.Running && !state.Connected {
		state.Running = current.Running
		state.Connected = current.Connected
	}
	if state.StatusSince.IsZero() {
		if !current.StatusSince.IsZero() && browserSessionStateHealthKey(current) == browserSessionStateHealthKey(state) {
			state.StatusSince = current.StatusSince
		} else {
			state.StatusSince = state.ObservedAt
		}
	}
	if state.ObservedAt.IsZero() {
		state.ObservedAt = state.StatusSince
	}
	return state
}

func PrepareManagedSessionBrowserProfileTransition(current SharedSessionBrowserProfileState, next SharedSessionBrowserProfileState, reconnectWindow time.Duration) SharedSessionBrowserProfileState {
	current = normalizeBrowserSessionProfileState(current)
	next = normalizeBrowserSessionProfileState(next)
	switch strings.ToLower(strings.TrimSpace(next.RuntimeTarget)) {
	case "node", "sandbox":
	default:
		return next
	}
	if current.Profile == "" {
		return next
	}
	if next.Backend == "" {
		next.Backend = current.Backend
	}
	if next.Profile == "" {
		next.Profile = current.Profile
	}
	if next.RuntimeTarget == "" {
		next.RuntimeTarget = current.RuntimeTarget
	}
	if next.BrowserApp == "" {
		next.BrowserApp = current.BrowserApp
	}
	if next.Connected {
		return next
	}
	switch strings.ToLower(strings.TrimSpace(next.Status)) {
	case "", "running", "started", "ready", "connected":
		switch strings.ToLower(strings.TrimSpace(current.Status)) {
		case "reconnecting", "starting", "disconnected", "crashed":
			if next.Running || current.Running {
				next.Status = strings.TrimSpace(current.Status)
				if next.Note == "" {
					next.Note = strings.TrimSpace(current.Note)
				}
			}
		}
	case "reconnecting":
		if strings.EqualFold(strings.TrimSpace(current.Status), "reconnecting") &&
			!current.StatusSince.IsZero() &&
			time.Since(current.StatusSince) >= reconnectWindow {
			now := next.ObservedAt
			if now.IsZero() {
				now = time.Now()
			}
			next.ObservedAt = now
			next.StatusSince = now
		}
	}
	return next
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
